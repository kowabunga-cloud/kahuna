/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"testing"

	"github.com/kowabunga-cloud/kahuna/internal/sdk"
)

// ProjectResources.Update / Model round-trip

func TestProjectResourcesUpdateAndModel(t *testing.T) {
	sdkRes := sdk.ProjectResources{
		Vcpus:     8,
		Memory:    16 * 1024 * 1024 * 1024,
		Storage:   500 * 1024 * 1024 * 1024,
		Instances: 10,
	}
	var pr ProjectResources
	pr.Update(sdkRes)

	if pr.VCPUs != 8 {
		t.Errorf("VCPUs: expected 8, got %d", pr.VCPUs)
	}
	if pr.MemorySize != uint64(sdkRes.Memory) {
		t.Errorf("MemorySize: expected %d, got %d", sdkRes.Memory, pr.MemorySize)
	}
	if pr.StorageSize != uint64(sdkRes.Storage) {
		t.Errorf("StorageSize: expected %d, got %d", sdkRes.Storage, pr.StorageSize)
	}
	if pr.InstancesCount != 10 {
		t.Errorf("InstancesCount: expected 10, got %d", pr.InstancesCount)
	}

	m := pr.Model()
	if m.Vcpus != 8 {
		t.Errorf("Model Vcpus: expected 8, got %d", m.Vcpus)
	}
	if m.Memory != sdkRes.Memory {
		t.Errorf("Model Memory: expected %d, got %d", sdkRes.Memory, m.Memory)
	}
	if m.Storage != sdkRes.Storage {
		t.Errorf("Model Storage: expected %d, got %d", sdkRes.Storage, m.Storage)
	}
	if m.Instances != 10 {
		t.Errorf("Model Instances: expected 10, got %d", m.Instances)
	}
}

// Project.AllowInstanceCreationOrUpdate

func newTestProject(quotaInstances, quotaVCPUs uint16, quotaMem uint64) *Project {
	return &Project{
		Resource: NewResource("test-project", "", 1),
		Quotas: ProjectResources{
			InstancesCount: quotaInstances,
			VCPUs:          quotaVCPUs,
			MemorySize:     quotaMem,
		},
	}
}

func TestAllowInstanceCreationUnlimited(t *testing.T) {
	p := newTestProject(0, 0, 0)
	if !p.AllowInstanceCreationOrUpdate(100, 64, 1024*1024*1024*1024) {
		t.Error("zero quotas should allow any resource amounts")
	}
}

func TestAllowInstanceCreationWithinQuota(t *testing.T) {
	p := newTestProject(5, 16, 32*1024*1024*1024)
	p.Usage.InstancesCount = 2
	p.Usage.VCPUs = 8
	p.Usage.MemorySize = 8 * 1024 * 1024 * 1024
	if !p.AllowInstanceCreationOrUpdate(1, 4, 4*1024*1024*1024) {
		t.Error("should allow creation within quota")
	}
}

func TestAllowInstanceCreationExceedsInstanceCount(t *testing.T) {
	p := newTestProject(3, 0, 0)
	p.Usage.InstancesCount = 3
	if p.AllowInstanceCreationOrUpdate(1, 0, 0) {
		t.Error("should deny when instance quota is already at limit")
	}
}

func TestAllowInstanceCreationExceedsCPU(t *testing.T) {
	p := newTestProject(0, 8, 0)
	p.Usage.VCPUs = 6
	if p.AllowInstanceCreationOrUpdate(0, 4, 0) {
		t.Error("should deny when adding CPUs would exceed quota")
	}
}

func TestAllowInstanceCreationExceedsMemory(t *testing.T) {
	p := newTestProject(0, 0, 16*1024*1024*1024)
	p.Usage.MemorySize = 12 * 1024 * 1024 * 1024
	if p.AllowInstanceCreationOrUpdate(0, 0, 8*1024*1024*1024) {
		t.Error("should deny when adding memory would exceed quota")
	}
}

func TestAllowInstanceCreationExactLimit(t *testing.T) {
	p := newTestProject(5, 0, 0)
	p.Usage.InstancesCount = 4
	if !p.AllowInstanceCreationOrUpdate(1, 0, 0) {
		t.Error("should allow creation that exactly reaches the limit")
	}
}

// Project.AllowVolumeCreationOrUpdate

func TestAllowVolumeCreationUnlimited(t *testing.T) {
	p := newTestProject(0, 0, 0)
	if !p.AllowVolumeCreationOrUpdate(1024 * 1024 * 1024 * 1024) {
		t.Error("zero storage quota should allow any volume size")
	}
}

func TestAllowVolumeCreationWithinQuota(t *testing.T) {
	p := newTestProject(0, 0, 0)
	p.Quotas.StorageSize = 100 * 1024 * 1024 * 1024
	p.Usage.StorageSize = 50 * 1024 * 1024 * 1024
	if !p.AllowVolumeCreationOrUpdate(20 * 1024 * 1024 * 1024) {
		t.Error("should allow volume creation within quota")
	}
}

func TestAllowVolumeCreationExceedsQuota(t *testing.T) {
	p := newTestProject(0, 0, 0)
	p.Quotas.StorageSize = 100 * 1024 * 1024 * 1024
	p.Usage.StorageSize = 90 * 1024 * 1024 * 1024
	if p.AllowVolumeCreationOrUpdate(20 * 1024 * 1024 * 1024) {
		t.Error("should deny volume creation that would exceed quota")
	}
}
