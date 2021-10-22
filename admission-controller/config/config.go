/*
Copyright © 2021 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package config

import (
	"fmt"
	"github.com/intel-secl/k8s-extensions/v4/admission-controller/constants"
	"github.com/sirupsen/logrus"
	"os"

	"strconv"
	"strings"
)

type Config struct {
	Port int //Port for the admission controller to listen on
	//Server Certificate to be used for TLS handshake
	ServerCert string
	//Server Key to be used for TLS handshake
	ServerKey string

	LogLevel string

	LogMaxLength int
}

func GetAdmissionControllerConfig() (*Config, error) {

	var (
		port         int
		logMaxLength int
		logLevel     string
		err          error
	)

	//PORT for the extended scheduler to listen.
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

	portEnv := os.Getenv(constants.PortEnv)
	if portEnv == "" {
		fmt.Printf("%s cannot be empty setting to default value %d",
			constants.PortEnv, constants.PortDefault)
		port = constants.PortDefault
	} else if port, err = strconv.Atoi(portEnv); err != nil {
		fmt.Printf("Error while parsing variable config %s error: %v, defaulting to %d \n",
			constants.PortEnv, err, constants.PortDefault)
		port = constants.PortDefault
	} else if port <= 0 {
		fmt.Printf("%s should be > 0, defaulting to %d\n",
			constants.PortEnv, constants.PortDefault)
		port = constants.PortDefault
	}

	return &Config{
		Port:         port,
		LogLevel:     logLevel,
		ServerCert:   constants.TlsCertPath,
		ServerKey:    constants.TlsKeyPath,
		LogMaxLength: logMaxLength,
	}, nil
}
