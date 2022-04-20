/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package constants

const (
	InstallPath = "/opt/isecl-k8s-extensions/"
	LogFilePath = "/var/log/isecl-k8s-scheduler/isecl-scheduler.log"
	HttpLogFile = "/var/log/isecl-k8s-scheduler/isecl-k8s-scheduler-http.log"
	FilePerms   = 0664
)

const (
	HVSAttestation       = "HVS"
	SGXAttestation       = "SGX"
	HvsSignedTrustReport = "HvsSignedTrustReport"
	SgxSignedTrustReport = "SgxSignedTrustReport"
)

const (
	AssetTags        = "assetTags"
	HardwareFeatures = "hardwareFeatures"
	HvsTrustValidTo  = "hvsTrustValidTo"
)

const (
	SgxEnabled      = "sgxEnabled"
	SgxSupported    = "sgxSupported"
	TcbUpToDate     = "tcbUpToDate"
	EpcSize         = "epcSize"
	FlcEnabled      = "flcEnabled"
	SgxTrustValidTo = "sgxTrustValidTo"
)

// Env param handles
const (
	PortEnv              = "PORT"
	HvsIhubPubKeyPathEnv = "HVS_IHUB_PUBLIC_KEY_PATH"
	SgxIhubPubKeyPathEnv = "SGX_IHUB_PUBLIC_KEY_PATH"
	TlsCertPathEnv       = "TLS_CERT_PATH"
	TlsKeyPath           = "TLS_KEY_PATH"
	TagPrefixEnv         = "TAG_PREFIX"
	LogLevelEnv          = "LOG_LEVEL"
	LogMaxLengthEnv      = "LOG_MAX_LENGTH"
)

// Default values
const (
	LogLevelDefault     = "INFO"
	LogMaxLengthDefault = 1500
	TagPrefixDefault    = "isecl."
	PortDefault         = 8888
)
