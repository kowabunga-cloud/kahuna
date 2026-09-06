/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"strings"
	"testing"
)

func TestVirtualInstanceDescription_Q35Upgrade(t *testing.T) {
	// x86_64 must be upgraded to q35 regardless of input machine
	descX86 := NewVirtualInstanceDescription(TemplateOsLinux, "test-vm-x86", "description", "x86_64", "pc", "/usr/bin/qemu-system-x86_64", 2147483648, 2)
	if descX86.domain.OS.Type.Machine != "q35" {
		t.Errorf("expected machine 'q35' for x86_64, got %q", descX86.domain.OS.Type.Machine)
	}

	// amd64 must also be upgraded to q35
	descAmd64 := NewVirtualInstanceDescription(TemplateOsLinux, "test-vm-amd64", "description", "amd64", "pc-i440fx-7.0", "/usr/bin/qemu-system-x86_64", 2147483648, 2)
	if descAmd64.domain.OS.Type.Machine != "q35" {
		t.Errorf("expected machine 'q35' for amd64, got %q", descAmd64.domain.OS.Type.Machine)
	}

	// non-x86 (e.g. aarch64) should retain its designated machine
	descArm := NewVirtualInstanceDescription(TemplateOsLinux, "test-vm-arm", "description", "aarch64", "virt", "/usr/bin/qemu-system-aarch64", 2147483648, 2)
	if descArm.domain.OS.Type.Machine != "virt" {
		t.Errorf("expected machine 'virt' for aarch64, got %q", descArm.domain.OS.Type.Machine)
	}
}

func TestNewVirtualDisk_IsoSata(t *testing.T) {
	isoDisk := NewVirtualDisk(VolumeTypeIso, "pool", "sda", "cloudinit.iso", "127.0.0.1", "secret-uuid", 6789)
	if isoDisk.Device != "cdrom" {
		t.Errorf("expected device 'cdrom', got %q", isoDisk.Device)
	}
	if isoDisk.Target.Bus != "sata" {
		t.Errorf("expected bus 'sata' for ISO volume, got %q", isoDisk.Target.Bus)
	}

	rawDisk := NewVirtualDisk(VolumeTypeRaw, "pool", "vda", "os.raw", "127.0.0.1", "secret-uuid", 6789)
	if rawDisk.Device != "disk" {
		t.Errorf("expected device 'disk', got %q", rawDisk.Device)
	}
	if rawDisk.Target.Bus != "virtio" {
		t.Errorf("expected bus 'virtio' for raw volume, got %q", rawDisk.Target.Bus)
	}
}

func TestVirtualInstanceDescription_XMLGeneration(t *testing.T) {
	desc := NewVirtualInstanceDescription(TemplateOsLinux, "test-vm", "test description", "x86_64", "pc", "/usr/bin/qemu-system-x86_64", 1073741824, 1)
	xmlStr, err := desc.XML()
	if err != nil {
		t.Fatalf("unexpected error generating XML: %v", err)
	}
	if !strings.Contains(xmlStr, "machine='q35'") && !strings.Contains(xmlStr, `machine="q35"`) {
		t.Errorf("expected XML to contain machine='q35', got: %s", xmlStr)
	}
}
