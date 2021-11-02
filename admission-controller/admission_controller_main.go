/*
Copyright © 2021 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	commLog "github.com/intel-secl/intel-secl/v4/pkg/lib/common/log"
	commLogMsg "github.com/intel-secl/intel-secl/v4/pkg/lib/common/log/message"
	commLogInt "github.com/intel-secl/intel-secl/v4/pkg/lib/common/log/setup"
	"github.com/intel-secl/k8s-extensions/v4/admission-controller/config"
	"github.com/intel-secl/k8s-extensions/v4/admission-controller/constants"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/api/admission/v1beta1"
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

var (
	universalDeserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
)

var configuration *rest.Config

const logFilePath = "/var/log/isecl-k8s-extensions/isecl-admission-controller.log"

var (
	defaultLog = commLog.GetDefaultLogger()
)

func configureLogs(logFile *os.File, loglevel string, maxLength int) error {

	lv, err := logrus.ParseLevel(loglevel)
	if err != nil {
		return errors.Wrap(err, "Failed to initiate loggers. Invalid log level: "+loglevel)
	}
	ioWriterDefault := io.MultiWriter(os.Stdout, logFile)
	f := commLog.LogFormatter{MaxLength: maxLength}
	commLogInt.SetLogger(commLog.DefaultLoggerName, lv, &f, ioWriterDefault, false)

	defaultLog.Info(commLogMsg.LogInit)
	return nil
}

func startServer(router *mux.Router, admissionControllerConfig config.Config) error {
	tlsconfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
	}

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGKILL)
	//get a webserver instance, that contains a muxer, middleware and configuration settings

	//initialize http server config
	httpWriter := os.Stderr
	if httpLogFile, err := os.OpenFile(constants.HttpLogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666); err != nil {
		defaultLog.Tracef("service:Start() %+v", err)
	} else {
		defer func() {
			derr := httpLogFile.Close()
			if derr != nil {
				defaultLog.WithError(derr).Error("Error closing file")
			}
		}()
		httpWriter = httpLogFile
	}

	httpLog := stdlog.New(httpWriter, "", 0)
	h := &http.Server{
		Addr:      fmt.Sprintf(":%d", admissionControllerConfig.Port),
		Handler:   handlers.RecoveryHandler(handlers.RecoveryLogger(httpLog), handlers.PrintRecoveryStack(true))(handlers.CombinedLoggingHandler(os.Stderr, router)),
		ErrorLog:  httpLog,
		TLSConfig: tlsconfig,
	}

	// dispatch web server go routine
	go func() {
		if err := h.ListenAndServeTLS(admissionControllerConfig.ServerCert, admissionControllerConfig.ServerKey); err != nil {
			defaultLog.Errorf("failed to start service %+v", err)
			stop <- syscall.SIGTERM
		}
	}()
	defaultLog.Info("Service started")

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to gracefully shutdown webserver: %v\n", err)
		return nil
	}

	return nil
}

func main() {

	// default to service account in cluster token
	c, err := rest.InClusterConfig()
	if err != nil {
		defaultLog.Error("Failed to read k8s cluster configuration")
		return
	}
	configuration = c

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open log file err %v", err)
		return
	}

	admissionControllerConfig, err := config.GetAdmissionControllerConfig()

	err = configureLogs(logFile, admissionControllerConfig.LogLevel, admissionControllerConfig.LogMaxLength)
	if err != nil {
		defaultLog.Fatalf("Error while configuring logs %v", err)
	}

	router := mux.NewRouter()

	router.HandleFunc("/mutate", HandleMutate).Methods(http.MethodPost)

	err = startServer(router, *admissionControllerConfig)
	if err != nil {
		defaultLog.Error("Error starting server")
		return
	}
	defaultLog.Info("ISecL Admission Controller exit")
}

func HandleMutate(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		defaultLog.Error("Error reading admission controller body")
		return
	}

	var admissionReviewReq v1beta1.AdmissionReview

	//To convert the request body into struct
	if _, _, err := universalDeserializer.Decode(body, nil, &admissionReviewReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		defaultLog.Errorf("could not deserialize request: %v", err)
		return
	} else if admissionReviewReq.Request == nil {
		w.WriteHeader(http.StatusBadRequest)
		defaultLog.Error("malformed admission review: request is nil")
		return
	}

	var node apiv1.Node
	var noScheduleTaint apiv1.Taint
	var noExecuteTaint apiv1.Taint

	err = json.Unmarshal(admissionReviewReq.Request.Object.Raw, &node)

	if err != nil {
		defaultLog.Errorf("could not unmarshal node on admission request: %v", err)
		return
	}

	err = json.Unmarshal(admissionReviewReq.Request.Object.Raw, &node)
	if err != nil {
		defaultLog.Errorf("could not unmarshal node on admission request: %v", err)
		return
	}

	defaultLog.Debugf("node is %v", node)

	taints := node.Spec.Taints
	noScheduleTaint.Key = constants.TaintNameNoschedule
	noScheduleTaint.Value = "true"
	noScheduleTaint.Effect = "NoSchedule"

	noExecuteTaint.Key = constants.TaintNameNoexecute
	noExecuteTaint.Value = "true"
	noExecuteTaint.Effect = "NoExecute"

	taints = append(taints, noScheduleTaint)
	taints = append(taints, noExecuteTaint)

	var patches []patchOperation

	patches = append(patches, patchOperation{
		Op:    "add",
		Path:  "/spec/taints",
		Value: taints,
	})

	//convert to byte array
	patchBytes, err := json.Marshal(patches)
	if err != nil {
		defaultLog.Errorf("could not marshal JSON patch: %v", err)
		return
	}

	admissionReviewResponse := v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			UID:     admissionReviewReq.Request.UID,
			Allowed: true,
		},
	}

	admissionReviewResponse.Response.Patch = patchBytes
	bytes, err := json.Marshal(&admissionReviewResponse)
	if err != nil {
		defaultLog.Errorf("Error while marshaling response: %v", err)
		return
	}

	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	w.Write(bytes)
	defaultLog.Infof("Successfully added taint to Node %v", node.Name)
}
