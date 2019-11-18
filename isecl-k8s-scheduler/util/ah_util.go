/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package util

import (
	"io/ioutil"
)

//GetAHPublicKey is used for reading and return the public key from particular file location
func GetAHPublicKey() []byte {
	pubKey, err := ioutil.ReadFile(AH_KEY_FILE)
	if err != nil {
		Log.Fatalf("Error in reading the hub pem file,%v", err)
	}
	return pubKey
}
