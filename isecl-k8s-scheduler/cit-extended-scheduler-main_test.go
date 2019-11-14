/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestExtendedScheduler(t *testing.T) {
	fmt.Println("Starting extended scheduler Test...")
	gin.SetMode(gin.TestMode)

	testrouter, srv := SetupRouter()

	//test 404 not found error code for the post operation
	req, err := http.NewRequest("POST", "/filter", nil)
	if err != nil {
		fmt.Println(err)
	}

	resp := httptest.NewRecorder()
	testrouter.ServeHTTP(resp, req)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("Expecting status 404 not found : got : %v", resp.Code)
	}

	//test 200 code to check that the extended scheduler server is up
	req, err = http.NewRequest("GET", "/", nil)
	if err != nil {
		fmt.Println(err)
	}

	resp = httptest.NewRecorder()
	testrouter.ServeHTTP(resp, req)
	if resp.Code != 200 {
		t.Fatalf("Expecting status 200 found : got : %v", resp.Code)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
