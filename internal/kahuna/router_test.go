/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestTokenAuthCache(t *testing.T) {
	c := newTokenAuthCache()
	c.set("key1", cachedTokenAuth{userId: testAuthzUserId, projectAdmin: true})

	entry, found := c.get("key1")
	if !found || entry.userId != testAuthzUserId || !entry.projectAdmin {
		t.Fatalf("expected cached entry for key1, got %+v (found=%v)", entry, found)
	}
	if _, found := c.get("key2"); found {
		t.Error("unknown key should not be found")
	}
	for k := range c.entries {
		if k == "key1" {
			t.Error("plain-text API key should not be used as cache key")
		}
	}

	c.flush()
	if _, found := c.get("key1"); found {
		t.Error("entry should be gone after flush")
	}
}

func TestTokenAuthCacheExpiry(t *testing.T) {
	c := newTokenAuthCache()
	c.set("key1", cachedTokenAuth{userId: testAuthzUserId})
	c.entries[apiKeyDigest("key1")] = cachedTokenAuth{userId: testAuthzUserId, expiresAt: time.Now().Add(-time.Second)}
	if _, found := c.get("key1"); found {
		t.Error("expired entry should not be returned")
	}
}

func TestTokenAuthCacheBounded(t *testing.T) {
	c := newTokenAuthCache()
	for i := range tokenAuthCacheMaxEntries + 10 {
		c.set(fmt.Sprintf("key%d", i), cachedTokenAuth{userId: testAuthzUserId})
	}
	if len(c.entries) > tokenAuthCacheMaxEntries {
		t.Errorf("cache grew to %d entries, max is %d", len(c.entries), tokenAuthCacheMaxEntries)
	}
	if _, found := c.get(fmt.Sprintf("key%d", tokenAuthCacheMaxEntries+9)); !found {
		t.Error("most recently set entry should be cached")
	}
}

func TestSecretTokenMatches(t *testing.T) {
	cases := []struct {
		provided, expected string
		match              bool
	}{
		{"", "", false},
		{"abc", "", false},
		{"", "abc", false},
		{"abd", "abc", false},
		{"abc", "abc", true},
	}
	for _, c := range cases {
		if got := secretTokenMatches(c.provided, c.expected); got != c.match {
			t.Errorf("secretTokenMatches(%q, %q) = %v, want %v", c.provided, c.expected, got, c.match)
		}
	}
}

func TestUserTokenHandlersRejectEmptyToken(t *testing.T) {
	GetCache().Init(true, CacheTypeInMemory, 16, 15)
	u := User{Resource: NewResource("victim", "", MongoCollectionUserSchemaVersion)}
	GetCache().Set(MongoCollectionUserName, u.String(), &u)

	handlers := map[string]func(http.ResponseWriter, *http.Request){
		"/confirm":               userRegistrationHandler,
		"/confirmForgotPassword": userPasswordRenewalHandler,
	}
	for route, handler := range handlers {
		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest("GET", route+"?user="+u.String()+"&token=", nil))
		if w.Code != http.StatusForbidden {
			t.Errorf("%s with empty token returned %d, want %d", route, w.Code, http.StatusForbidden)
		}
	}
}
