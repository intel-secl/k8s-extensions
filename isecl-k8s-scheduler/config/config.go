/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

var tagPrefixRegex = regexp.MustCompile("(^[a-zA-Z0-9_///.-]*$)")

const LogFile = "/var/log/isecl-k8s-extensions/isecl-k8s-scheduler.log"
const HttpLogFile = "/var/log/isecl-k8s-extensions/isecl-k8s-scheduler-http.log"

var Log *logrus.Logger

type Config struct {
	Port int //Port for the Extended scheduler to listen on
	//Server Certificate to be used for TLS handshake
	ServerCert string
	//Server Key to be used for TLS handshake
	ServerKey string
	//Integration Hub Key to be used for parsing signed trust report
	IntegrationHubPublicKey string

	LogLevel string

	LogMaxLength int

	TagPrefix string
}

func (e *Config) GetLogger() *logrus.Logger {
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

func (e *Config) SetLogger() {
	logLevel := strings.ToUpper(e.LogLevel)
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

func GetExtendedSchedulerConfig() (*Config, error) {

	//PORT for the extended scheduler to listen.
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil{
		fmt.Fprintln(os.Stdout, "Error while parsing Env variable PORT, setting to default value 8888")
		port = 8888
	}

	integrationHubPublicKey := os.Getenv("IHUB_PUBLIC_KEY_FILE_PATH")
	if integrationHubPublicKey == ""{
		return nil, errors.New("Env variable IHUB_PUBLIC_KEY_FILE_PATH is empty")
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == ""{
		fmt.Fprintln(os.Stdout,"Env variable LOG_LEVEL is empty, setting to default value Info")
		logLevel = "INFO"
	}

	logMaxLen, err := strconv.Atoi(os.Getenv("LOG_MAX_LENGTH"))
	if err != nil{
		fmt.Fprintln(os.Stdout, "Env variable LOG_MAX_LENGTH is empty, setting to default value 1500")
		logMaxLen = 500
	}

	serverCert := os.Getenv("TLS_CERT_PATH")
	if serverCert == ""{
		return nil, errors.New("Env variable TLS_CERT_PATH is empty")
	}

	serverKey := os.Getenv("TLS_KEY_PATH")
	if serverKey == ""{
		return nil, errors.New("Env variable TLS_KEY_PATH is empty")
	}

	tagPrefix := os.Getenv("TLS_KEY_PATH")
	if !tagPrefixRegex.MatchString(tagPrefix) {
		return nil, errors.New("Invalid string formatted input")
	}

	return &Config{
		Port: port,
		IntegrationHubPublicKey: integrationHubPublicKey,
		LogLevel: logLevel,
		ServerCert: serverCert,
		ServerKey: serverKey,
		TagPrefix: tagPrefix,
		LogMaxLength: logMaxLen,
	}, nil
}
