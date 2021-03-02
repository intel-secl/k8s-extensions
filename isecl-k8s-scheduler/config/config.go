/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"intel/isecl/k8s-extended-scheduler/v3/constants"
)

var tagPrefixRegex = regexp.MustCompile("(^[a-zA-Z0-9_///.-]*$)")

type Config struct {
	Port int //Port for the Extended scheduler to listen on
	//Server Certificate to be used for TLS handshake
	ServerCert string
	//Server Key to be used for TLS handshake
	ServerKey string
	//Integration Hub Key to be used for parsing signed trust report
	IntegrationHubPublicKeys map[string][]byte

	LogLevel string

	LogMaxLength int

	TagPrefix string
}

func GetExtendedSchedulerConfig() (*Config, error) {

	var (
		port         int
		logMaxLength int
		logLevel     string
		err          error
	)

	//PORT for the extended scheduler to listen.
	portEnv := os.Getenv(constants.PortEnv)
	if portEnv == "" {
		fmt.Printf("%s cannot be empty setting to default value %d\n",
			constants.PortEnv, constants.PortDefault)
		port = constants.PortDefault
	} else if port, err = strconv.Atoi(portEnv); err != nil {
		fmt.Printf("Error while parsing variable config %s error: %v, "+
			"defaulting to %d \n", constants.PortEnv, err, constants.PortDefault)
		port = constants.PortDefault
	}

	iHubPublicKeys := make(map[string][]byte, 2)

	iHubPubKeyPath := filepath.Clean(strings.TrimSpace(os.Getenv(constants.HvsIhubPubKeyPathEnv)))
	if iHubPubKeyPath != "." {
		iHubPublicKeys[constants.HVSAttestation], err = ioutil.ReadFile(iHubPubKeyPath)
		if err != nil {
			return nil, errors.Errorf("Error while reading file %s - %+v", iHubPubKeyPath, err)
		}
	}

	// Get IHub public key from ihub with skc attestation type
	iHubPubKeyPath = filepath.Clean(strings.TrimSpace(os.Getenv(constants.SgxIhubPubKeyPathEnv)))
	if iHubPubKeyPath != "." {
		iHubPublicKeys[constants.SGXAttestation], err = ioutil.ReadFile(iHubPubKeyPath)
		if err != nil {
			return nil, errors.Errorf("Error while reading file %s - %+v", iHubPubKeyPath, err)
		}
	}

	if len(iHubPublicKeys) == 0 {
		return nil, errors.Errorf("IHub public key must be set through %s or %s",
			constants.SgxIhubPubKeyPathEnv, constants.HvsIhubPubKeyPathEnv)
	}

	logLevelEnv := os.Getenv(constants.LogLevelEnv)
	if logLevelEnv == "" {
		fmt.Printf("%s cannot be empty setting to default value %s\n",
			constants.LogLevelEnv, constants.LogLevelDefault)
		logLevel = constants.LogLevelDefault
	} else {
		logrusLvl, err := logrus.ParseLevel(strings.ToUpper(logLevelEnv))
		if err != nil {
			fmt.Printf("%s is invalid loglevel. Setting to default value %s\n",
				constants.LogLevelEnv, constants.LogLevelDefault)
			logLevel = constants.LogLevelDefault
		} else {
			logLevel = logrusLvl.String()
		}
	}

	logMaxLengthEnv := os.Getenv(constants.LogMaxLengthEnv)
	if logMaxLengthEnv == "" {
		fmt.Printf("%s cannot be empty setting to default value %d\n",
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

	serverCert := os.Getenv(constants.TlsCertPathEnv)
	if serverCert == "" {
		return nil, fmt.Errorf("env variable %s is empty", constants.TlsCertPathEnv)
	}

	serverKey := os.Getenv(constants.TlsKeyPath)
	if serverKey == "" {
		return nil, fmt.Errorf("env variable %s is empty", constants.TlsKeyPath)
	}

	tagPrefix := os.Getenv(constants.TagPrefixEnv)
	if tagPrefix == "" {
		fmt.Printf("%s cannot be empty setting to default value %d\n",
			constants.TagPrefixEnv, constants.TagPrefixDefault)
		tagPrefix = constants.TagPrefixDefault
	} else if !tagPrefixRegex.MatchString(tagPrefix) {
		return nil, fmt.Errorf("invalid string formatted input for %s", constants.TagPrefixEnv)
	}

	return &Config{
		Port:                     port,
		IntegrationHubPublicKeys: iHubPublicKeys,
		LogLevel:                 logLevel,
		ServerCert:               serverCert,
		ServerKey:                serverKey,
		TagPrefix:                tagPrefix,
		LogMaxLength:             logMaxLength,
	}, nil
}
