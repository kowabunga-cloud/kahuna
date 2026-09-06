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
	descX86 := NewVirtualInstanceDescription(TemplateOsLinux, "test-vm-x86", "description", "x86_64", "pc", "/usr/bin/qemu-system-x86_64", 2147483648, 2, true)
	if descX86.domain.OS.Type.Machine != "q35" {
		t.Errorf("expected machine 'q35' for x86_64, got %q", descX86.domain.OS.Type.Machine)
	}

	// amd64 must also be upgraded to q35
	descAmd64 := NewVirtualInstanceDescription(TemplateOsLinux, "test-vm-amd64", "description", "amd64", "pc-i440fx-7.0", "/usr/bin/qemu-system-x86_64", 2147483648, 2, true)
	if descAmd64.domain.OS.Type.Machine != "q35" {
		t.Errorf("expected machine 'q35' for amd64, got %q", descAmd64.domain.OS.Type.Machine)
	}

	// non-x86 (e.g. aarch64) should retain its designated machine
	descArm := NewVirtualInstanceDescription(TemplateOsLinux, "test-vm-arm", "description", "aarch64", "virt", "/usr/bin/qemu-system-aarch64", 2147483648, 2, false)
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
	desc := NewVirtualInstanceDescription(TemplateOsLinux, "test-vm", "test description", "x86_64", "pc", "/usr/bin/qemu-system-x86_64", 1073741824, 1, false)
	xmlStr, err := desc.XML()
	if err != nil {
		t.Fatalf("unexpected error generating XML: %v", err)
	}
	if !strings.Contains(xmlStr, "machine='q35'") && !strings.Contains(xmlStr, `machine="q35"`) {
		t.Errorf("expected XML to contain machine='q35', got: %s", xmlStr)
	}
}

func TestVirtualInstanceDescription_UefiFirmware(t *testing.T) {
	// uefi = true should set firmware='efi'
	descUefi := NewVirtualInstanceDescription(TemplateOsLinux, "test-uefi", "description", "x86_64", "q35", "/usr/bin/qemu-system-x86_64", 1073741824, 1, true)
	if descUefi.domain.OS.Firmware != "efi" {
		t.Errorf("expected OS firmware 'efi', got %q", descUefi.domain.OS.Firmware)
	}
	xmlUefi, err := descUefi.XML()
	if err != nil {
		t.Fatalf("unexpected error generating XML: %v", err)
	}
	if !strings.Contains(xmlUefi, "firmware='efi'") && !strings.Contains(xmlUefi, `firmware="efi"`) {
		t.Errorf("expected XML to contain firmware='efi', got: %s", xmlUefi)
	}

	// uefi = false should not set firmware
	descBios := NewVirtualInstanceDescription(TemplateOsLinux, "test-bios", "description", "x86_64", "q35", "/usr/bin/qemu-system-x86_64", 1073741824, 1, false)
	if descBios.domain.OS.Firmware != "" {
		t.Errorf("expected empty OS firmware for legacy BIOS, got %q", descBios.domain.OS.Firmware)
	}
	xmlBios, err := descBios.XML()
	if err != nil {
		t.Fatalf("unexpected error generating XML: %v", err)
	}
	if strings.Contains(xmlBios, "firmware=") {
		t.Errorf("expected XML to not contain firmware attribute, got: %s", xmlBios)
	}
}

