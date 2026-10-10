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

func TestParseConfigRejectsInvalidJwtSettings(t *testing.T) {
	const bootstrap = `
global:
  bootstrap:
    user: admin
    pubkey: ssh-ed25519 AAAA
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
