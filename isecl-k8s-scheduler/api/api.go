/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package api

import (
	"bytes"
	"encoding/json"
	commLog "github.com/intel-secl/intel-secl/v4/pkg/lib/common/log"
	"github.com/intel-secl/k8s-extensions/v4/isecl-k8s-scheduler/algorithm"
	"io/ioutil"
	"net/http"

	schedulerapi "k8s.io/kube-scheduler/extender/v1"
)

var defaultLog = commLog.GetDefaultLogger()

type ResourceStore struct {
	IHubPubKeys map[string][]byte
	TagPrefix   string
}

type FilterHandler struct {
	ResourceStore ResourceStore
}

//FilterHandler is the filter host.
func (f *FilterHandler) Filter(w http.ResponseWriter, r *http.Request) {
	var args schedulerapi.ExtenderArgs

	if r.Body == nil || r.ContentLength == 0 {
		defaultLog.Errorf("Error: Empty request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, _ := ioutil.ReadAll(r.Body)
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	err := dec.Decode(&args)
	if err != nil {
		defaultLog.Errorf("Error marshalling json data: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defaultLog.Infof("Post received at ISecL extended scheduler, ExtenderArgs: %v", args)
	//Create a binding for args passed to the POST api
	result, err := algorithm.FilteredHost(&args, f.ResourceStore.IHubPubKeys, f.ResourceStore.TagPrefix)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		defaultLog.Errorf("Error while serving request %v", err)
		return
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		defaultLog.Errorf("Error while json marshalling of response %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = bytes.NewBuffer(resultBytes).WriteTo(w)
	if err != nil {
		defaultLog.Errorf("Error while writing response %v", err)
	}
	return
}
