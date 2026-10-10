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
