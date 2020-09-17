/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package algorithm

import (
	"regexp"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	v1 "k8s.io/api/core/v1"
)

const (
	ahreport string = "assetTags"
	trusted  string = "trusted"
	sgxEnabled string = "sgx-enabled"
	sgxSupported string = "sgx-supported"
	tcbUpToDate string = "tcbUpToDate"
	epcSize string = "epc-size"
	flcEnabled string = "flc-enabled"
)

func keyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}

//ValidatePodWithAnnotation is to validate signed trusted and location report with pod keys and values
func ValidatePodWithAnnotation(nodeData []v1.NodeSelectorRequirement, claims jwt.MapClaims, trustprefix string) bool {

	assetClaims := make(map[string]interface{})
	if keyExists(claims, ahreport) {
		assetClaims = claims[ahreport].(map[string]interface{})
		Log.Infof("ValidatePodWithAnnotation - Validating node: %v, claims: %v", nodeData, assetClaims)
	}
	for _, val := range nodeData {
		// if val is trusted, it can be directly found in claims
		switch val.Key {
		case "SGX-Enabled":
			sigVal := claims[sgxEnabled]
			for _, nodeVal := range val.Values {

				sigValStr := sigVal.(string)
				if nodeVal == sigValStr {
					continue
				} else {
					Log.Infof("ValidatePodWithAnnotation - Trust Check - Mismatch in %v field. Actual: %v | In Signature: %v ", val.Key, nodeVal, sigVal)
					return false
				}
			}
		case "SGX-Supported":
			sigVal := claims[sgxSupported]
			for _, nodeVal := range val.Values {

				sigValStr := sigVal.(string)
				if nodeVal == sigValStr {
					continue
				} else {
					Log.Infof("ValidatePodWithAnnotation - Trust Check - Mismatch in %v field. Actual: %v | In Signature: %v ", val.Key, nodeVal, sigVal)
					return false
				}
			}
		case "TCBUpToDate":
			sigVal := claims[tcbUpToDate]
			for _, nodeVal := range val.Values {

				sigValStr := sigVal.(string)
				if nodeVal == sigValStr {
					continue
				} else {
					Log.Infof("ValidatePodWithAnnotation - Trust Check - Mismatch in %v field. Actual: %v | In Signature: %v ", val.Key, nodeVal, sigVal)
					return false
				}
			}
		case "EPC-Memory":
			sigVal := claims[epcSize]
			for _, nodeVal := range val.Values {

				sigValStr := sigVal.(string)
				if nodeVal == sigValStr {
					continue
				} else {
					Log.Infof("ValidatePodWithAnnotation - Trust Check - Mismatch in %v field. Actual: %v | In Signature: %v ", val.Key, nodeVal, sigVal)
					return false
				}
			}
		case "FLC-Enabled":
			sigVal := claims[flcEnabled]
			for _, nodeVal := range val.Values {

				sigValStr := sigVal.(string)
				if nodeVal == sigValStr {
					continue
				} else {
					Log.Infof("ValidatePodWithAnnotation - Trust Check - Mismatch in %v field. Actual: %v | In Signature: %v ", val.Key, nodeVal, sigVal)
					return false
				}
			}
		}

	}
return true
}

//ValidateNodeByTime is used for validate time for each node with current system time(Expiry validation)
func ValidateNodeByTime(claims jwt.MapClaims) int {
	trustedTimeFlag := 0
	if timeVal, ok := claims["validTo"].(string); ok {
		reg, err := regexp.Compile("[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+")
		Log.Infoln("ValidateNodeByTime reg %v err %v", reg, err)
		if err != nil {
			Log.Errorf("Error parsing validTo time: %v", err)
			return trustedTimeFlag
		}
		newstr := reg.ReplaceAllString(timeVal, "")
		Log.Infoln("ValidateNodeByTime newstr %v ", newstr)
		trustedValidToTime := strings.Replace(timeVal, newstr, "", -1)
		Log.Infoln("ValidateNodeByTime trustedValidToTime %v ", trustedValidToTime)

		t := time.Now().UTC()
		timeDiff := strings.Compare(trustedValidToTime, t.Format(time.RFC3339))
		Log.Infof("ValidateNodeByTime - ValidTo - %s |  current - %s | Diff - %d", trustedValidToTime, timeVal, timeDiff)
		if timeDiff >= 0 {
			Log.Infof("ValidateNodeByTime -timeDiff: %d ", timeDiff)
			trustedTimeFlag = 1
			Log.Infoln("ValidateNodeByTime -trustedTimeFlag ", trustedTimeFlag)
		} else {
			Log.Infof("ValidateNodeByTime - Node outside expiry time - ValidTo - %s |  current - %s", timeVal, t)
		}

	}

	return trustedTimeFlag
}
