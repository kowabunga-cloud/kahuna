/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLimitRequestBody(t *testing.T) {
	var readErr error
	h := limitRequestBody(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, readErr = io.ReadAll(r.Body)
	}))

	small := httptest.NewRequest("POST", "/api/v1/login", bytes.NewReader(make([]byte, 1024)))
	h.ServeHTTP(httptest.NewRecorder(), small)
	if readErr != nil {
		t.Errorf("small body should be readable: %v", readErr)
	}

	huge := httptest.NewRequest("POST", "/api/v1/login", bytes.NewReader(make([]byte, HttpMaxRequestBodyBytes+1)))
	h.ServeHTTP(httptest.NewRecorder(), huge)
	if readErr == nil {
		t.Error("oversized body should fail to be read")
	}
}
