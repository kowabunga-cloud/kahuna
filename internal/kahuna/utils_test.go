/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"encoding/xml"
	"net"
	"testing"
)

// SetDefaultStr

func TestSetDefaultStrNil(t *testing.T) {
	if got := SetDefaultStr(nil); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestSetDefaultStrValue(t *testing.T) {
	s := "hello"
	if got := SetDefaultStr(&s); got != "hello" {
		t.Errorf("expected %q, got %q", "hello", got)
	}
}

// SetFieldStr

func TestSetFieldStrEmpty(t *testing.T) {
	field := "original"
	SetFieldStr(&field, "")
	if field != "original" {
		t.Errorf("field should be unchanged, got %q", field)
	}
}

func TestSetFieldStrNonEmpty(t *testing.T) {
	field := "original"
	SetFieldStr(&field, "updated")
	if field != "updated" {
		t.Errorf("expected %q, got %q", "updated", field)
	}
}

// XmlMarshal / XmlUnmarshal

type testXMLStruct struct {
	XMLName xml.Name `xml:"item"`
	Value   string   `xml:"value"`
}

func TestXmlMarshalUnmarshal(t *testing.T) {
	in := testXMLStruct{Value: "hello"}
	out, err := XmlMarshal(in)
	if err != nil {
		t.Fatalf("XmlMarshal: %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty XML output")
	}

	var decoded testXMLStruct
	if err := XmlUnmarshal(out, &decoded); err != nil {
		t.Fatalf("XmlUnmarshal: %v", err)
	}
	if decoded.Value != in.Value {
		t.Errorf("expected Value=%q, got %q", in.Value, decoded.Value)
	}
}

func TestXmlUnmarshalInvalid(t *testing.T) {
	var v testXMLStruct
	if err := XmlUnmarshal("not xml", &v); err == nil {
		t.Error("expected error for invalid XML")
	}
}

// byteCountIEC

func TestByteCountIECBytes(t *testing.T) {
	if got := byteCountIEC(512); got != "512 B" {
		t.Errorf("expected '512 B', got %q", got)
	}
}

func TestByteCountIECKilobytes(t *testing.T) {
	if got := byteCountIEC(1024); got != "1.0 KiB" {
		t.Errorf("expected '1.0 KiB', got %q", got)
	}
}

func TestByteCountIECMegabytes(t *testing.T) {
	if got := byteCountIEC(1024 * 1024); got != "1.0 MiB" {
		t.Errorf("expected '1.0 MiB', got %q", got)
	}
}

func TestByteCountIECGigabytes(t *testing.T) {
	if got := byteCountIEC(1024 * 1024 * 1024); got != "1.0 GiB" {
		t.Errorf("expected '1.0 GiB', got %q", got)
	}
}

// VerifyDomain

func TestVerifyDomainValid(t *testing.T) {
	cases := []string{
		"example.com",
		"sub.example.com",
		"my-domain.org",
		"foo.bar.baz",
	}
	for _, c := range cases {
		if !VerifyDomain(c) {
			t.Errorf("expected %q to be a valid domain", c)
		}
	}
}

func TestVerifyDomainInvalid(t *testing.T) {
	cases := []string{
		"",
		"nodot",
		"has space.com",
	}
	for _, c := range cases {
		if VerifyDomain(c) {
			t.Errorf("expected %q to be an invalid domain", c)
		}
	}
}

// VerifyHostname

func TestVerifyHostnameValid(t *testing.T) {
	cases := []string{
		"myhost",
		"my-host",
		"host123",
		"123host",
		"a",
	}
	for _, c := range cases {
		if !VerifyHostname(c) {
			t.Errorf("expected %q to be a valid hostname", c)
		}
	}
}

func TestVerifyHostnameInvalid(t *testing.T) {
	cases := []string{
		"",
		"has space",
		"host!name",
	}
	for _, c := range cases {
		if VerifyHostname(c) {
			t.Errorf("expected %q to be an invalid hostname", c)
		}
	}
}

// VerifyEmail

func TestVerifyEmailValid(t *testing.T) {
	if err := VerifyEmail("user@example.com"); err != nil {
		t.Errorf("expected valid email, got error: %v", err)
	}
}

func TestVerifyEmailInvalid(t *testing.T) {
	cases := []string{
		"notanemail",
		"missing@",
		"@nodomain",
	}
	for _, c := range cases {
		if err := VerifyEmail(c); err == nil {
			t.Errorf("expected error for invalid email %q", c)
		}
	}
}

// HasChildRef / AddChildRef / RemoveChildRef / HasChildRefs

func TestHasChildRef(t *testing.T) {
	children := []string{"a", "b", "c"}
	if !HasChildRef(&children, "b") {
		t.Error("expected 'b' to be present")
	}
	if HasChildRef(&children, "z") {
		t.Error("expected 'z' to be absent")
	}
}

func TestAddChildRefNew(t *testing.T) {
	children := []string{"a"}
	AddChildRef(&children, "b")
	if len(children) != 2 || children[1] != "b" {
		t.Errorf("expected ['a','b'], got %v", children)
	}
}

func TestAddChildRefDuplicate(t *testing.T) {
	children := []string{"a", "b"}
	AddChildRef(&children, "b")
	if len(children) != 2 {
		t.Errorf("expected no duplicate, got %v", children)
	}
}

func TestRemoveChildRef(t *testing.T) {
	children := []string{"a", "b", "c"}
	RemoveChildRef(&children, "b")
	if len(children) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(children))
	}
	if HasChildRef(&children, "b") {
		t.Error("'b' should have been removed")
	}
}

