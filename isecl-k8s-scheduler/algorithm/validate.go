/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package algorithm

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	jwt "github.com/Waterdrips/jwt-go"
	"intel/isecl/k8s-extended-scheduler/v3/constants"
	v1 "k8s.io/api/core/v1"
)



func keyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}

//ValidatePodWithAnnotation is to validate signed trusted and location report with pod keys and values
func ValidatePodWithHvsAnnotation(nodeData []v1.NodeSelectorRequirement, claims jwt.MapClaims, trustprefix string)  bool {
	assetClaims := make(map[string]interface{})
	hardwareFeatureClaims := make(map[string]interface{})

	meExistInClaims := false
	if keyExists(claims, constants.AssetTags) {
		assetClaims = claims[constants.AssetTags].(map[string]interface{})
		defaultLog.Infof("ValidatePodWithHvsAnnotation - Validating Asset Tag Claims node: %v, claims: %v", nodeData, assetClaims)
	}

	if keyExists(claims, constants.HardwareFeatures) {
		hardwareFeatureClaims = claims[constants.HardwareFeatures].(map[string]interface{})
		defaultLog.Infof("ValidatePodWithHvsAnnotation - Validating Hardware Feature Claims node: %v, claims: %v", nodeData, hardwareFeatureClaims)
	}

	for _, val := range nodeData {
		if strings.Contains(val.Key, trustprefix) {
			val.Key = strings.Split(val.Key, trustprefix)[1]
		}

		// if val is trusted, it can be directly found in claims
		switch val.Key {
		case trustprefix + constants.Trusted:
			meExistInClaims = true
			for _, nodeVal := range val.Values {
				sigVal := claims[constants.Trusted]
				sigValTemp := sigVal.(bool)
				sigVal = strconv.FormatBool(sigValTemp)
				if nodeVal == sigVal {
					continue
				} else {
					defaultLog.Infof("ValidatePodWithHvsAnnotation - Trust Check - Mismatch in %v field. Actual: %v | In Signature: %v ", val.Key, nodeVal, sigVal)
					return false
				}
			}

		// validate asset tags
		case assetClaims[val.Key]:
			meExistInClaims = true
			aTagVal := assetClaims[val.Key]
			flag := false
			for _, match := range val.Values {
				if match == aTagVal {
					flag = true
				} else {
					defaultLog.Infof("ValidatePodWithHvsAnnotation - Asset Tags - Mismatch in %v field. Actual: %v", val.Key, aTagVal, match)
				}
			}
			if flag {
				continue
			} else {
				return false
			}

		// validate HW features
		case hardwareFeatureClaims[val.Key]:
			meExistInClaims = true
			hwFeatureValue := hardwareFeatureClaims[val.Key]
			flag := false
			for _, match := range val.Values {
				if match == hwFeatureValue {
					flag = true
				} else {
					defaultLog.Infof("ValidatePodWithHvsAnnotation - Hardware Features - Mismatch in %v field. Actual: %v", val.Key, hwFeatureValue, match)
				}
			}
			if flag {
				continue
			} else {
				return false
			}
		}
	}

	// Do not validate expiry for non isecl affinity rules
	if !meExistInClaims{
		return true
	}
	defaultLog.Info("Successfully validated with hvs signed trust report claims")
	trustTimeValid := ValidateNodeByTime(claims, constants.HvsTrustValidTo)

	return trustTimeValid
}