func TestVirtualInstanceDescription_Windows11Topology(t *testing.T) {
	// Windows with UEFI (Windows 11)
	descWinUefi := NewVirtualInstanceDescription(TemplateOsWindows, "test-win11", "Windows 11 VM", "x86_64", "q35", "/usr/bin/qemu-system-x86_64", 4294967296, 4, true)

	// Machine must be q35
	if descWinUefi.domain.OS.Type.Machine != "q35" {
		t.Errorf("expected machine 'q35', got %q", descWinUefi.domain.OS.Type.Machine)
	}

	// Firmware must be efi
	if descWinUefi.domain.OS.Firmware != "efi" {
		t.Errorf("expected firmware 'efi', got %q", descWinUefi.domain.OS.Firmware)
	}

	// Firmware features must include enrolled-keys and secure-boot
	if descWinUefi.domain.OS.FirmwareInfo == nil || len(descWinUefi.domain.OS.FirmwareInfo.Features) != 2 {
		t.Fatalf("expected 2 firmware features for Windows 11, got: %v", descWinUefi.domain.OS.FirmwareInfo)
	}
	hasEnrolledKeys := false
	hasSecureBoot := false
	for _, f := range descWinUefi.domain.OS.FirmwareInfo.Features {
		if f.Name == "enrolled-keys" && f.Enabled == "yes" {
			hasEnrolledKeys = true
		}
		if f.Name == "secure-boot" && f.Enabled == "yes" {
			hasSecureBoot = true
		}
	}
	if !hasEnrolledKeys || !hasSecureBoot {
		t.Errorf("expected enrolled-keys and secure-boot enabled, got features: %v", descWinUefi.domain.OS.FirmwareInfo.Features)
	}

	// SMM must be on
	if descWinUefi.domain.Features.SMM == nil || descWinUefi.domain.Features.SMM.State != "on" {
		t.Errorf("expected SMM feature on, got %v", descWinUefi.domain.Features.SMM)
	}

	// TPM 2.0 CRB emulator must be present
	if len(descWinUefi.domain.Devices.TPMs) != 1 {
		t.Fatalf("expected 1 TPM device, got %d", len(descWinUefi.domain.Devices.TPMs))
	}
	tpm := descWinUefi.domain.Devices.TPMs[0]
	if tpm.Model != "tpm-crb" {
		t.Errorf("expected TPM model 'tpm-crb', got %q", tpm.Model)
	}
	if tpm.Backend == nil || tpm.Backend.Emulator == nil || tpm.Backend.Emulator.Version != "2.0" {
		t.Errorf("expected TPM emulator version '2.0', got: %v", tpm.Backend)
	}

	// Clock offset must be localtime
	if descWinUefi.domain.Clock.Offset != "localtime" {
		t.Errorf("expected clock offset 'localtime', got %q", descWinUefi.domain.Clock.Offset)
	}

	// USB tablet input device must be present
	hasTablet := false
	for _, in := range descWinUefi.domain.Devices.Inputs {
		if in.Type == "tablet" && in.Bus == "usb" {
			hasTablet = true
			break
		}
	}
	if !hasTablet {
		t.Errorf("expected USB tablet input device in: %v", descWinUefi.domain.Devices.Inputs)
	}

	// Extended Hyper-V enlightenments must be on
	hv := descWinUefi.domain.Features.HyperV
	if hv == nil {
		t.Fatalf("expected Hyper-V features to be configured")
	}
	if hv.Relaxed == nil || hv.Relaxed.State != "on" ||
		hv.VAPIC == nil || hv.VAPIC.State != "on" ||
		hv.Spinlocks == nil || hv.Spinlocks.State != "on" ||
		hv.VPIndex == nil || hv.VPIndex.State != "on" ||
		hv.Runtime == nil || hv.Runtime.State != "on" ||
		hv.Synic == nil || hv.Synic.State != "on" ||
		hv.STimer == nil || hv.STimer.State != "on" ||
		hv.Reset == nil || hv.Reset.State != "on" ||
		hv.Frequencies == nil || hv.Frequencies.State != "on" ||
		hv.TLBFlush == nil || hv.TLBFlush.State != "on" ||
		hv.IPI == nil || hv.IPI.State != "on" {
		t.Errorf("expected all extended Hyper-V enlightenments to be 'on', got: %+v", hv)
	}

	// XML generation verification
	xmlStr, err := descWinUefi.XML()
	if err != nil {
		t.Fatalf("unexpected error generating Windows 11 XML: %v", err)
	}
	expectedStrings := []string{
		"q35",
		"firmware=\"efi\"",
		"enrolled-keys",
		"secure-boot",
		"smm state=\"on\"",
		"model=\"tpm-crb\"",
		"type=\"emulator\" version=\"2.0\"",
		"offset=\"localtime\"",
		"type=\"tablet\" bus=\"usb\"",
		"relaxed state=\"on\"",
		"vapic state=\"on\"",
		"spinlocks state=\"on\"",
		"vpindex state=\"on\"",
		"runtime state=\"on\"",
		"synic state=\"on\"",
		"stimer state=\"on\"",
		"reset state=\"on\"",
		"frequencies state=\"on\"",
		"tlbflush state=\"on\"",
		"ipi state=\"on\"",
	}
	for _, s := range expectedStrings {
		if !strings.Contains(xmlStr, s) {
			t.Errorf("expected Windows 11 XML to contain %q, but was missing in:\n%s", s, xmlStr)
		}
	}

	// Windows with legacy BIOS (non-UEFI)
	descWinBios := NewVirtualInstanceDescription(TemplateOsWindows, "test-win-bios", "Windows Legacy VM", "x86_64", "q35", "/usr/bin/qemu-system-x86_64", 4294967296, 4, false)
	if descWinBios.domain.OS.Firmware != "" {
		t.Errorf("expected empty firmware for legacy BIOS, got %q", descWinBios.domain.OS.Firmware)
	}
	if descWinBios.domain.OS.FirmwareInfo != nil {
		t.Errorf("expected nil FirmwareInfo for legacy BIOS, got %v", descWinBios.domain.OS.FirmwareInfo)
	}
	if descWinBios.domain.Features.SMM != nil {
		t.Errorf("expected nil SMM for legacy BIOS, got %v", descWinBios.domain.Features.SMM)
	}
	if len(descWinBios.domain.Devices.TPMs) != 0 {
		t.Errorf("expected no TPM for legacy BIOS, got %v", descWinBios.domain.Devices.TPMs)
	}
	// But clock offset is still localtime and tablet is present
	if descWinBios.domain.Clock.Offset != "localtime" {
		t.Errorf("expected localtime offset for legacy Windows, got %q", descWinBios.domain.Clock.Offset)
	}
}
