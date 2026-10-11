/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func decodeHealthStatus(t *testing.T, rec *httptest.ResponseRecorder) healthStatus {
	t.Helper()
	var s healthStatus
	if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil {
		t.Fatalf("invalid health status body %q: %v", rec.Body.String(), err)
	}
	return s
}

func TestLivenessHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	livenessHandler(rec, httptest.NewRequest(http.MethodGet, HealthRouteLiveness, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("got HTTP %d", rec.Code)
	}
	if s := decodeHealthStatus(t, rec); s.Status != "ok" || s.Version != Version {
		t.Fatalf("unexpected status: %+v", s)
	}
}

func TestReadinessHandlerWithoutDatabase(t *testing.T) {
	if GetDB().Client != nil {
		t.Skip("database is connected")
	}
	rec := httptest.NewRecorder()
	readinessHandler(rec, httptest.NewRequest(http.MethodGet, HealthRouteReadiness, nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("got HTTP %d", rec.Code)
	}
	if s := decodeHealthStatus(t, rec); s.Status != "unavailable" {
		t.Fatalf("unexpected status: %+v", s)
	}
}
