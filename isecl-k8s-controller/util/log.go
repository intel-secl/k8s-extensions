/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package util

import (
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

const LogFile = "/var/log/isecl-k8s-extensions/isecl-k8s-controller.log"

var Log *logrus.Logger

func GetLogger() *logrus.Logger {
	Log = logrus.New()
	logFile, err := os.OpenFile(LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err == nil {
		Log.SetOutput(logFile)
	} else {
		Log.SetOutput(os.Stdout)
	}
	Log.Formatter = &logrus.JSONFormatter{}

	Log.Info("Initialized log")
	return Log
}

func SetLogger(logLevel string) {
	logLevel = strings.ToUpper(logLevel)
	switch logLevel {
	case "DEBUG":
		Log.SetLevel(logrus.DebugLevel)
	case "INFO":
		Log.SetLevel(logrus.InfoLevel)
	case "WARNING":
		Log.SetLevel(logrus.WarnLevel)
	case "ERROR":
		Log.SetLevel(logrus.ErrorLevel)
	default:
		Log.SetLevel(logrus.InfoLevel)
	}
}
