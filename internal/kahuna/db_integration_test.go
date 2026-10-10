/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"fmt"
	"os"
	"sync"
	"testing"
)

// requireTestDB connects to the MongoDB instance pointed to by the
// KAHUNA_TEST_MONGO_URI environment variable, on a scratch database dropped
// once the test is over. Tests requiring it are skipped otherwise.
func requireTestDB(t *testing.T) *KowabungaDB {
	t.Helper()
	uri := os.Getenv("KAHUNA_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("KAHUNA_TEST_MONGO_URI not set, skipping MongoDB integration test")
	}

	db := GetDB()
	if err := db.Open(uri, "kahuna_integration_tests"); err != nil {
		t.Fatalf("unable to connect to test MongoDB: %v", err)
	}
	_ = db.DB.Drop(t.Context())
	t.Cleanup(func() {
		_ = db.DB.Drop(t.Context())
	})
	GetCache().Init(true, CacheTypeInMemory, 16, 15)
	return db
}

func TestIntegrationConcurrentChildRefsAndCounters(t *testing.T) {
	db := requireTestDB(t)

	prj := Project{Resource: NewResource("project", "", MongoCollectionProjectSchemaVersion)}
	if _, err := db.Insert(MongoCollectionProjectName, prj); err != nil {
		t.Fatal(err)
	}

	const n = 40
	instances := make([]Instance, n)
	for i := range instances {
		instances[i] = Instance{Resource: NewResource(fmt.Sprintf("vm%d", i), "", MongoCollectionInstanceSchemaVersion), CPU: 2, Memory: 1024}
		if _, err := db.Insert(MongoCollectionInstanceName, instances[i]); err != nil {
			t.Fatal(err)
		}
	}

	// concurrent creations, each working on its own (stale) copy of the project
	var wg sync.WaitGroup
	for i := range instances {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			p, err := FindProjectByID(prj.String())
			if err != nil {
				t.Error(err)
				return
			}
			if err := p.AddInstance(id); err != nil {
				t.Error(err)
			}
		}(instances[i].String())
	}
	wg.Wait()

	p, err := FindProjectByID(prj.String())
	if err != nil {
		t.Fatal(err)
	}
	if len(p.InstanceIDs) != n || p.Usage.InstancesCount != n || p.Usage.VCPUs != 2*n || p.Usage.MemorySize != 1024*n {
		t.Fatalf("lost updates: %d refs, usage %+v", len(p.InstanceIDs), p.Usage)
	}

	// adding an already referenced child is a no-op
	if err := p.AddInstance(instances[0].String()); err != nil {
		t.Fatal(err)
	}
	p, _ = FindProjectByID(prj.String())
	if len(p.InstanceIDs) != n || p.Usage.InstancesCount != n {
		t.Fatalf("duplicate add changed references or counters: %d refs, usage %+v", len(p.InstanceIDs), p.Usage)
	}

	// removals
	for i := range instances {
		p, _ := FindProjectByID(prj.String())
		if err := p.RemoveInstance(instances[i].String()); err != nil {
			t.Fatal(err)
		}
	}
	p, _ = FindProjectByID(prj.String())
	if len(p.InstanceIDs) != 0 || p.Usage.InstancesCount != 0 || p.Usage.VCPUs != 0 || p.Usage.MemorySize != 0 {
		t.Fatalf("unexpected state after removals: %d refs, usage %+v", len(p.InstanceIDs), p.Usage)
	}

	// counters never go below 0
	if err := p.UpdateInstanceUsage(-4, -2048); err != nil {
		t.Fatal(err)
	}
	p, _ = FindProjectByID(prj.String())
	if p.Usage.VCPUs != 0 || p.Usage.MemorySize != 0 {
		t.Fatalf("counters went negative: %+v", p.Usage)
	}
}

func TestIntegrationDefaultsAndMultiArrayRefs(t *testing.T) {
	db := requireTestDB(t)

	nfs := NFS{Resource: NewResource("nfs", "", MongoCollectionNfsSchemaVersion)}
	if _, err := db.Insert(MongoCollectionNfsName, nfs); err != nil {
		t.Fatal(err)
	}
	if err := nfs.AddKylo("k1", "e1"); err != nil {
		t.Fatal(err)
	}
	if err := nfs.AddKylo("k2", "e2"); err != nil {
		t.Fatal(err)
	}
	if err := nfs.RemoveKylo("k1", "e1"); err != nil {
		t.Fatal(err)
	}
	got, err := FindNfsByID(nfs.String())
	if err != nil {
		t.Fatal(err)
	}
	if len(got.KyloIDs) != 1 || got.KyloIDs[0] != "k2" || len(got.Exports) != 1 || got.Exports[0] != "e2" {
		t.Fatalf("unexpected NFS references: %v / %v", got.KyloIDs, got.Exports)
	}
}

