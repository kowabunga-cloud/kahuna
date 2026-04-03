/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"testing"
)

func TestQuotaToStringUnlimited(t *testing.T) {
	if got := quotaToString(0, false); got != "Unlimited" {
		t.Errorf("expected 'Unlimited', got %q", got)
	}
}

func TestQuotaToStringUnlimitedSize(t *testing.T) {
	if got := quotaToString(0, true); got != "Unlimited" {
		t.Errorf("expected 'Unlimited', got %q", got)
	}
}

func TestQuotaToStringCount(t *testing.T) {
	if got := quotaToString(5, false); got != "5" {
		t.Errorf("expected '5', got %q", got)
	}
}

func TestQuotaToStringSizeNonZero(t *testing.T) {
	// val=1073741824 bytes = 1 GiB; HumanByteSize should produce a non-empty string
	got := quotaToString(1073741824, true)
	if got == "" || got == "Unlimited" {
		t.Errorf("expected a human byte size string, got %q", got)
	}
}
