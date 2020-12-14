/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"intel/isecl/k8s-extended-scheduler/v3/constants"
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
	IntegrationHubPublicKeys map[string][]byte

	LogLevel string

	LogMaxLength int

	TagPrefix string
}

func GetExtendedSchedulerConfig() (*Config, error) {

	//PORT for the extended scheduler to listen.
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		fmt.Fprintln(os.Stdout, "Error while parsing Env variable PORT, setting to default value 8888")
		port = 8888
	}
	iHubPublicKeys := make(map[string][]byte, 2)

	// Get IHub public key from ihub with hvs attestation type
	iHubPubKeyPath := os.Getenv("HVS_IHUB_PUBLIC_KEY_PATH")
	if iHubPubKeyPath != "" {
		iHubPublicKeys[constants.HVSAttestation], err = ioutil.ReadFile(iHubPubKeyPath)
		if err != nil {
			return nil, errors.Errorf("Error while reading file %s\n", iHubPubKeyPath)
		}
	}

	// Get IHub public key from ihub with skc attestation type
	iHubPubKeyPath = os.Getenv("SGX_IHUB_PUBLIC_KEY_PATH")
	if iHubPubKeyPath != "" {
		iHubPublicKeys[constants.SGXAttestation], err = ioutil.ReadFile(iHubPubKeyPath)
		if err != nil {
			return nil, errors.Errorf("Error while reading file %s\n", iHubPubKeyPath)
		}
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		fmt.Fprintln(os.Stdout, "Env variable LOG_LEVEL is empty, setting to default value Info")
		logLevel = "INFO"
	}

	logMaxLen, err := strconv.Atoi(os.Getenv("LOG_MAX_LENGTH"))
	if err != nil {
		fmt.Fprintln(os.Stdout, "Env variable LOG_MAX_LENGTH is empty, setting to default value 1500")
		logMaxLen = 1500
	}

	serverCert := os.Getenv("TLS_CERT_PATH")
	if serverCert == "" {
		return nil, errors.New("Env variable TLS_CERT_PATH is empty")
	}

	serverKey := os.Getenv("TLS_KEY_PATH")
	if serverKey == "" {
		return nil, errors.New("Env variable TLS_KEY_PATH is empty")
	}

	tagPrefix := os.Getenv("TAG_PREFIX")
	if !tagPrefixRegex.MatchString(tagPrefix) {
		return nil, errors.New("Invalid string formatted input")
	}

	return &Config{
		Port:                     port,
		IntegrationHubPublicKeys: iHubPublicKeys,
		LogLevel:                 logLevel,
		ServerCert:               serverCert,
		ServerKey:                serverKey,
		TagPrefix:                tagPrefix,
		LogMaxLength:             logMaxLen,
	}, nil
}
