/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"testing"
)

// Kaktus.UsageScore

func TestKaktusUsageScoreZero(t *testing.T) {
	k := &Kaktus{}
	if got := k.UsageScore(); got != 0 {
		t.Errorf("expected 0 for empty kaktus, got %d", got)
	}
}

func TestKaktusUsageScoreInstances(t *testing.T) {
	k := &Kaktus{Usage: KaktusResources{InstancesCount: 3}}
	expected := 3 * KaktusScoreFactorInstancesCount
	if got := k.UsageScore(); got != expected {
		t.Errorf("expected %d, got %d", expected, got)
	}
}

func TestKaktusUsageScoreVCPUs(t *testing.T) {
	k := &Kaktus{Usage: KaktusResources{VCPUs: 8}}
	expected := 8 * KaktusScoreFactorVCPUs
	if got := k.UsageScore(); got != expected {
		t.Errorf("expected %d, got %d", expected, got)
	}
}

func TestKaktusUsageScoreMemory(t *testing.T) {
	// 4 GiB of memory
	k := &Kaktus{Usage: KaktusResources{MemorySize: 4 * 1024 * 1024 * 1024}}
	expected := 4 * KaktusScoreFactorMemory
	if got := k.UsageScore(); got != expected {
		t.Errorf("expected %d (4 GiB * factor), got %d", expected, got)
	}
}

func TestKaktusUsageScoreCombined(t *testing.T) {
	k := &Kaktus{Usage: KaktusResources{
		InstancesCount: 2,
		VCPUs:          4,
		MemorySize:     2 * 1024 * 1024 * 1024, // 2 GiB
	}}
	expected := 2*KaktusScoreFactorInstancesCount + 4*KaktusScoreFactorVCPUs + 2*KaktusScoreFactorMemory
	if got := k.UsageScore(); got != expected {
		t.Errorf("expected %d, got %d", expected, got)
	}
}

// KaktusCPU.Model

func TestKaktusCPUModel(t *testing.T) {
	cpu := KaktusCPU{
		Arch:    "x86_64",
		Cores:   8,
		Modele:  "Intel Xeon",
		Sockets: 2,
		Threads: 16,
		Vendor:  "Intel",
	}
	m := cpu.Model()
	if m.Arch != "x86_64" {
		t.Errorf("Arch: expected %q, got %q", "x86_64", m.Arch)
	}
	if m.Cores != 8 {
		t.Errorf("Cores: expected 8, got %d", m.Cores)
	}
	if m.Model != "Intel Xeon" {
		t.Errorf("Model: expected %q, got %q", "Intel Xeon", m.Model)
	}
	if m.Sockets != 2 {
		t.Errorf("Sockets: expected 2, got %d", m.Sockets)
	}
	if m.Threads != 16 {
		t.Errorf("Threads: expected 16, got %d", m.Threads)
	}
	if m.Vendor != "Intel" {
		t.Errorf("Vendor: expected %q, got %q", "Intel", m.Vendor)
	}
}

// KaktusCapabilities.Model

func TestKaktusCapabilitiesModel(t *testing.T) {
	caps := KaktusCapabilities{
		CPU:    KaktusCPU{Arch: "arm64", Cores: 4},
		Memory: 16384,
	}
	m := caps.Model()
	if m.Cpu.Arch != "arm64" {
		t.Errorf("CPU.Arch: expected %q, got %q", "arm64", m.Cpu.Arch)
	}
	if m.Memory != 16384 {
		t.Errorf("Memory: expected 16384, got %d", m.Memory)
	}
}
