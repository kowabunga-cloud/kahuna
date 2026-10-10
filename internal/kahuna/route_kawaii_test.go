/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"net/http"
	"testing"

	"github.com/kowabunga-cloud/kahuna/internal/sdk"
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

func TestKawaiiFirewallFromModel(t *testing.T) {
	fw, err := kawaiiFirewallFromModel(sdk.KawaiiFirewall{
		Ingress: []sdk.KawaiiFirewallIngressRule{{Ports: "22"}},
		Egress:  []sdk.KawaiiFirewallEgressRule{{Ports: "443", Protocol: "udp"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if fw.EgressPolicy != KawaiiFirewallPolicyAccept {
		t.Errorf("default egress policy = %q", fw.EgressPolicy)
	}
	if len(fw.Ingress) != 1 || fw.Ingress[0].Source != KawaiiFirewallWildcardNetwork || fw.Ingress[0].Protocol != KawaiiFirewallProtocolTCP {
		t.Errorf("ingress defaults not applied: %+v", fw.Ingress)
	}
	if len(fw.Egress) != 1 || fw.Egress[0].Destination != KawaiiFirewallWildcardNetwork || fw.Egress[0].Protocol != "udp" {
		t.Errorf("egress defaults not applied: %+v", fw.Egress)
	}

	if _, err := kawaiiFirewallFromModel(sdk.KawaiiFirewall{Egress: []sdk.KawaiiFirewallEgressRule{{Ports: "not-a-port"}}}); err == nil {
		t.Error("invalid egress ports should be rejected")
	}
	if _, err := kawaiiDNatRulesFromModel([]sdk.KawaiiDNatRule{{Ports: "not-a-port"}}); err == nil {
		t.Error("invalid DNAT ports should be rejected")
	}
}
