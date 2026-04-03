/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"testing"
)

func newTestSubnet(cidr string, reserved, gwPool []*IPRange) *Subnet {
	return &Subnet{
		Resource: NewResource("test-subnet", "", 1),
		CIDR:     cidr,
		Reserved: reserved,
		GwPool:   gwPool,
	}
}

// IPRange.Size

func TestIPRangeSizeValid(t *testing.T) {
	r := &IPRange{First: "10.0.0.1", Last: "10.0.0.10"}
	if got := r.Size(); got != 10 {
		t.Errorf("expected size 10, got %d", got)
	}
}

func TestIPRangeSizeSingle(t *testing.T) {
	r := &IPRange{First: "10.0.0.5", Last: "10.0.0.5"}
	if got := r.Size(); got != 1 {
		t.Errorf("expected size 1, got %d", got)
	}
}

func TestIPRangeSizeInvalidFirst(t *testing.T) {
	r := &IPRange{First: "not-an-ip", Last: "10.0.0.10"}
	if got := r.Size(); got != 0 {
		t.Errorf("expected 0 for invalid first IP, got %d", got)
	}
}

func TestIPRangeSizeInvalidLast(t *testing.T) {
	r := &IPRange{First: "10.0.0.1", Last: "not-an-ip"}
	if got := r.Size(); got != 0 {
		t.Errorf("expected 0 for invalid last IP, got %d", got)
	}
}

// Subnet.Size

func TestSubnetSize24(t *testing.T) {
	s := newTestSubnet("10.0.0.0/24", nil, nil)
	if got := s.Size(); got != 24 {
		t.Errorf("expected 24, got %d", got)
	}
}

func TestSubnetSize16(t *testing.T) {
	s := newTestSubnet("192.168.0.0/16", nil, nil)
	if got := s.Size(); got != 16 {
		t.Errorf("expected 16, got %d", got)
	}
}

func TestSubnetSizeInvalidCIDR(t *testing.T) {
	s := newTestSubnet("not-a-cidr", nil, nil)
	if got := s.Size(); got != 0 {
		t.Errorf("expected 0 for invalid CIDR, got %d", got)
	}
}

// Subnet.IsValid

func TestSubnetIsValidInRange(t *testing.T) {
	s := newTestSubnet("10.0.0.0/24", nil, nil)
	if !s.IsValid("10.0.0.100") {
		t.Error("10.0.0.100 should be valid in 10.0.0.0/24")
	}
}

func TestSubnetIsValidOutOfRange(t *testing.T) {
	s := newTestSubnet("10.0.0.0/24", nil, nil)
	if s.IsValid("10.0.1.1") {
		t.Error("10.0.1.1 should not be valid in 10.0.0.0/24")
	}
}

func TestSubnetIsValidInvalidIP(t *testing.T) {
	s := newTestSubnet("10.0.0.0/24", nil, nil)
	if s.IsValid("not-an-ip") {
		t.Error("invalid IP string should return false")
	}
}

func TestSubnetIsValidInvalidCIDR(t *testing.T) {
	s := newTestSubnet("bad-cidr", nil, nil)
	if s.IsValid("10.0.0.1") {
		t.Error("invalid CIDR should return false")
	}
}

// Subnet.IsInReservedPool

func TestSubnetIsInReservedPoolMatch(t *testing.T) {
	reserved := []*IPRange{{First: "10.0.0.1", Last: "10.0.0.10"}}
	s := newTestSubnet("10.0.0.0/24", reserved, nil)
	if !s.IsInReservedPool("10.0.0.5") {
		t.Error("10.0.0.5 should be in reserved pool")
	}
}

func TestSubnetIsInReservedPoolBoundaryFirst(t *testing.T) {
	reserved := []*IPRange{{First: "10.0.0.1", Last: "10.0.0.10"}}
	s := newTestSubnet("10.0.0.0/24", reserved, nil)
	if !s.IsInReservedPool("10.0.0.1") {
		t.Error("first IP of reserved range should be in pool")
	}
}

func TestSubnetIsInReservedPoolBoundaryLast(t *testing.T) {
	reserved := []*IPRange{{First: "10.0.0.1", Last: "10.0.0.10"}}
	s := newTestSubnet("10.0.0.0/24", reserved, nil)
	if !s.IsInReservedPool("10.0.0.10") {
		t.Error("last IP of reserved range should be in pool")
	}
}

func TestSubnetIsInReservedPoolNoMatch(t *testing.T) {
	reserved := []*IPRange{{First: "10.0.0.1", Last: "10.0.0.10"}}
	s := newTestSubnet("10.0.0.0/24", reserved, nil)
	if s.IsInReservedPool("10.0.0.100") {
		t.Error("10.0.0.100 should not be in reserved pool")
	}
}

func TestSubnetIsInReservedPoolEmpty(t *testing.T) {
	s := newTestSubnet("10.0.0.0/24", nil, nil)
	if s.IsInReservedPool("10.0.0.5") {
		t.Error("should return false when no reserved ranges")
	}
}

// Subnet.IsInGwPool

func TestSubnetIsInGwPoolMatch(t *testing.T) {
	gw := []*IPRange{{First: "10.0.0.250", Last: "10.0.0.254"}}
	s := newTestSubnet("10.0.0.0/24", nil, gw)
	if !s.IsInGwPool("10.0.0.252") {
		t.Error("10.0.0.252 should be in gw pool")
	}
}

func TestSubnetIsInGwPoolNoMatch(t *testing.T) {
	gw := []*IPRange{{First: "10.0.0.250", Last: "10.0.0.254"}}
	s := newTestSubnet("10.0.0.0/24", nil, gw)
	if s.IsInGwPool("10.0.0.100") {
		t.Error("10.0.0.100 should not be in gw pool")
	}
}

func TestSubnetIsInGwPoolEmpty(t *testing.T) {
	s := newTestSubnet("10.0.0.0/24", nil, nil)
	if s.IsInGwPool("10.0.0.252") {
		t.Error("should return false when no gw pool ranges")
	}
}
