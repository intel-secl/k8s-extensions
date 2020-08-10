/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"intel/isecl/k8s-extended-scheduler/v2/api"
	"intel/isecl/k8s-extended-scheduler/v2/util"
	"io/ioutil"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Config struct {
	Trusted string `"json":"trusted"`
}

const TrustedPrefixConf = "/opt/isecl-k8s-extensions/config/"

var Log = util.GetLogger()

func getPrefixFromConf(path string) (string, error) {
	out, err := ioutil.ReadFile(path)
	if err != nil {
		Log.Errorf("Error: %s %v", path, err)
		return "", err
	}
	s := Config{}
	err = json.Unmarshal(out, &s)
	if err != nil {
		Log.Errorf("Error:  %v", err)
		return "", err
	}
	return s.Trusted, nil
}

func extendedScheduler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = bytes.NewBuffer([]byte("ISecL Extended Scheduler")).WriteTo(w)
	return
}

func startServer(router *mux.Router) error {
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

	// fetch all the cmd line args
	port, serverCrt, serverKey := util.GetCmdlineArgs()

	//initialize http server config
	httpWriter := os.Stderr
	if httpLogFile, err := os.OpenFile(util.HttpLogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666); err != nil {
		Log.Tracef("service:Start() %+v", err)
	} else {
		defer httpLogFile.Close()
		httpWriter = httpLogFile
	}

	httpLog := stdlog.New(httpWriter, "", 0)
	h := &http.Server{
		Addr:      fmt.Sprintf(":%d", port),
		Handler:   handlers.RecoveryHandler(handlers.RecoveryLogger(httpLog), handlers.PrintRecoveryStack(true))(handlers.CombinedLoggingHandler(os.Stderr, router)),
		ErrorLog:  httpLog,
		TLSConfig: tlsconfig,
	}

	//run the server instance

	// dispatch web server go routine
	go func() {
		if err := h.ListenAndServeTLS(serverCrt, serverKey); err != nil {
			Log.Errorf("failed to start service %+v", err)
			stop <- syscall.SIGTERM
		}
	}()
	Log.Info("Service started")

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
	Log.Info("Starting ISecL Extended Scheduler...")
	var err error

	logLevel := flag.String("loglevel", "debug", "Path to a kube config. ")
	flag.Parse()
	util.SetLogger(*logLevel)
	api.Confpath, err = getPrefixFromConf(TrustedPrefixConf + "tag_prefix.conf")

	if err != nil {
		Log.Fatalf("Error while parsing tag prefix %v", err)
	}

	router := mux.NewRouter()

	//hadler for the post operation
	router.HandleFunc("/filter", api.FilterHandler).Methods("POST")
	router.HandleFunc("/", extendedScheduler).Methods("GET")

	err = startServer(router)
	if err != nil {
		Log.Error("Error starting server")
	}
	Log.Infof(" ISecL Extended Scheduler Server exit")
}
