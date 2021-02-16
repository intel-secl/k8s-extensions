/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	commLog "github.com/intel-secl/intel-secl/v3/pkg/lib/common/log"
	commLogMsg "github.com/intel-secl/intel-secl/v3/pkg/lib/common/log/message"
	commLogInt "github.com/intel-secl/intel-secl/v3/pkg/lib/common/log/setup"
	"intel/isecl/k8s-extended-scheduler/v3/api"
	"intel/isecl/k8s-extended-scheduler/v3/config"
	"io"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var defaultLog = commLog.GetDefaultLogger()

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

func extendedScheduler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = bytes.NewBuffer([]byte("ISecL Extended Scheduler")).WriteTo(w)
	return
}

func startServer(router *mux.Router, extenedSchedulerConfig config.Config) error {
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
	if httpLogFile, err := os.OpenFile(config.HttpLogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666); err != nil {
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
		Addr:      fmt.Sprintf(":%d", extenedSchedulerConfig.Port),
		Handler:   handlers.RecoveryHandler(handlers.RecoveryLogger(httpLog), handlers.PrintRecoveryStack(true))(handlers.CombinedLoggingHandler(os.Stderr, router)),
		ErrorLog:  httpLog,
		TLSConfig: tlsconfig,
	}

	//run the server instance

	// dispatch web server go routine
	go func() {
		if err := h.ListenAndServeTLS(extenedSchedulerConfig.ServerCert, extenedSchedulerConfig.ServerKey); err != nil {
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
	var err error

	logFile, err := os.OpenFile("/var/log/isecl-k8s-extensions/isecl-scheduler.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
	if err != nil {
		fmt.Println("Unable to open log file")
		return
	}

	// fetch all the cmd line args
	extendedSchedConfig, err := config.GetExtendedSchedulerConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while getting parsing variables %v\n", err.Error())
		return
	}

	err = configureLogs(logFile, extendedSchedConfig.LogLevel, extendedSchedConfig.LogMaxLength)
	if err != nil {
		defaultLog.Fatalf("Error while configuring logs %v", err)
	}

	router := mux.NewRouter()

	resourceStore := api.ResourceStore{
		IHubPubKeys: extendedSchedConfig.IntegrationHubPublicKeys,
		TagPrefix:   extendedSchedConfig.TagPrefix,
	}
	filterHandler := api.FilterHandler{ResourceStore: resourceStore}
	//handler for the post operation
	router.HandleFunc("/filter", filterHandler.Filter).Methods("POST")
	router.HandleFunc("/", extendedScheduler).Methods("GET")

	err = startServer(router, *extendedSchedConfig)
	if err != nil {
		defaultLog.Error("Error starting server")
	}
	defaultLog.Infof(" ISecL Extended Scheduler Server exit")
}
