/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/kowabunga-cloud/common/klog"
)

const (
	HealthRouteLiveness  = "/healthz"
	HealthRouteReadiness = "/readyz"

	healthReadinessTimeout = 2 * time.Second
)

type healthStatus struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func writeHealthStatus(w http.ResponseWriter, code int, status string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(healthStatus{Status: status, Version: Version})
}

// livenessHandler reports the process is up and serving HTTP requests
// (e.g. Kubernetes liveness probe).
func livenessHandler(w http.ResponseWriter, r *http.Request) {
	writeHealthStatus(w, http.StatusOK, "ok")
}

// readinessHandler reports whether requests can be served, i.e. whether the
// database is reachable (e.g. Kubernetes readiness probe).
func readinessHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), healthReadinessTimeout)
	defer cancel()

	err := GetDB().Ping(ctx)
	if err != nil {
		klog.Warningf("readiness check failed: %v", err)
		writeHealthStatus(w, http.StatusServiceUnavailable, "unavailable")
		return
	}
	writeHealthStatus(w, http.StatusOK, "ok")
}
