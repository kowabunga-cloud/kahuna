/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"context"
	"fmt"

	"github.com/kowabunga-cloud/kahuna/internal/sdk"
)

func NewKawaiiRouter() sdk.Router {
	return sdk.NewKawaiiAPIController(&KawaiiService{}, sdk.WithKawaiiAPIErrorHandler(ApiErrorHandler))
}

type KawaiiService struct{}

func (s *KawaiiService) DeleteKawaii(ctx context.Context, kawaiiId string) (sdk.ImplResponse, error) {
	// ensure Kawaii exists
	k, err := FindKawaiiByID(kawaiiId)
	if err != nil {
		return HttpNotFound(err)
	}

	if k.HasChildren() {
		return HttpConflict(nil)
	}
	// remove Kawaii
	err = k.Delete()
	if err != nil {
		return HttpServerError(err)
	}

	return HttpOK(nil)
}

func (s *KawaiiService) ListKawaiis(ctx context.Context) (sdk.ImplResponse, error) {
	return HttpOK(FindResourceIDs(MongoCollectionKawaiiName))
}

func (s *KawaiiService) ReadKawaii(ctx context.Context, kawaiiId string) (sdk.ImplResponse, error) {
	k, err := FindKawaiiByID(kawaiiId)
	if err != nil {
		return HttpNotFound(err)
	}

	payload := k.Model()
	LogHttpResponse(payload)
	return HttpOK(payload)
}

func (s *KawaiiService) UpdateKawaii(ctx context.Context, kawaiiId string, kawaii sdk.KawaiiUpdate) (sdk.ImplResponse, error) {
	LogHttpRequest(RA("kawaiiId", kawaiiId), RA("kawaii", kawaii))

	// Get our kawaii
	gw, err := FindKawaiiByID(kawaiiId)
	if err != nil {
		return HttpNotFound(err)
	}

	// converts kawaii from model to object, unspecified settings are left unchanged
	fw := gw.Firewall
	if kawaii.Firewall != nil {
		fw, err = kawaiiFirewallFromModel(*kawaii.Firewall)
		if err != nil {
			return HttpBadParams(err)
		}
	}
	natRules := gw.DNatRules
	if kawaii.Dnat != nil {
		natRules, err = kawaiiDNatRulesFromModel(kawaii.Dnat)
		if err != nil {
			return HttpBadParams(err)
		}
	}

	// update Kawaii
	err = gw.Update(kawaii.Description, fw, natRules)
	if err != nil {
		return HttpServerError(err)
	}

	payload := gw.Model()
	LogHttpResponse(payload)
	return HttpOK(payload)
}

// IPSEC Connection
func (s *KawaiiService) CreateKawaiiIpSec(ctx context.Context, kawaiiId string, ipsec sdk.KawaiiIpSec) (sdk.ImplResponse, error) {
	LogHttpRequest(RA("kawaiiId", kawaiiId), RA("ipSecConnection", ipsec))

	// Ensure kawaii exists
	_, err := FindKawaiiByID(kawaiiId)
	if err != nil {
		return HttpNotFound(err)
	}
	existingIpsecs, err := FindIPsecByKawaii(kawaiiId)
	if err != nil {
		return HttpNotFound(err)
	}
	for _, existingIpsec := range existingIpsecs {
		if existingIpsec.Name == ipsec.Name || existingIpsec.RemotePeer == ipsec.RemoteIp {
			return HttpConflict(fmt.Errorf("IP Sec connection (name and remotePeer) must be unique : %s/%s", existingIpsec.Name, existingIpsec.RemotePeer))
		}
	}
	fw := KawaiiFirewall{Ingress: []KawaiiFirewallIngressRule{}}
	if ipsec.Firewall != nil {
		fw.Ingress, err = kawaiiIngressRulesFromModel(ipsec.Firewall.Ingress)
		if err != nil {
			return HttpBadParams(err)
		}
	}
	ipsecConn, err := NewKawaiiIPsecConnection(kawaiiId, ipsec.Name, ipsec.Description, ipsec.RemoteIp,
		ipsec.RemoteSubnet, ipsec.PreSharedKey,
		ipsec.DpdTimeout, ipsec.DpdTimeoutAction, ipsec.StartAction, ipsec.RekeyTime,
		ipsec.Phase1Lifetime, int64(ipsec.Phase1DhGroupNumber), ipsec.Phase1IntegrityAlgorithm, ipsec.Phase1EncryptionAlgorithm,
		ipsec.Phase2Lifetime, int64(ipsec.Phase2DhGroupNumber), ipsec.Phase2IntegrityAlgorithm, ipsec.Phase2EncryptionAlgorithm,
		fw)
	if err != nil {
		return HttpServerError(err)
	}
	payload := ipsecConn.Model()
	LogHttpResponse(payload)
	return HttpCreated(payload)
}

// kawaiiIngressRulesFromModel validates and converts firewall ingress rules,
// filling in defaults (any source, TCP).
func kawaiiIngressRulesFromModel(rules []sdk.KawaiiFirewallIngressRule) ([]KawaiiFirewallIngressRule, error) {
	res := []KawaiiFirewallIngressRule{}
	for _, in := range rules {
		err := IsValidPortListExpression(in.Ports)
		if err != nil {
			return nil, err
		}

		rule := KawaiiFirewallIngressRule{
			Source:   in.Source,
			Protocol: in.Protocol,
			Ports:    in.Ports,
		}
		if rule.Source == "" {
			rule.Source = KawaiiFirewallWildcardNetwork
		}
		if rule.Protocol == "" {
			rule.Protocol = KawaiiFirewallProtocolTCP
		}
		res = append(res, rule)
	}
	return res, nil
}

