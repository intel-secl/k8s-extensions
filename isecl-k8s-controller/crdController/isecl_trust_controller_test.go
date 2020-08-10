/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package crdController

import (
	trust_schema "intel/isecl/k8s-custom-controller/v2/crdSchema/api/hostattribute/v1beta1"
	"testing"
	corev1 "k8s.io/api/core/v1"
)

func TestGetPLCrdDef(t *testing.T) {
	expecPlCrd := CrdDefinition{
		Plural:   "hostattributescrds",
		Singular: "hostattributecrd",
		Group:    "crd.isecl.intel.com",
		Kind:     "HostAttributeCrd",
	}
	recvPlCrd := GetHACrdDef()
	if expecPlCrd != recvPlCrd {
		t.Fatalf("Changes found in HA CRD Definition ")
		t.Fatalf("Expected :%v however Received: %v ", expecPlCrd, recvPlCrd)
	}
	t.Logf("Test GetPLCrd Def success")
}

func TestGetPlObjLabel(t *testing.T) {
	trustObj := trust_schema.HostList{
		Hostname:     "Node123",
		Trusted:      "true",
		Expiry:       "12-23-45T123.91.12",
		SignedReport: "495270d6242e2c67e24e22bad49dgdah",
		Assettag: map[string]string{
			"country.us":  "true",
			"country.uk":  "true",
			"state.ca":    "true",
			"city.seatle": "true",
		},
	}
	node := &corev1.Node{}
	path := "/opt/isecl-k8s-extensions/bin/tag_prefix.conf"
	recvlabel, recannotate := GetHaObjLabel(trustObj, node, path)
	if _, ok := recvlabel[getPrefixFromConf(path)+"trusted"]; ok {
		t.Logf("Found in HA label Trusted field")
	} else {
		t.Fatalf("Could not get label trusted from HA Report")
	}
	if _, ok := recvlabel["country.us"]; ok {
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
