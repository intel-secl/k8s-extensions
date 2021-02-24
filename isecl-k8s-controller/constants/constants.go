/*
Copyright © 2021 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package constants

// Env param handles
const (
	LogLevelEnv            = "LOG_LEVEL"
	LogMaxLengthEnv        = "LOG_MAX_LENGTH"
	TaintUntrustedNodesEnv = "TAINT_UNTRUSTED_NODES"
	TagPrefixEnv           = "TAG_PREFIX"
	KubeconfEnv            = "KUBECONF"
)

// Default values
const (
	LogLevelDefault            = "INFO"
	LogMaxLengthDefault        = 1500
	TagPrefixDefault           = "isecl."
	TaintUntrustedNodesDefault = false
	FilePerms                  = 0664
)

const (
	WgName         = "iseclcontroller"
	MinThreadiness = 1
	ErrExitCode    = 1
)
