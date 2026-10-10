/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestFindDanglingRefs(t *testing.T) {
	missing := map[string]error{
		"gone":    mongo.ErrNoDocuments,
		"badhex":  bson.ErrInvalidHex,
		"timeout": context.DeadlineExceeded,
	}
	find := func(id string) error { return missing[id] }

	dangling, err := findDanglingRefs([]string{"ok", "gone", "badhex"}, find, "Project p", "instance")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dangling) != 2 || dangling[0] != "gone" || dangling[1] != "badhex" {
		t.Errorf("dangling = %v, want [gone badhex]", dangling)
	}

	// transient errors must abort rather than drop valid references
	if _, err := findDanglingRefs([]string{"gone", "timeout"}, find, "Project p", "instance"); err == nil {
		t.Error("lookup errors other than not-found should abort")
	}
}
