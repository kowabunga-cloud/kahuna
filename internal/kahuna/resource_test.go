/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"testing"
	"time"
)

// NewResource

func TestNewResourceFields(t *testing.T) {
	before := time.Now()
	r := NewResource("myname", "mydesc", 2)
	after := time.Now()

	if r.Name != "myname" {
		t.Errorf("expected Name=%q, got %q", "myname", r.Name)
	}
	if r.Description != "mydesc" {
		t.Errorf("expected Description=%q, got %q", "mydesc", r.Description)
	}
	if r.SchemaVersion != 2 {
		t.Errorf("expected SchemaVersion=2, got %d", r.SchemaVersion)
	}
	if r.CreatedAt.Before(before) || r.CreatedAt.After(after) {
		t.Errorf("CreatedAt %v not in expected range [%v, %v]", r.CreatedAt, before, after)
	}
	if !r.CreatedAt.Equal(r.UpdatedAt) {
		t.Errorf("expected CreatedAt == UpdatedAt on creation")
	}
}

func TestNewResourceIDNonZero(t *testing.T) {
	r := NewResource("x", "", 1)
	if r.ID.IsZero() {
		t.Error("expected non-zero ID")
	}
}

func TestNewResourceString(t *testing.T) {
	r := NewResource("x", "", 1)
	hex := r.String()
	if len(hex) != 24 {
		t.Errorf("expected 24-char hex ID, got %q (len=%d)", hex, len(hex))
	}
}

// Resource.Updated

func TestResourceUpdated(t *testing.T) {
	r := NewResource("x", "", 1)
	original := r.UpdatedAt
	time.Sleep(time.Millisecond)
	r.Updated()
	if !r.UpdatedAt.After(original) {
		t.Error("UpdatedAt should advance after Updated()")
	}
}

// Resource.UpdateResourceDefaults

func TestUpdateResourceDefaultsBoth(t *testing.T) {
	r := NewResource("old", "olddesc", 1)
	r.UpdateResourceDefaults("new", "newdesc")
	if r.Name != "new" {
		t.Errorf("expected Name=%q, got %q", "new", r.Name)
	}
	if r.Description != "newdesc" {
		t.Errorf("expected Description=%q, got %q", "newdesc", r.Description)
	}
}

func TestUpdateResourceDefaultsEmpty(t *testing.T) {
	r := NewResource("original", "origdesc", 1)
	r.UpdateResourceDefaults("", "")
	if r.Name != "original" {
		t.Errorf("empty name should not overwrite; got %q", r.Name)
	}
	if r.Description != "origdesc" {
		t.Errorf("empty desc should not overwrite; got %q", r.Description)
	}
}

func TestUpdateResourceDefaultsNameOnly(t *testing.T) {
	r := NewResource("old", "keepme", 1)
	r.UpdateResourceDefaults("new", "")
	if r.Name != "new" {
		t.Errorf("expected Name=%q, got %q", "new", r.Name)
	}
	if r.Description != "keepme" {
		t.Errorf("description should be unchanged, got %q", r.Description)
	}
}

// NewResourceCost

func TestNewResourceCostCurrencyDefault(t *testing.T) {
	c := NewResourceCost(9.99, "")
	if c.Currency != CostCurrencyDefault {
		t.Errorf("expected default currency %q, got %q", CostCurrencyDefault, c.Currency)
	}
	if c.Price != 9.99 {
		t.Errorf("expected price 9.99, got %v", c.Price)
	}
}

func TestNewResourceCostCurrencyCustom(t *testing.T) {
	c := NewResourceCost(1.0, "USD")
	if c.Currency != "USD" {
		t.Errorf("expected currency %q, got %q", "USD", c.Currency)
	}
}

func TestResourceCostModel(t *testing.T) {
	c := NewResourceCost(3.14, "GBP")
	m := c.Model()
	if m.Price != 3.14 {
		t.Errorf("expected Price=3.14, got %v", m.Price)
	}
	if m.Currency != "GBP" {
		t.Errorf("expected Currency=%q, got %q", "GBP", m.Currency)
	}
}

// NewResourceMetadata

func TestNewResourceMetadata(t *testing.T) {
	m := NewResourceMetadata("env", "prod")
	if m.Key != "env" {
		t.Errorf("expected Key=%q, got %q", "env", m.Key)
	}
	if m.Value != "prod" {
		t.Errorf("expected Value=%q, got %q", "prod", m.Value)
	}
}

func TestResourceMetadataModel(t *testing.T) {
	m := NewResourceMetadata("region", "us-east-1")
	sdk := m.Model()
	if sdk.Key != "region" {
		t.Errorf("expected Key=%q, got %q", "region", sdk.Key)
	}
	if sdk.Value != "us-east-1" {
		t.Errorf("expected Value=%q, got %q", "us-east-1", sdk.Value)
	}
}
