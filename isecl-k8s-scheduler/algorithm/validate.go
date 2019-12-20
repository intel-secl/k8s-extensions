/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package algorithm

import (
	jwt "github.com/dgrijalva/jwt-go"
	"k8s.io/api/core/v1"
	"regexp"
	"strconv"
	"strings"
	"time"
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
func ValidatePodWithAnnotation(nodeData []v1.NodeSelectorRequirement, claims jwt.MapClaims, trustprefix string) bool {

	if !keyExists(claims, ahreport){
		Log.Errorf("ValidatePodWithAnnotation - Asset Tags not found for node.")
		return false
	}

	assetClaims := claims[ahreport].(map[string]interface{})
        Log.Infoln("ValidatePodWithAnnotation - Validating node %v claims %v", nodeData, assetClaims)


	for _, val := range nodeData {
		//if val is trusted, it can be directly found in claims
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
							Log.Infoln("ValidatePodWithAnnotation - Trust Check - Mismatch in %v field. Actual: %v | In Signature: %v ", val.Key, nodeVal, sigVal)
							return false
						}
					} else {
						if nodeVal == sigVal {
							continue
						} else {
							 Log.Infoln("ValidatePodWithAnnotation - Trust Check - Mismatch in %v field. Actual: %v | In Signature: %v ", val.Key, nodeVal, sigVal)
							return false
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
						}else{
							Log.Infoln("ValidatePodWithAnnotation - Geo Asset Tags - Mismatch in %v field. Actual: %v | In Signature: %v ", geoKey, match, newVal)
						}
					}
					if flag {
						continue
					} else {
						return false
					}
				}

			}
		}
	}
	return true
}

//ValidateNodeByTime is used for validate time for each node with current system time(Expiry validation)
func ValidateNodeByTime(claims jwt.MapClaims) int {
	trustedTimeFlag := 0
	if timeVal, ok := claims["valid_to"].(string); ok {
		reg, err := regexp.Compile("[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+")
		Log.Infoln("ValidateNodeByTime reg %v err %v",reg,err)
		if err != nil {
			Log.Errorf("Error parsing valid_to time: %v",err)
			return trustedTimeFlag
		}
		newstr := reg.ReplaceAllString(timeVal, "")
		Log.Infoln("ValidateNodeByTime newstr %v ",newstr)
		trustedValidToTime := strings.Replace(timeVal, newstr, "", -1)
		Log.Infoln("ValidateNodeByTime trustedValidToTime %v ",trustedValidToTime)

		t := time.Now().UTC()
		timeDiff := strings.Compare(trustedValidToTime, t.Format(time.RFC3339))
		Log.Infoln("ValidateNodeByTime - ValidTo - %s |  current - %s | Diff - %s", trustedValidToTime, timeVal, timeDiff)
		if timeDiff >= 0 {
			Log.Infoln("ValidateNodeByTime -timeDiff ", timeDiff)
			trustedTimeFlag = 1
			Log.Infoln("ValidateNodeByTime -trustedTimeFlag ", trustedTimeFlag)
		}else{
			Log.Infof("ValidateNodeByTime - Node outside expiry time - ValidTo - %s |  current - %s", timeVal, t)
		}

	}

	return trustedTimeFlag
}