// kawaiiFirewallFromModel validates and converts a Kawaii firewall
// definition, filling in defaults (accept egress policy, any source or
// destination, TCP).
func kawaiiFirewallFromModel(model sdk.KawaiiFirewall) (KawaiiFirewall, error) {
	fw := KawaiiFirewall{
		EgressPolicy: model.EgressPolicy,
	}
	if fw.EgressPolicy == "" {
		fw.EgressPolicy = KawaiiFirewallPolicyAccept
	}

	ingress, err := kawaiiIngressRulesFromModel(model.Ingress)
	if err != nil {
		return fw, err
	}
	fw.Ingress = ingress

	for _, out := range model.Egress {
		err := IsValidPortListExpression(out.Ports)
		if err != nil {
			return fw, err
		}

		rule := KawaiiFirewallEgressRule{
			Destination: out.Destination,
			Protocol:    out.Protocol,
			Ports:       out.Ports,
		}
		if rule.Destination == "" {
			rule.Destination = KawaiiFirewallWildcardNetwork
		}
		if rule.Protocol == "" {
			rule.Protocol = KawaiiFirewallProtocolTCP
		}
		fw.Egress = append(fw.Egress, rule)
	}

	return fw, nil
}

// kawaiiDNatRulesFromModel validates and converts destination NAT rules,
// filling in defaults (TCP).
func kawaiiDNatRulesFromModel(rules []sdk.KawaiiDNatRule) ([]KawaiiDNatRule, error) {
	natRules := []KawaiiDNatRule{}
	for _, r := range rules {
		err := IsValidPortListExpression(r.Ports)
		if err != nil {
			return nil, err
		}

		rule := KawaiiDNatRule{
			PrivateIP: r.Destination,
			Protocol:  r.Protocol,
			Ports:     r.Ports,
		}
		if rule.Protocol == "" {
			rule.Protocol = KawaiiFirewallProtocolTCP
		}
		natRules = append(natRules, rule)
	}
	return natRules, nil
}

// findKawaiiIPsec returns the IPsec connection only if it belongs to the given
// Kawaii. Request authorization is granted on the Kawaii ID from the path, so
// an IPsec connection from another Kawaii (and project) must not be reachable.
func findKawaiiIPsec(kawaiiId, ipsecId string) (*KawaiiIPsec, error) {
	ipsec, err := FindIPsecByID(ipsecId)
	if err != nil {
		return nil, err
	}
	if ipsec.KawaiiID != kawaiiId {
		return nil, fmt.Errorf("IPsec connection %s does not belong to Kawaii %s", ipsecId, kawaiiId)
	}
	return ipsec, nil
}

func (s *KawaiiService) DeleteKawaiiIpSec(ctx context.Context, kawaiiId, ipsecId string) (sdk.ImplResponse, error) {
	// ensure IPsec connection exists
	k, err := findKawaiiIPsec(kawaiiId, ipsecId)
	if err != nil {
		return HttpNotFound(err)
	}

	// remove IPsec connection
	err = k.Delete()
	if err != nil {
		return HttpServerError(err)
	}

	return HttpOK(nil)
}

func (s *KawaiiService) ListKawaiiIpSecs(ctx context.Context, kawaiiId string) (sdk.ImplResponse, error) {
	ipsecs, err := FindIPsecByKawaii(kawaiiId)
	if err != nil {
		return HttpNotFound(err)
	}
	payload := []string{}
	for _, k := range ipsecs {
		payload = append(payload, k.String())
	}

	return HttpOK(payload)
}

func (s *KawaiiService) ReadKawaiiIpSec(ctx context.Context, kawaiiId, ipsecId string) (sdk.ImplResponse, error) {
	// Ensure Ip Sec Conn exists
	ipsec, err := findKawaiiIPsec(kawaiiId, ipsecId)
	if err != nil {
		return HttpNotFound(err)
	}
	payload := ipsec.Model()
	LogHttpResponse(payload)
	return HttpOK(payload)
}

func (s *KawaiiService) UpdateKawaiiIpSec(ctx context.Context, kawaiiId, ipsecId string, ipsecSpec sdk.KawaiiIpSecUpdate) (sdk.ImplResponse, error) {
	LogHttpRequest(RA("kawaiiId", kawaiiId), RA("ipsec", ipsecId))

	// Ensure Ip Sec Conn exists
	ipsecConn, err := findKawaiiIPsec(kawaiiId, ipsecId)
	if err != nil {
		return HttpNotFound(err)
	}
	// unspecified firewall is left unchanged
	var fw *KawaiiFirewall
	if ipsecSpec.Firewall != nil {
		ingress, err := kawaiiIngressRulesFromModel(ipsecSpec.Firewall.Ingress)
		if err != nil {
			return HttpBadParams(err)
		}
		fw = &KawaiiFirewall{Ingress: ingress}
	}
	err = ipsecConn.Update(ipsecSpec.Name, ipsecSpec.Description, ipsecSpec.RemoteIp,
		ipsecSpec.RemoteSubnet, ipsecSpec.PreSharedKey,
		ipsecSpec.DpdTimeout, ipsecSpec.DpdTimeoutAction, ipsecSpec.StartAction, ipsecSpec.RekeyTime,
		ipsecSpec.Phase1Lifetime, int64(ipsecSpec.Phase1DhGroupNumber), ipsecSpec.Phase1IntegrityAlgorithm, ipsecSpec.Phase1EncryptionAlgorithm,
		ipsecSpec.Phase2Lifetime, int64(ipsecSpec.Phase2DhGroupNumber), ipsecSpec.Phase2IntegrityAlgorithm, ipsecSpec.Phase2EncryptionAlgorithm,
		fw)
	if err != nil {
		return HttpServerError(err)
	}
	payload := ipsecConn.Model()
	LogHttpResponse(payload)
	return HttpOK(payload)
}