func TestRemoveChildRefNotPresent(t *testing.T) {
	children := []string{"a", "b"}
	RemoveChildRef(&children, "z")
	if len(children) != 2 {
		t.Errorf("expected slice unchanged, got %v", children)
	}
}

func TestHasChildRefs(t *testing.T) {
	empty := []string{}
	nonempty := []string{"x"}

	if HasChildRefs(empty) {
		t.Error("expected false for all-empty slices")
	}
	if !HasChildRefs(empty, nonempty) {
		t.Error("expected true when at least one slice is non-empty")
	}
	if !HasChildRefs(nonempty) {
		t.Error("expected true for non-empty slice")
	}
}

// bytesToGB

func TestBytesToGB(t *testing.T) {
	cases := []struct {
		input    int64
		expected int64
	}{
		{0, 0},
		{1024 * 1024 * 1024, 1},
		{10 * 1024 * 1024 * 1024, 10},
		{512 * 1024 * 1024, 0}, // less than 1 GiB truncates to 0
	}
	for _, c := range cases {
		if got := bytesToGB(c.input); got != c.expected {
			t.Errorf("bytesToGB(%d) = %d, want %d", c.input, got, c.expected)
		}
	}
}

// FindNextBitIp

func TestFindNextBitIpValid(t *testing.T) {
	_, ipnet, _ := net.ParseCIDR("10.0.0.0/24")
	ip, err := FindNextBitIp(*ipnet, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip.String() != "10.0.0.1" {
		t.Errorf("expected 10.0.0.1, got %s", ip.String())
	}
}

func TestFindNextBitIpBeyondSubnet(t *testing.T) {
	_, ipnet, _ := net.ParseCIDR("10.0.0.0/30")
	// /30 has 4 addresses (0-3); incrementing by 5 goes outside
	_, err := FindNextBitIp(*ipnet, 5)
	if err == nil {
		t.Error("expected error for IP outside subnet")
	}
}

// IsValidPortListExpression

func TestIsValidPortListExpressionSingle(t *testing.T) {
	if err := IsValidPortListExpression("80"); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestIsValidPortListExpressionMultiple(t *testing.T) {
	if err := IsValidPortListExpression("80,443,8080"); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestIsValidPortListExpressionRange(t *testing.T) {
	if err := IsValidPortListExpression("1024-2048"); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestIsValidPortListExpressionMixed(t *testing.T) {
	if err := IsValidPortListExpression("80,443,1000-2000"); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestIsValidPortListExpressionBelowMin(t *testing.T) {
	if err := IsValidPortListExpression("0"); err == nil {
		t.Error("expected error for port 0 (below min)")
	}
}

func TestIsValidPortListExpressionAboveMax(t *testing.T) {
	if err := IsValidPortListExpression("65536"); err == nil {
		t.Error("expected error for port 65536 (above max)")
	}
}

func TestIsValidPortListExpressionInvalidRange(t *testing.T) {
	// last < first
	if err := IsValidPortListExpression("2000-1000"); err == nil {
		t.Error("expected error for inverted range")
	}
}

func TestIsValidPortListExpressionTooManyRangeParts(t *testing.T) {
	if err := IsValidPortListExpression("100-200-300"); err == nil {
		t.Error("expected error for range with more than 2 elements")
	}
}

func TestIsValidPortListExpressionNonNumeric(t *testing.T) {
	if err := IsValidPortListExpression("abc"); err == nil {
		t.Error("expected error for non-numeric port")
	}
}

func TestIsValidPortListExpressionBoundary(t *testing.T) {
	if err := IsValidPortListExpression("1"); err != nil {
		t.Errorf("port 1 should be valid, got: %v", err)
	}
	if err := IsValidPortListExpression("65535"); err != nil {
		t.Errorf("port 65535 should be valid, got: %v", err)
	}
}
