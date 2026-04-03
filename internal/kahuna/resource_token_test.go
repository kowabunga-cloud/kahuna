/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Token.HasExpired

func TestTokenHasExpiredNoExpiry(t *testing.T) {
	tok := &Token{Expire: false}
	if tok.HasExpired() {
		t.Error("non-expiring token should not be expired")
	}
}

func TestTokenHasExpiredFutureDate(t *testing.T) {
	future := time.Now().Add(24 * time.Hour).Format(time.DateOnly)
	tok := &Token{Expire: true, ExpirationDate: future}
	if tok.HasExpired() {
		t.Errorf("token expiring tomorrow (%s) should not be expired yet", future)
	}
}

func TestTokenHasExpiredPastDate(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour).Format(time.DateOnly)
	tok := &Token{Expire: true, ExpirationDate: past}
	if !tok.HasExpired() {
		t.Errorf("token that expired yesterday (%s) should be expired", past)
	}
}

func TestTokenHasExpiredInvalidDate(t *testing.T) {
	tok := &Token{Expire: true, ExpirationDate: "not-a-date"}
	if !tok.HasExpired() {
		t.Error("token with unparseable date should be treated as expired")
	}
}

// Token.Verify

func TestTokenVerifyCorrectKey(t *testing.T) {
	key := "mysecretapikey"
	hash, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	tok := &Token{ApiKeyHash: string(hash)}
	if err := tok.Verify(key); err != nil {
		t.Errorf("expected no error for correct key, got: %v", err)
	}
}

func TestTokenVerifyWrongKey(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correctkey"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	tok := &Token{ApiKeyHash: string(hash)}
	if err := tok.Verify("wrongkey"); err == nil {
		t.Error("expected error for wrong key")
	}
}
