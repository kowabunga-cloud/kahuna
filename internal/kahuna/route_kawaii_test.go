/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"net/http"
	"testing"
)

func TestKawaiiIpSecScopedToKawaii(t *testing.T) {
	GetCache().Init(true, CacheTypeInMemory, 16, 15)
	ipsec := KawaiiIPsec{
		Resource:     NewResource("victim", "", MongoCollectionKawaiiIPsecSchemaVersion),
		KawaiiID:     "cccccccccccccccccccccccc",
		PreSharedKey: "secret",
	}
	GetCache().Set(MongoCollectionIPsecName, ipsec.String(), &ipsec)

	s := &KawaiiService{}
	ownKawaii := "bbbbbbbbbbbbbbbbbbbbbbbb"

	if res, _ := s.ReadKawaiiIpSec(t.Context(), ownKawaii, ipsec.String()); res.Code != http.StatusNotFound {
		t.Errorf("reading another Kawaii's IPsec returned %d, want %d", res.Code, http.StatusNotFound)
	}
	if res, _ := s.DeleteKawaiiIpSec(t.Context(), ownKawaii, ipsec.String()); res.Code != http.StatusNotFound {
		t.Errorf("deleting another Kawaii's IPsec returned %d, want %d", res.Code, http.StatusNotFound)
	}
	if res, _ := s.ReadKawaiiIpSec(t.Context(), ipsec.KawaiiID, ipsec.String()); res.Code != http.StatusOK {
		t.Errorf("reading own IPsec returned %d, want %d", res.Code, http.StatusOK)
	}
}
