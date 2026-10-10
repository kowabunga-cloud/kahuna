/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import "testing"

func TestValidateKwarantineGroups(t *testing.T) {
	GetCache().Init(true, CacheTypeInMemory, 16, 15)
	own := Kwarantine{Resource: NewResource("own", "", MongoCollectionKwarantineSchemaVersion), ProjectID: "bbbbbbbbbbbbbbbbbbbbbbbb"}
	foreign := Kwarantine{Resource: NewResource("foreign", "", MongoCollectionKwarantineSchemaVersion), ProjectID: "cccccccccccccccccccccccc"}
	GetCache().Set(MongoCollectionKwarantineName, own.String(), &own)
	GetCache().Set(MongoCollectionKwarantineName, foreign.String(), &foreign)

	if err := ValidateKwarantineGroups(own.ProjectID, nil); err != nil {
		t.Errorf("no groups should be valid: %v", err)
	}
	if err := ValidateKwarantineGroups(own.ProjectID, []string{own.String()}); err != nil {
		t.Errorf("own group should be valid: %v", err)
	}
	if err := ValidateKwarantineGroups(own.ProjectID, []string{own.String(), foreign.String()}); err == nil {
		t.Error("another project's group should be rejected")
	}
}
