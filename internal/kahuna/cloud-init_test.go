/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"os"
	"path/filepath"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	TestCloudInitConfigDir = "/tmp/kowabunga/cloud-init"
)

func TestFormatAdapterDeviceNameLinux(t *testing.T) {
	ci := &CloudInit{}
	cases := []struct {
		index    int
		expected string
	}{
		{0, "ens3"},
		{1, "ens4"},
		{2, "ens5"},
	}
	for _, c := range cases {
		if got := ci.formatAdapterDeviceName(c.index, false); got != c.expected {
			t.Errorf("index=%d: expected %q, got %q", c.index, c.expected, got)
		}
	}
}

func TestFormatAdapterDeviceNameWindows(t *testing.T) {
	ci := &CloudInit{}
	cases := []struct {
		index    int
		expected string
	}{
		{0, "interface0"},
		{1, "interface1"},
		{2, "interface2"},
	}
	for _, c := range cases {
		if got := ci.formatAdapterDeviceName(c.index, true); got != c.expected {
			t.Errorf("index=%d: expected %q, got %q", c.index, c.expected, got)
		}
	}
}

func TestNewCloudInitVolumeSuffix(t *testing.T) {
	ci, err := NewCloudInit("myvm", CloudinitProfileLinux)
	if err != nil {
		t.Fatalf("NewCloudInit: %v", err)
	}
	defer func() { _ = os.RemoveAll(ci.TmpDir) }()

	expected := "myvm" + CloudInitVolumeSuffix
	if ci.Name != expected {
		t.Errorf("expected Name=%q, got %q", expected, ci.Name)
	}
	if ci.OS != CloudinitProfileLinux {
		t.Errorf("expected OS=%q, got %q", CloudinitProfileLinux, ci.OS)
	}
	if ci.TmpDir == "" {
		t.Error("expected non-empty TmpDir")
	}
}

func TestWindowsUserDataTemplate(t *testing.T) {
	testResultDir := TestCloudInitConfigDir
	err := os.MkdirAll(testResultDir, 0777)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	res := Resource{
		ID:   bson.NewObjectID(),
		Name: "roottest",
	}
	sub := &Subnet{
		Resource: res,
		CIDR:     "10.0.0.0/24",
		Gateway:  "10.0.0.1",
		DNS:      "superdns",
		Reserved: []*IPRange{
			{
				First: "10.0.0.1",
				Last:  "10.0.0.5",
			},
		},
		GwPool: []*IPRange{
			{
				First: "10.0.0.250",
				Last:  "10.0.0.252",
			},
		},
		Routes: []string{
			"10.3.0.0/24",
		},
		Application: "test",
		AdapterIDs:  []string{},
	}

	routesByInterface := make(map[string]Subnet)
	routesByInterface["dummyadapter"] = *sub
	data := UserDataSettings{
		Hostname:          "test-host",
		Domain:            "superdomain.com",
		RootPassword:      "superpass",
		ServiceUser:       "kowabunga",
		ServiceUserPubKey: "randomKey",
		MetadataAlias:     "curl smthg",
		InterfacesSubnet:  routesByInterface,
	}

	ci := &CloudInit{
		Name:     "ci",
		OS:       "windows",
		TmpDir:   testResultDir,
		IsoImage: "windows",
		IsoSize:  10,
	}
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	tpl := filepath.Join(dir, "/../../config/templates/windows/user_data.yml")
	err = ci.SetData(tpl, "kw_user_data_tests", data)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}
