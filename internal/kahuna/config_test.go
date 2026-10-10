/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func parseTestConfig(t *testing.T, contents string) error {
	t.Helper()
	path := filepath.Join(t.TempDir(), "kahuna.yml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path) // #nosec G304 -- test file
	if err != nil {
		t.Fatal(err)
	}
	_, err = ParseConfig(f)
	return err
}

func TestParseConfigRejectsInvalidBootstrapSettings(t *testing.T) {
	cases := map[string]string{
		"invalid user":   "global:\n  bootstrap:\n    user: \"root\\nruncmd\"\n    pubkey: ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIP6uPgayxgfvqC0PdKBFEFFhkFs/joe341L9QEZAbieY test\n",
		"invalid pubkey": "global:\n  bootstrap:\n    user: admin\n    pubkey: not-a-key\n",
	}
	for name, cfg := range cases {
		err := parseTestConfig(t, cfg)
		if err == nil || !strings.Contains(err.Error(), "bootstrap") {
			t.Errorf("%s: expected bootstrap config error, got %v", name, err)
		}
	}
}

func TestParseConfigRejectsInvalidJwtSettings(t *testing.T) {
	const bootstrap = `
global:
  bootstrap:
    user: admin
    pubkey: ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIP6uPgayxgfvqC0PdKBFEFFhkFs/joe341L9QEZAbieY test
`
	cases := map[string]string{
		"missing signature": bootstrap + "  jwt:\n    lifetimeHours: 24\n",
		"missing lifetime":  bootstrap + "  jwt:\n    signature: secret\n",
	}
	for name, cfg := range cases {
		err := parseTestConfig(t, cfg)
		if err == nil || !strings.Contains(err.Error(), "JWT") {
			t.Errorf("%s: expected JWT config error, got %v", name, err)
		}
	}
}
