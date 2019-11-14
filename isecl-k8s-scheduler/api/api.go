/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package api

import (
	"k8s_scheduler_cit_extension-k8s_extended_scheduler/algorithm"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	schedulerapi "k8s.io/kubernetes/plugin/pkg/scheduler/api"
)

var Confpath string

//FilterHandler is the filter host.
func FilterHandler(c *gin.Context) {
	var args schedulerapi.ExtenderArgs
	glog.Infof("Post received at  ISecL extended scheduler: %v", args)
	//Create a binding for args passed to the POST api
	if c.BindJSON(&args) == nil {
		prefixString := Confpath
		result, err := algorithm.FilteredHost(&args, prefixString)
		if err == nil {
			c.JSON(200, result)
		} else {
			c.JSON(500, err)
		}
	}
}
