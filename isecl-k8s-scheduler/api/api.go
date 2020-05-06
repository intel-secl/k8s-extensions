/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package api

import (
	"bytes"
	"encoding/json"
	"intel/isecl/k8s-extended-scheduler/v2/algorithm"
	"intel/isecl/k8s-extended-scheduler/v2/util"
	"io/ioutil"
	"net/http"

	schedulerapi "k8s.io/kube-scheduler/extender/v1"
)

var Confpath string
var Log = util.GetLogger()

//FilterHandler is the filter host.
func FilterHandler(w http.ResponseWriter, r *http.Request) {
	var args schedulerapi.ExtenderArgs
	data, _ := ioutil.ReadAll(r.Body)
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	err := dec.Decode(&args)
	if err != nil {
		Log.Errorf("Error marshaling json data: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	Log.Infof("Post received at ISecL extended scheduler, ExtenderArgs: %v", args)
	//Create a binding for args passed to the POST api
	prefixString := Confpath
	result, err := algorithm.FilteredHost(&args, prefixString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Log.Errorf("Error while serving request %v", err)
		return
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Log.Errorf("Error while json marshalling of response %v", err)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	bytes.NewBuffer(resultBytes).WriteTo(w)
	return
}