func TestIntegrationConcurrentAllocations(t *testing.T) {
	db := requireTestDB(t)
	if err := db.EnsureIndexes(); err != nil {
		t.Fatal(err)
	}

	const n = 20

	// IP addresses
	subnet := Subnet{Resource: NewResource("subnet", "", MongoCollectionSubnetSchemaVersion), CIDR: "10.0.0.0/24", Gateway: "10.0.0.1"}
	if _, err := db.Insert(MongoCollectionSubnetName, subnet); err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	ips := map[string]bool{}
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			a, err := NewAdapter(subnet.String(), fmt.Sprintf("nic%d", i), "", "", nil, false, true)
			if err != nil {
				t.Error(err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if ips[a.Addresses[0]] {
				t.Errorf("IP %s assigned twice", a.Addresses[0])
			}
			ips[a.Addresses[0]] = true
		}(i)
	}
	wg.Wait()
	if len(ips) != n {
		t.Fatalf("expected %d distinct IPs, got %d", n, len(ips))
	}
	if _, err := NewAdapter(subnet.String(), "dup", "", "", []string{"10.0.0.2"}, false, false); err == nil {
		t.Error("explicitly assigning an already used IP should fail")
	}

	// VRRP IDs
	prj := Project{Resource: NewResource("project", "", MongoCollectionProjectSchemaVersion)}
	if _, err := db.Insert(MongoCollectionProjectName, prj); err != nil {
		t.Fatal(err)
	}
	vrids := map[int]bool{}
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p, err := FindProjectByID(prj.String())
			if err != nil {
				t.Error(err)
				return
			}
			vrid, err := p.AllocateVRID()
			if err != nil {
				t.Error(err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if vrids[vrid] {
				t.Errorf("VRID %d allocated twice", vrid)
			}
			vrids[vrid] = true
		}()
	}
	wg.Wait()
	p, _ := FindProjectByID(prj.String())
	if len(vrids) != n || len(p.VrrpIDs) != n {
		t.Fatalf("expected %d distinct VRIDs, got %d (%d stored)", n, len(vrids), len(p.VrrpIDs))
	}

	// private subnets
	free := Subnet{Resource: NewResource("free", "", MongoCollectionSubnetSchemaVersion), CIDR: "10.1.0.0/24"}
	if _, err := db.Insert(MongoCollectionSubnetName, free); err != nil {
		t.Fatal(err)
	}
	winners := 0
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s, err := FindSubnetByID(free.String())
			if err != nil {
				t.Error(err)
				return
			}
			claimed, err := s.claimForProject(fmt.Sprintf("project-%d", i))
			if err != nil {
				t.Error(err)
				return
			}
			if claimed {
				mu.Lock()
				winners++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	if winners != 1 {
		t.Fatalf("subnet claimed by %d projects", winners)
	}
}

func TestIntegrationResumableDeletions(t *testing.T) {
	db := requireTestDB(t)

	prj := Project{Resource: NewResource("project", "", MongoCollectionProjectSchemaVersion), VrrpIDs: []int{1, 2}}
	if _, err := db.Insert(MongoCollectionProjectName, prj); err != nil {
		t.Fatal(err)
	}
	subnet := Subnet{Resource: NewResource("subnet", "", MongoCollectionSubnetSchemaVersion), CIDR: "10.0.0.0/24", Gateway: "10.0.0.1"}
	if _, err := db.Insert(MongoCollectionSubnetName, subnet); err != nil {
		t.Fatal(err)
	}
	vip, err := NewAdapter(subnet.String(), "vip", "", "", nil, false, true)
	if err != nil {
		t.Fatal(err)
	}

	// a previous deletion attempt already destroyed the Kompute and one VIP
	// adapter, then failed
	mzr := MultiZonesResource{
		Resource:          NewResource("mzr", "", MongoCollectionMzrSchemaVersion),
		ProjectID:         prj.String(),
		KomputeIDs:        []string{"aaaaaaaaaaaaaaaaaaaaaaaa"},
		PrivateAdapterIDs: map[string]string{"z1": "bbbbbbbbbbbbbbbbbbbbbbbb", "z2": vip.String()},
		VirtualIPs:        []VirtualIP{{VRRP: 1}, {VRRP: 2}},
	}
	if _, err := db.Insert(MongoCollectionMzrName, mzr); err != nil {
		t.Fatal(err)
	}

	if err := mzr.Delete(); err != nil {
		t.Fatalf("resumed MZR deletion failed: %v", err)
	}
	if _, err := FindAdapterByID(vip.String()); err == nil {
		t.Error("remaining VIP adapter should have been deleted")
	}
	p, _ := FindProjectByID(prj.String())
	if len(p.VrrpIDs) != 0 {
		t.Errorf("VRRP IDs not released: %v", p.VrrpIDs)
	}
	if _, err := FindMZRByID(mzr.String()); err == nil {
		t.Error("MZR should have been deleted")
	}
}
