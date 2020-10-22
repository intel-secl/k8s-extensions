/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package crdController

import (
	trust_schema "intel/isecl/k8s-custom-controller/v3/crdSchema/api/hostattribute/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestGetPlCrdDef(t *testing.T) {
	expectPlCrd := CrdDefinition{
		Plural:   "hostattributes",
		Singular: "hostattribute",
		Group:    "crd.isecl.intel.com",
		Kind:     "HostAttributeCrd",
	}
	recvPlCrd := GetHACrdDef()
	if reflect.DeepEqual(expectPlCrd, recvPlCrd) {
		t.Errorf("Expected :%v however Received: %v ", expectPlCrd, recvPlCrd)
	}
	t.Logf("Test GetPLCrd Def success")
}

func TestGetPlObjLabel(t *testing.T) {
	trustObj := trust_schema.Host{
		Hostname:          "Node123",
		Trusted:           true,
		Expiry:            time.Now().AddDate(1, 0, 0),
		ISeclSignedReport: "495270d6242e2c67e24e22bad49dgdah",
		SgxSignedReport:   "495270d6242e2c67e24e22bad49dgdah",
		AssetTag: map[string]string{
			"country.us":   "true",
			"country.uk":   "true",
			"state.ca":     "true",
			"city.seattle": "true",
		},
	}

	node := &corev1.Node{}

	tagConfPath := "../tag-prefix-config/tag_prefix.conf"
	t.Log(os.Getwd())
	recvlabel, recannotate, _ := GetHaObjLabel(trustObj, node, tagConfPath)
	prefix, _ := getPrefixFromConf(tagConfPath)
	if _, ok := recvlabel[prefix+"trusted"]; ok {
		t.Logf("Found in HA label Trusted field")
	} else {
		t.Fatalf("Could not get label trusted from HA Report")
	}
	if _, ok := recvlabel[prefix+"country.us"]; ok {
		t.Logf("Found HA label in AssetTag report")
	} else {
		t.Fatalf("Could not get required label from HA Report")
	}
	if _, ok := recvlabel["TrustTagExpiry"]; ok {
		t.Logf("Found in HA label TrustTagExpiry field")
	} else {
		t.Fatalf("Could not get label TrustTagExpiry from HA Report")
	}
	if _, ok := recannotate["TrustTagSignedReport"]; ok {
		t.Logf("Found in HA annotation TrustTagSignedReport ")
	} else {
		t.Fatalf("Could not get annotation TrustTagSignedReport from HA Report")
	}
	t.Logf("Test getHaObjLabel success")
}
