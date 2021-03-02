/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package main

import (
	"fmt"
	commLog "github.com/intel-secl/intel-secl/v3/pkg/lib/common/log"
	commLogMsg "github.com/intel-secl/intel-secl/v3/pkg/lib/common/log/message"
	commLogInt "github.com/intel-secl/intel-secl/v3/pkg/lib/common/log/setup"
	"intel/isecl/k8s-custom-controller/v3/constants"
	"io"
	"regexp"
	"strings"

	"intel/isecl/k8s-custom-controller/v3/crdController"
	"os"
	"strconv"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

// GetClientConfig returns rest config, if path not specified assume in cluster config
func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

const logFilePath = "/var/log/isecl-k8s-extensions/isecl-controller.log"

var (
	defaultLog     = commLog.GetDefaultLogger()
	tagPrefixRegex = regexp.MustCompile("(^[a-zA-Z0-9_///.-]*$)")
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

func main() {

	fmt.Println("Starting ISecL Custom Controller")

	var (
		logMaxLength        int
		logLevel            string
		taintUntrustedNodes bool
		err                 error
	)

	logLevelEnv := os.Getenv(constants.LogLevelEnv)
	if logLevelEnv == "" {
		fmt.Printf("%s cannot be empty setting to default value %s",
			constants.LogLevelEnv, constants.LogLevelDefault)
		logLevel = constants.LogLevelDefault
	} else {
		logrusLvl, err := logrus.ParseLevel(strings.ToUpper(logLevelEnv))
		if err != nil {
			fmt.Printf("%s is invalid loglevel. Setting to default value %s",
				constants.LogLevelEnv, constants.LogLevelDefault)
			logLevel = constants.LogLevelDefault
		} else {
			logLevel = logrusLvl.String()
		}
	}

	logMaxLengthEnv := os.Getenv(constants.LogMaxLengthEnv)
	if logMaxLengthEnv == "" {
		fmt.Printf("%s cannot be empty setting to default value %d",
			constants.LogMaxLengthEnv, constants.LogMaxLengthDefault)
		logMaxLength = constants.LogMaxLengthDefault
	} else if logMaxLength, err = strconv.Atoi(logMaxLengthEnv); err != nil {
		fmt.Printf("Error while parsing variable config %s error: %v, defaulting to %d \n",
			constants.LogMaxLengthEnv, err, constants.LogMaxLengthDefault)
		logMaxLength = constants.LogMaxLengthDefault
	} else if logMaxLength <= 0 {
		fmt.Printf("%s should be > 0, defaulting to %d\n",
			constants.LogMaxLengthEnv, constants.LogMaxLengthDefault)
		logMaxLength = constants.LogMaxLengthDefault
	}

	taintUntrustedNodesEnv := os.Getenv(constants.TaintUntrustedNodesEnv)
	if taintUntrustedNodesEnv == "" {
		fmt.Printf("%s cannot be empty setting to default value %d",
			constants.TaintUntrustedNodesEnv, constants.TaintUntrustedNodesDefault)
		taintUntrustedNodes = constants.TaintUntrustedNodesDefault
	} else if taintUntrustedNodes, err = strconv.ParseBool(taintUntrustedNodesEnv); err != nil {
		fmt.Printf("Error while parsing variable config %s error: %v, defaulting to %d \n",
			constants.TaintUntrustedNodesEnv, err, constants.TaintUntrustedNodesDefault)
		taintUntrustedNodes = constants.TaintUntrustedNodesDefault
	}

	tagPrefix := os.Getenv(constants.TagPrefixEnv)
	if tagPrefix == "" {
		fmt.Printf("%s cannot be empty setting to default value %d",
			constants.TagPrefixEnv, constants.TagPrefixDefault)
		tagPrefix = constants.TagPrefixDefault
	} else if !tagPrefixRegex.MatchString(tagPrefix) {
		fmt.Fprintf(os.Stderr, "%s has an unsupported value. Exiting.", constants.TagPrefixEnv)
		os.Exit(constants.ErrExitCode)
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, constants.FilePerms)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open log file")
		return
	}

	// configure logs
	err = configureLogs(logFile, logLevel, logMaxLength)
	if err != nil {
		defaultLog.Fatalf("Error while configuring logs %v", err)
	}

	// load cluster configuration
	kubeConf := os.Getenv(constants.KubeconfEnv)
	config, err := GetClientConfig(kubeConf)
	if err != nil {
		defaultLog.Errorf("Error in config %v", err)
		return
	}

	crdController.TaintUntrustedNodes = taintUntrustedNodes

	// Create a queue
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), constants.WgName)

	indexer, informer := crdController.NewIseclHAIndexerInformer(config, queue, &sync.Mutex{}, tagPrefix)

	controller := crdController.NewIseclHAController(queue, indexer, informer)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(constants.MinThreadiness, stop)

	defaultLog.Info("Waiting for updates on ISecl Custom Resource Definitions")

	// Wait forever
	select {}
}
