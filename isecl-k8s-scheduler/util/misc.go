/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package util

import (
	"strconv"
	"os"
	"strings"
	"github.com/tkanos/gonfig"
	"github.com/sirupsen/logrus"
)

var AH_KEY_FILE string
const LogFile = "/var/log/isecl-k8s-extensions/isecl-k8s-scheduler.log"
const SchedConf = "/opt/isecl-k8s-extensions/isecl-k8s-scheduler/config/isecl-extended-scheduler-config.json"

var Log *logrus.Logger
func GetLogger() *logrus.Logger{
	Log = logrus.New()
        logFile, err := os.OpenFile(LogFile, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0755)
        if err != nil {
                Log.Fatal(err)
        }
        Log.Formatter = &logrus.JSONFormatter{}
        Log.SetOutput(logFile)
	Log.Info("Initialized log")
	return Log
}

func SetLogger(logLevel string){
	logLevel = strings.ToUpper(logLevel)
	switch logLevel{
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



func GetCmdlineArgs() (string, string, string, string) {
	type extenedSchedConfig struct {
		Url  string //Extended scheduler url
		Port int    //Port for the Extended scheduler to listen on
		//Server Certificate to be used for TLS handshake
		ServerCert string
		//Server Key to be used for TLS handshake
		ServerKey string
		//Attestation Hub Key to be used for parsing signed trust report
		AttestationHubKey string
	}

	conf := extenedSchedConfig{}
	err := gonfig.GetConf(SchedConf, &conf)
	if err != nil {
		Log.Fatalf("Error: Please ensure extended schduler configuration is present in curent dir,%v", err)
	}

	//PORT for the extended scheduler to listen.
	port_no := conf.Port
	port := strconv.Itoa(port_no)

	AH_KEY_FILE = (conf.AttestationHubKey)
	return conf.Url, port, conf.ServerCert, conf.ServerKey
}
