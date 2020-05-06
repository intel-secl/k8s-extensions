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
	ahreport string = "asset_tags"
	trusted  string = "trusted"
)

func keyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}

//ValidatePodWithAnnotation is to validate signed trusted and location report with pod keys and values
func ValidatePodWithAnnotation(nodeData []v1.NodeSelectorRequirement, claims jwt.MapClaims, trustprefix string) (bool, bool) {
	iseclLabelExists := false
	if !keyExists(claims, ahreport) {
		Log.Errorf("ValidatePodWithAnnotation - Asset Tags not found for node.")
		return false, iseclLabelExists
	}

	assetClaims := claims[ahreport].(map[string]interface{})
	Log.Infof("ValidatePodWithAnnotation - Validating node: %v, claims: %v", nodeData, assetClaims)

	for _, val := range nodeData {
		if strings.Contains(val.Key, trustprefix) {
			iseclLabelExists = true
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
							Log.Infof("ValidatePodWithAnnotation - Trust Check - Mismatch in %v field. Actual: %v | In Signature: %v ", val.Key, nodeVal, sigVal)
							return false, iseclLabelExists
						}
					} else {
						if nodeVal == sigVal {
							continue
						} else {
							Log.Infof("ValidatePodWithAnnotation - Trust Check - Mismatch in %v field. Actual: %v | In Signature: %v ", val.Key, nodeVal, sigVal)
							return false, iseclLabelExists
						}
					}
				}
			}
		} else {
			if geoKey, ok := assetClaims[val.Key]; ok {
				assetTagList, ok := geoKey.([]interface{})
				if ok {
					flag := false
					//Taking only first value from asset tag list assuming only one value will be there
					geoVal := assetTagList[0]
					newVal := geoVal.(string)
					newVal = strings.Replace(newVal, " ", "", -1)
					newVal = trustprefix + newVal
					for _, match := range val.Values {
						if match == newVal {
							flag = true
						} else {
							Log.Infof("ValidatePodWithAnnotation - Geo Asset Tags - Mismatch in %v field. Actual: %v | In Signature: %v ", geoKey, match, newVal)
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
	}
	return true, iseclLabelExists
}

//ValidateNodeByTime is used for validate time for each node with current system time(Expiry validation)
func ValidateNodeByTime(claims jwt.MapClaims) int {
	trustedTimeFlag := 0
	if timeVal, ok := claims["valid_to"].(string); ok {

		reg, err := regexp.Compile("[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+")
		Log.Debugf("ValidateNodeByTime reg %v", reg)
		if err != nil {
			Log.Errorf("Error parsing valid_to time: %v", err)
			return trustedTimeFlag
		}
		newstr := reg.ReplaceAllString(timeVal, "")
		Log.Debugf("ValidateNodeByTime newstr: %s", newstr)
		trustedValidToTime := strings.Replace(timeVal, newstr, "", -1)
		Log.Infof("ValidateNodeByTime trustedValidToTime: %s", trustedValidToTime)

		t := time.Now().UTC()
		timeDiff := strings.Compare(trustedValidToTime, t.Format(time.RFC3339))
		Log.Infof("ValidateNodeByTime - ValidTo - %s |  current - %s | Diff - %d", trustedValidToTime, timeVal, timeDiff)
		if timeDiff >= 0 {
			Log.Infof("ValidateNodeByTime -timeDiff: %d ", timeDiff)
			trustedTimeFlag = 1
			Log.Infof("ValidateNodeByTime -trustedTimeFlag: %d", trustedTimeFlag)
		} else {
			Log.Infof("ValidateNodeByTime - Node outside expiry time - ValidTo - %s |  current - %s", timeVal, t)
		}

	}

	return trustedTimeFlag
}
