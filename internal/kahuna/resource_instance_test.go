/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"testing"
)

func TestDiskLetterForIndexSingleLetters(t *testing.T) {
	cases := []struct {
		index    int
		expected string
	}{
		{0, "a"},
		{1, "b"},
		{25, "z"},
	}
	for _, c := range cases {
		if got := diskLetterForIndex(c.index); got != c.expected {
			t.Errorf("diskLetterForIndex(%d) = %q, want %q", c.index, got, c.expected)
		}
	}
}

func TestDiskLetterForIndexDoubleLetters(t *testing.T) {
	cases := []struct {
		index    int
		expected string
	}{
		{26, "aa"},
		{27, "ab"},
		{51, "az"},
		{52, "ba"},
		{77, "bz"},
		{78, "ca"},
	}
	for _, c := range cases {
		if got := diskLetterForIndex(c.index); got != c.expected {
			t.Errorf("diskLetterForIndex(%d) = %q, want %q", c.index, got, c.expected)
		}
	}
}

func TestDiskLetterForIndexMonotonic(t *testing.T) {
	// Verify that consecutive indices produce distinct, ordered strings
	prev := diskLetterForIndex(0)
	for i := 1; i <= 100; i++ {
		curr := diskLetterForIndex(i)
		if curr == prev {
			t.Errorf("diskLetterForIndex(%d) == diskLetterForIndex(%d) = %q", i, i-1, curr)
		}
		prev = curr
	}
}

func TestInstanceRejectsForeignProjectResources(t *testing.T) {
	i := &Instance{Resource: NewResource("vm", "", MongoCollectionInstanceSchemaVersion), ProjectID: "bbbbbbbbbbbbbbbbbbbbbbbb"}

	v := &Volume{Resource: NewResource("disk", "", MongoCollectionVolumeSchemaVersion), ProjectID: "cccccccccccccccccccccccc"}
	if err := i.verifyVolumeOwnership(v); err == nil {
		t.Error("volume from another project should be rejected")
	}

	a := &Adapter{Resource: NewResource("nic", "", MongoCollectionAdapterSchemaVersion)}
	s := &Subnet{Resource: NewResource("subnet", "", MongoCollectionSubnetSchemaVersion), ProjectID: "cccccccccccccccccccccccc"}
	if err := i.verifyAdapterOwnership(a, s); err == nil {
		t.Error("adapter on another project's subnet should be rejected")
	}
}
