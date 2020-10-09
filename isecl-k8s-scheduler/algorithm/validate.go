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

	jwt "github.com/dgrijalva/jwt-go"
	v1 "k8s.io/api/core/v1"
)

const (
	ahreport   string = "assetTags"
	trusted    string = "trusted"
	hwFeatures string = "hardwareFeatures"
)

func keyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}

//ValidatePodWithAnnotation is to validate signed trusted and location report with pod keys and values
func ValidatePodWithAnnotation(nodeData []v1.NodeSelectorRequirement, claims jwt.MapClaims, trustprefix string) (bool, bool) {
	iseclLabelExists := false
	assetClaims := make(map[string]interface{})
	hardwareFeatureClaims := make(map[string]interface{})

	if keyExists(claims, ahreport) {
		assetClaims = claims[ahreport].(map[string]interface{})
		defaultLog.Infof("ValidatePodWithAnnotation - Validating Asset Tag Claims node: %v, claims: %v", nodeData, assetClaims)
	}

	if keyExists(claims, hwFeatures) {
		hardwareFeatureClaims = claims[hwFeatures].(map[string]interface{})
		defaultLog.Infof("ValidatePodWithAnnotation - Validating Hardware Feature Claims node: %v, claims: %v", nodeData, hardwareFeatureClaims)
	}

	for _, val := range nodeData {
		if strings.Contains(val.Key, trustprefix) {
			iseclLabelExists = true
			val.Key = strings.Split(val.Key, trustprefix)[1]
		}

		// if val is trusted, it can be directly found in claims
		if sigVal, ok := claims[trusted]; ok {
			tr := trustprefix + trusted
			if val.Key == tr {
				for _, nodeVal := range val.Values {
					if sigVal == true || sigVal == false {
						sigValTemp := sigVal.(bool)
						sigVal := strconv.FormatBool(sigValTemp)
						if nodeVal == sigVal {
							continue
						} else {
							defaultLog.Infof("ValidatePodWithAnnotation - Trust Check - Mismatch in %v field. Actual: %v | In Signature: %v ", val.Key, nodeVal, sigVal)
							return false, iseclLabelExists
						}
					} else {
						if nodeVal == sigVal {
							continue
						} else {
							defaultLog.Infof("ValidatePodWithAnnotation - Trust Check - Mismatch in %v field. Actual: %v | In Signature: %v ", val.Key, nodeVal, sigVal)
							return false, iseclLabelExists
						}
					}
				}
			}

			// validate asset tags
			if aTagVal, ok := assetClaims[val.Key]; ok {
				flag := false
				for _, match := range val.Values {
					if match == aTagVal {
						flag = true
					} else {
						defaultLog.Infof("ValidatePodWithAnnotation - Asset Tags - Mismatch in %v field. Actual: %v", val.Key, aTagVal, match)
					}
				}
				if flag {
					continue
				} else {
					return false, iseclLabelExists
				}
			}

			// validate HW features
			if hwKey, ok := hardwareFeatureClaims[val.Key]; ok {
				flag := false
				for _, match := range val.Values {
					if match == hwKey {
						flag = true
					} else {
						defaultLog.Infof("ValidatePodWithAnnotation - Hardware Features - Mismatch in %v field. Actual: %v", val.Key, hwKey, match)
					}
				}
				if flag {
					continue
				} else {
					return false, iseclLabelExists
				}
			}
		}
	}
	return true, iseclLabelExists
}

//ValidateNodeByTime is used for validate time for each node with current system time(Expiry validation)
func ValidateNodeByTime(claims jwt.MapClaims) int {
	trustedTimeFlag := 0
	if timeVal, ok := claims["validTo"].(string); ok {

		reg, err := regexp.Compile("[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+")
		defaultLog.Debugf("ValidateNodeByTime reg %v", reg)
		if err != nil {
			defaultLog.Errorf("Error parsing validTo time: %v", err)
			return trustedTimeFlag
		}
		newstr := reg.ReplaceAllString(timeVal, "")
		defaultLog.Debugf("ValidateNodeByTime newstr: %s", newstr)
		trustedValidToTime := strings.Replace(timeVal, newstr, "", -1)
		defaultLog.Infof("ValidateNodeByTime trustedValidToTime: %s", trustedValidToTime)

		t := time.Now().UTC()
		timeDiff := strings.Compare(trustedValidToTime, t.Format(time.RFC3339))
		defaultLog.Infof("ValidateNodeByTime - ValidTo - %s |  current - %s | Diff - %d", trustedValidToTime, timeVal, timeDiff)
		if timeDiff >= 0 {
			defaultLog.Infof("ValidateNodeByTime -timeDiff: %d ", timeDiff)
			trustedTimeFlag = 1
			defaultLog.Infof("ValidateNodeByTime -trustedTimeFlag: %d", trustedTimeFlag)
		} else {
			defaultLog.Infof("ValidateNodeByTime - Node outside expiry time - ValidTo - %s |  current - %s", timeVal, t)
		}

	}

	return trustedTimeFlag
}
