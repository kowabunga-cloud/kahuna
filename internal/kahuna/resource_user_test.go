/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func userWithRole(role string) *User {
	return &User{Resource: NewResource("testuser", "", 1), Role: role}
}

func TestIsSuperAdmin(t *testing.T) {
	if !userWithRole(UserRoleSuperAdmin).IsSuperAdmin() {
		t.Error("superAdmin should return true for IsSuperAdmin")
	}
	if userWithRole(UserRoleProjectAdmin).IsSuperAdmin() {
		t.Error("projectAdmin should return false for IsSuperAdmin")
	}
	if userWithRole(UserRoleStandard).IsSuperAdmin() {
		t.Error("standard user should return false for IsSuperAdmin")
	}
}

func TestIsProjectAdmin(t *testing.T) {
	if !userWithRole(UserRoleProjectAdmin).IsProjectAdmin() {
		t.Error("projectAdmin should return true for IsProjectAdmin")
	}
	if !userWithRole(UserRoleSuperAdmin).IsProjectAdmin() {
		t.Error("superAdmin should return true for IsProjectAdmin")
	}
	if userWithRole(UserRoleStandard).IsProjectAdmin() {
		t.Error("standard user should return false for IsProjectAdmin")
	}
}

func TestUserVerifyCorrectPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	u := &User{PasswordHash: string(hash)}
	if err := u.Verify("correctpass"); err != nil {
		t.Errorf("expected no error for correct password, got: %v", err)
	}
}

func TestUserVerifyWrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	u := &User{PasswordHash: string(hash)}
	if err := u.Verify("wrongpass"); err == nil {
		t.Error("expected error for wrong password")
	}
}
