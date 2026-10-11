/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kowabunga-cloud/kahuna/internal/sdk"
)

func decodeApiError(t *testing.T, rec *httptest.ResponseRecorder) ApiError {
	t.Helper()
	var e ApiError
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
		t.Fatalf("response body is not a JSON ApiError: %v (%q)", err, rec.Body.String())
	}
	return e
}

func TestNewApiError(t *testing.T) {
	e := NewApiError(http.StatusNotFound, errors.New("no such resource"))
	if e.Status != http.StatusNotFound || e.Error != "Not Found" || e.Reason != "no such resource" {
		t.Fatalf("unexpected client error: %+v", e)
	}

	// server-side failure details must not be disclosed
	e = NewApiError(http.StatusInternalServerError, errors.New("db connection string"))
	if e.Status != http.StatusInternalServerError || e.Reason != "" {
		t.Fatalf("unexpected server error: %+v", e)
	}

	// quota errors are meaningful to clients
	e = NewApiError(http.StatusInsufficientStorage, errors.New("quota exceeded"))
	if e.Reason != "quota exceeded" {
		t.Fatalf("unexpected quota error: %+v", e)
	}
}

func TestHttpErrorHelpers(t *testing.T) {
	tests := []struct {
		name string
		fn   func(error) (sdk.ImplResponse, error)
		code int
	}{
		{"BadParams", HttpBadParams, http.StatusBadRequest},
		{"Unauthorized", HttpUnauthorized, http.StatusUnauthorized},
		{"Forbidden", HttpForbidden, http.StatusForbidden},
		{"NotFound", HttpNotFound, http.StatusNotFound},
		{"Conflict", HttpConflict, http.StatusConflict},
		{"ServerError", HttpServerError, http.StatusUnprocessableEntity},
		{"InternalError", HttpInternalError, http.StatusInternalServerError},
		{"Quota", HttpQuota, http.StatusInsufficientStorage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.fn(nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			body, ok := res.Body.(ApiError)
			if res.Code != tt.code || !ok || body.Status != int32(tt.code) {
				t.Fatalf("unexpected response: %+v", res)
			}
		})
	}
}

func TestApiErrorHandler(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		result *sdk.ImplResponse
		code   int
		reason string
	}{
		{"parsing", &sdk.ParsingError{Param: "subnetSize", Err: errors.New("too big")}, nil, http.StatusBadRequest, ""},
		{"required", &sdk.RequiredError{Field: "name"}, nil, http.StatusBadRequest, "field 'name' is required."},
		{"service", errors.New("no such resource"), &sdk.ImplResponse{Code: http.StatusNotFound}, http.StatusNotFound, "no such resource"},
		{"unknown", errors.New("boom"), &sdk.ImplResponse{}, http.StatusInternalServerError, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/project", nil)
			ApiErrorHandler(rec, req, tt.err, tt.result)
			if rec.Code != tt.code {
				t.Fatalf("got HTTP %d, want %d", rec.Code, tt.code)
			}
			e := decodeApiError(t, rec)
			if e.Status != int32(tt.code) {
				t.Fatalf("unexpected status in body: %+v", e)
			}
			if tt.reason != "" && e.Reason != tt.reason {
				t.Fatalf("got reason %q, want %q", e.Reason, tt.reason)
			}
		})
	}
}

func TestWriteApiError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteApiError(rec, http.StatusForbidden, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("got HTTP %d", rec.Code)
	}
	if e := decodeApiError(t, rec); e.Status != http.StatusForbidden || e.Error != "Forbidden" {
		t.Fatalf("unexpected body: %+v", e)
	}
}
