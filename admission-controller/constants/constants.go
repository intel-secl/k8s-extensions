/*
Copyright © 2021 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package constants

const (
	LogLevelEnv     = "LOG_LEVEL"
	LogMaxLengthEnv = "LOG_MAX_LENGTH"
	PortEnv         = "PORT"
	HttpLogFile     = "/var/log/isecl-k8s-extensions/isecl-admission-controller-http.log"
)

const (
	LogLevelDefault     = "INFO"
	LogMaxLengthDefault = 1500
	PortDefault         = 8889
	TlsCertPath         = "/etc/webhook/certs/tls.crt"
	TlsKeyPath          = "/etc/webhook/certs/tls.key"
)

const (
	TaintNameNoschedule = "untrusted"
	TaintNameNoexecute  = "untrusted"
)
