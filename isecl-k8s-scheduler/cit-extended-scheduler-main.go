/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"k8s_scheduler_cit_extension-k8s_extended_scheduler/api"
	"k8s_scheduler_cit_extension-k8s_extended_scheduler/util"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

type Config struct {
	Trusted string `"json":"trusted"`
}

const TrustedPrefixConf = "/opt/isecl-k8s-extensions/config/"

func getPrefixFromConf(path string) (string, error) {
	out, err := ioutil.ReadFile(path)
	if err != nil {
		glog.Errorf("Error: %s %v", path, err)
		return "", err
	}
	s := Config{}
	err = json.Unmarshal(out, &s)
	if err != nil {
		glog.Errorf("Error:  %v", err)
		return "", err
	}
	return s.Trusted, nil
}

func extendedScheduler(c *gin.Context) {
	c.JSON(200, gin.H{"result": "ISecL Extended Scheduler"})
	return
}

func SetupRouter() (*gin.Engine, *http.Server) {
	//get a webserver instance, that contains a muxer, middleware and configuration settings
	router := gin.Default()
	// fetch all the cmd line args
	url, port, server_crt, server_key := util.GetCmdlineArgs()

	//initialize http server config
	server := &http.Server{
		Addr:    url + ":" + port,
		Handler: router,
	}

	//run the server instance
	go func() {
		// service connections
		if err := server.ListenAndServeTLS(server_crt, server_key); err != nil {
			glog.Errorf("listen: %s\n", err)
		}
	}()

	router.GET("/", extendedScheduler)

	return router, server
}

func main() {
	glog.V(4).Infof("Starting ISecL Extended Scheduler...")

	var err error

	api.Confpath, err = getPrefixFromConf(TrustedPrefixConf + "tag_prefix.conf")
	if err != nil {
		log.Fatalf("Error:in trustedprefixconf %v", err)
	}

	router, server := SetupRouter()

	//hadler for the post operation
	router.POST("filter", api.FilterHandler)

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	glog.Infof("Shutting down ISecL Extended Scheduler Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		glog.Infof(" ISecL Extended Scheduler Server Shutdown:", err)
	}
	glog.Infof(" ISecL Extended Scheduler Server exit")
}