//ValidatePodWithSgxAnnotation is to validate sgx signed trusted and location report with pod keys and values
func ValidatePodWithSgxAnnotation(nodeData []v1.NodeSelectorRequirement, claims jwt.MapClaims, trustprefix string) bool {
	meExistInClaims := false
	for _, val := range nodeData {
		if strings.Contains(val.Key, trustprefix) {
			val.Key = strings.Split(val.Key, trustprefix)[1]
		}
		// if val is trusted, it can be directly found in claims
		switch val.Key {
		// validate SKC features
		case "SGX-Enabled":
			meExistInClaims = true
			sigVal := claims[constants.SgxEnabled]
			for _, nodeVal := range val.Values {

				sigValStr := sigVal.(string)
				if nodeVal == sigValStr {
					continue
				} else {
					defaultLog.Infof("ValidatePodWithSgxAnnotation - Trust Check - Mismatch in %v field. Actual: %v | In Signature: %v ", val.Key, nodeVal, sigVal)
					return false
				}
			}
		case "SGX-Supported":
			meExistInClaims = true
			sigVal := claims[constants.SgxSupported]
			for _, nodeVal := range val.Values {

				sigValStr := sigVal.(string)
				if nodeVal == sigValStr {
					continue
				} else {
					defaultLog.Infof("ValidatePodWithSgxAnnotation - Trust Check - Mismatch in %v field. Actual: %v | In Signature: %v ", val.Key, nodeVal, sigVal)
					return false
				}
			}
		case "TCBUpToDate":
			meExistInClaims = true
			sigVal := claims[constants.TcbUpToDate]
			for _, nodeVal := range val.Values {

				sigValStr := sigVal.(string)
				if nodeVal == sigValStr {
					continue
				} else {
					defaultLog.Infof("ValidatePodWithSgxAnnotation - Trust Check - Mismatch in %v field. Actual: %v | In Signature: %v ", val.Key, nodeVal, sigVal)
					return false
				}
			}
		case "EPC-Memory":
			meExistInClaims = true
			sigVal := claims[constants.EpcSize]
			for _, nodeVal := range val.Values {

				sigValStr := sigVal.(string)
				if nodeVal == sigValStr {
					continue
				} else {
					defaultLog.Infof("ValidatePodWithSgxAnnotation - Trust Check - Mismatch in %v field. Actual: %v | In Signature: %v ", val.Key, nodeVal, sigVal)
					return false
				}
			}
		case "FLC-Enabled":
			meExistInClaims = true
			sigVal := claims[constants.FlcEnabled]
			for _, nodeVal := range val.Values {

				sigValStr := sigVal.(string)
				if nodeVal == sigValStr {
					continue
				} else {
					defaultLog.Infof("ValidatePodWithSgxAnnotation - Trust Check - Mismatch in %v field. Actual: %v | In Signature: %v ", val.Key, nodeVal, sigVal)
					return false
				}
			}

		}
	}

	// Do not validate expiry for non isecl affinity rules or matching expression present in k8s manifests
	if !meExistInClaims{
		return true
	}

	defaultLog.Info("Successfully validated with sgx signed trust report claims")

	trustTimeValid := ValidateNodeByTime(claims, constants.SgxTrustValidTo)
	return trustTimeValid
}

//ValidateNodeByTime is used for validate time for each node with current system time(Expiry validation)
func ValidateNodeByTime(claims jwt.MapClaims, validTo string) bool {
	if timeVal, ok := claims[validTo].(string); ok {

		reg, err := regexp.Compile("[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+")
		defaultLog.Debugf("ValidateNodeByTime reg %v", reg)
		if err != nil {
			defaultLog.Errorf("Error parsing validTo time: %v", err)
			return false
		}
		newstr := reg.ReplaceAllString(timeVal, "")
		defaultLog.Debugf("ValidateNodeByTime newstr: %s", newstr)
		trustedValidToTime := strings.Replace(timeVal, newstr, "", -1)
		defaultLog.Infof("ValidateNodeByTime trustedValidToTime: %s", trustedValidToTime)

		t := time.Now().UTC()
		timeDiff := strings.Compare(trustedValidToTime, t.Format(time.RFC3339))
		defaultLog.Infof("ValidateNodeByTime - ValidTo - %s |  current - %s | Diff - %d", trustedValidToTime, timeVal, timeDiff)
		if timeDiff >= 0 {
			defaultLog.Infof("ValidateNodeByTime Attested node validity time check passed -timeDiff: %d ", timeDiff)
			return true
		} else {
			defaultLog.Infof("ValidateNodeByTime - Node outside expiry time - ValidTo - %s |  current - %s", timeVal, t)
			return false
		}
	}
	return false
}
