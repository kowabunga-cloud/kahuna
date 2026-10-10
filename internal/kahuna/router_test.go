/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"net/http/httptest"
	"testing"
)

func TestReqIsAuthenticatedRejectsNonBearerScheme(t *testing.T) {
	for _, h := range []string{"abc", "Bearer", "Basic dXNlcjpwYXNzd29yZA=="} {
		r := httptest.NewRequest("GET", "/api/v1/region", nil)
		r.Header.Set(HttpHeaderAuthorization, h)
		if _, ok := reqIsAuthenticated(r); ok {
			t.Errorf("Authorization %q should not authenticate", h)
		}
	}
}

const (
	testAuthzUserId    = "aaaaaaaaaaaaaaaaaaaaaaaa"
	testAuthzOwnedId   = "bbbbbbbbbbbbbbbbbbbbbbbb"
	testAuthzForeignId = "cccccccccccccccccccccccc"
	testAuthzOtherUser = "dddddddddddddddddddddddd"
)

func authzRequest(t *testing.T, method, target string) bool {
	t.Helper()
	GetCache().Init(true, CacheTypeInMemory, 16, 15)
	GetCache().Set(CacheNsUserResources, userResourcesCacheKey(testAuthzUserId), []string{testAuthzOwnedId})

	r := httptest.NewRequest(method, target, nil)
	r = r.WithContext(ctxSetUserId(r.Context(), testAuthzUserId))
	return reqIsAuthorized(r)
}

func TestReqIsAuthorized(t *testing.T) {
	cases := []struct {
		name     string
		method   string
		target   string
		expected bool
	}{
		{"own user route", "GET", SdkBaseRoute + "/user/" + testAuthzUserId, true},
		{"other user route", "GET", SdkBaseRoute + "/user/" + testAuthzOtherUser, false},
		{"public region listing", "GET", SdkBaseRoute + "/region", true},
		{"owned resource", "GET", SdkBaseRoute + "/instance/" + testAuthzOwnedId, true},
		{"foreign resource", "GET", SdkBaseRoute + "/instance/" + testAuthzForeignId, false},
		{"foreign resource, owned id in query", "DELETE",
			SdkBaseRoute + "/instance/" + testAuthzForeignId + "?x=" + testAuthzOwnedId, false},
		{"foreign resource, user route in query", "DELETE",
			SdkBaseRoute + "/instance/" + testAuthzForeignId + "?x=" + SdkBaseRoute + "/user/" + testAuthzUserId, false},
		{"foreign resource, public route in query", "GET",
			SdkBaseRoute + "/instance/" + testAuthzForeignId + "?x=" + SdkBaseRoute + "/region", false},
		{"user route not at path start", "GET",
			"/foo" + SdkBaseRoute + "/user/" + testAuthzUserId, false},
	}
	for _, c := range cases {
		if got := authzRequest(t, c.method, c.target); got != c.expected {
			t.Errorf("%s: %s %s authorized = %v, want %v", c.name, c.method, c.target, got, c.expected)
		}
	}
}

func TestUserResourcesCacheInvalidation(t *testing.T) {
	GetCache().Init(true, CacheTypeInMemory, 16, 15)
	GetCache().Set(CacheNsUserResources, userResourcesCacheKey(testAuthzUserId), []string{testAuthzOwnedId})

	r := httptest.NewRequest("GET", SdkBaseRoute+"/instance/"+testAuthzOwnedId, nil)
	r = r.WithContext(ctxSetUserId(r.Context(), testAuthzUserId))
	if !reqIsAuthorized(r) {
		t.Fatal("owned resource should be authorized from cached mapping")
	}

	// once invalidated, the stale mapping must no longer be used (and with no
	// DB connection, the user can't be resolved, so access is denied)
	invalidateUserResources()
	if reqIsAuthorized(r) {
		t.Error("stale user resources mapping should not be used after invalidation")
	}
}
