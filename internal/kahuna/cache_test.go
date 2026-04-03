/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"testing"
)

func TestCacheKey(t *testing.T) {
	kc := &KowabungaCache{}
	cases := []struct {
		ns, key, expected string
	}{
		{"users", "abc123", "users/abc123"},
		{"", "key", "/key"},
		{"ns", "", "ns/"},
		{"", "", "/"},
	}
	for _, c := range cases {
		if got := kc.key(c.ns, c.key); got != c.expected {
			t.Errorf("key(%q, %q) = %q, want %q", c.ns, c.key, got, c.expected)
		}
	}
}

func TestCacheDisabledSetIsNoop(t *testing.T) {
	kc := &KowabungaCache{enabled: false}
	// Should not panic when disabled
	kc.Set("ns", "key", "value")
}

func TestCacheDisabledGetReturnsError(t *testing.T) {
	kc := &KowabungaCache{enabled: false}
	var result string
	err := kc.Get("ns", "key", &result)
	if err == nil {
		t.Error("Get on disabled cache should return an error")
	}
	if err.Error() != CacheErrDisabled {
		t.Errorf("expected %q, got %q", CacheErrDisabled, err.Error())
	}
}

func TestCacheDisabledDeleteReturnsError(t *testing.T) {
	kc := &KowabungaCache{enabled: false}
	err := kc.Delete("ns", "key")
	if err == nil {
		t.Error("Delete on disabled cache should return an error")
	}
	if err.Error() != CacheErrDisabled {
		t.Errorf("expected %q, got %q", CacheErrDisabled, err.Error())
	}
}

func TestCacheInitDisabled(t *testing.T) {
	kc := &KowabungaCache{}
	kc.Init(false, CacheTypeInMemory, 16, 15)
	if kc.enabled {
		t.Error("cache should remain disabled after Init(false, ...)")
	}
}

func TestCacheInitUnsupportedType(t *testing.T) {
	kc := &KowabungaCache{}
	kc.Init(true, "unsupported-type", 16, 15)
	if kc.enabled {
		t.Error("cache should be disabled after Init with unsupported type")
	}
}

func TestCacheInitMemoryEnabled(t *testing.T) {
	kc := &KowabungaCache{}
	kc.Init(true, CacheTypeInMemory, 16, 15)
	if !kc.enabled {
		t.Error("cache should be enabled after Init with memory type")
	}
}

func TestCacheSetAndGet(t *testing.T) {
	kc := &KowabungaCache{}
	kc.Init(true, CacheTypeInMemory, 16, 15)

	type payload struct{ Name string }
	kc.Set("ns", "k1", payload{Name: "hello"})

	var result payload
	err := kc.Get("ns", "k1", &result)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if result.Name != "hello" {
		t.Errorf("expected Name=%q, got %q", "hello", result.Name)
	}
}

func TestCacheDelete(t *testing.T) {
	kc := &KowabungaCache{}
	kc.Init(true, CacheTypeInMemory, 16, 15)

	kc.Set("ns", "k2", "some-value")

	if err := kc.Delete("ns", "k2"); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	var result string
	err := kc.Get("ns", "k2", &result)
	if err == nil {
		t.Error("Get after Delete should return an error (cache miss)")
	}
}
