/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"net/http"
	"testing"

	"github.com/kowabunga-cloud/kahuna/internal/sdk"
)

func TestCreateProjectZoneKomputeRejectsNegativeSizes(t *testing.T) {
	GetCache().Init(true, CacheTypeInMemory, 16, 15)
	region := Region{Resource: NewResource("region", "", MongoCollectionRegionSchemaVersion)}
	zone := Zone{Resource: NewResource("zone", "", MongoCollectionZoneSchemaVersion), RegionID: region.String()}
	prj := Project{Resource: NewResource("project", "", MongoCollectionProjectSchemaVersion)}
	GetCache().Set(MongoCollectionRegionName, region.String(), &region)
	GetCache().Set(MongoCollectionZoneName, zone.String(), &zone)
	GetCache().Set(MongoCollectionProjectName, prj.String(), &prj)

	s := &ProjectService{}
	cases := []sdk.Kompute{
		{Name: "vm", Vcpus: 1, Memory: 1, Disk: 1024, DataDisk: -1023},
		{Name: "vm", Vcpus: -1, Memory: 1, Disk: 1},
		{Name: "vm", Vcpus: 1, Memory: -1, Disk: 1},
		{Name: "vm", Vcpus: 1, Memory: 1, Disk: -1},
	}
	for _, k := range cases {
		res, _ := s.CreateProjectZoneKompute(t.Context(), prj.String(), zone.String(), k, "", "", false)
		if res.Code != http.StatusBadRequest {
			t.Errorf("Kompute %+v returned %d, want %d", k, res.Code, http.StatusBadRequest)
		}
	}
}

func TestDeleteProjectWithChildrenConflicts(t *testing.T) {
	GetCache().Init(true, CacheTypeInMemory, 16, 15)
	prj := Project{
		Resource:  NewResource("project", "", MongoCollectionProjectSchemaVersion),
		KawaiiIDs: []string{"cccccccccccccccccccccccc"},
	}
	GetCache().Set(MongoCollectionProjectName, prj.String(), &prj)

	res, _ := (&ProjectService{}).DeleteProject(t.Context(), prj.String())
	if res.Code != http.StatusConflict {
		t.Errorf("deleting a project with a Kawaii returned %d, want %d", res.Code, http.StatusConflict)
	}
}

func TestListProjectInstancesFilteredByLocation(t *testing.T) {
	GetCache().Init(true, CacheTypeInMemory, 16, 15)
	cache := func(coll string, id string, v any) { GetCache().Set(coll, id, v) }

	r1 := Region{Resource: NewResource("r1", "", MongoCollectionRegionSchemaVersion)}
	r2 := Region{Resource: NewResource("r2", "", MongoCollectionRegionSchemaVersion)}
	zA := Zone{Resource: NewResource("a", "", MongoCollectionZoneSchemaVersion), RegionID: r1.String()}
	zB := Zone{Resource: NewResource("b", "", MongoCollectionZoneSchemaVersion), RegionID: r1.String()}
	zC := Zone{Resource: NewResource("c", "", MongoCollectionZoneSchemaVersion), RegionID: r2.String()}
	kA := Kaktus{Resource: NewResource("ka", "", MongoCollectionKaktusSchemaVersion), ZoneID: zA.String()}
	kB := Kaktus{Resource: NewResource("kb", "", MongoCollectionKaktusSchemaVersion), ZoneID: zB.String()}
	kC := Kaktus{Resource: NewResource("kc", "", MongoCollectionKaktusSchemaVersion), ZoneID: zC.String()}
	iA := Instance{Resource: NewResource("ia", "", MongoCollectionInstanceSchemaVersion), KaktusID: kA.String()}
	iB := Instance{Resource: NewResource("ib", "", MongoCollectionInstanceSchemaVersion), KaktusID: kB.String()}
	iC := Instance{Resource: NewResource("ic", "", MongoCollectionInstanceSchemaVersion), KaktusID: kC.String()}
	prj := Project{
		Resource:    NewResource("project", "", MongoCollectionProjectSchemaVersion),
		InstanceIDs: []string{iA.String(), iB.String(), iC.String()},
	}
	for _, z := range []*Zone{&zA, &zB, &zC} {
		cache(MongoCollectionZoneName, z.String(), z)
	}
	for _, k := range []*Kaktus{&kA, &kB, &kC} {
		cache(MongoCollectionKaktusName, k.String(), k)
	}
	for _, i := range []*Instance{&iA, &iB, &iC} {
		cache(MongoCollectionInstanceName, i.String(), i)
	}
	cache(MongoCollectionProjectName, prj.String(), &prj)

	s := &ProjectService{}
	check := func(name string, res sdk.ImplResponse, expected ...string) {
		t.Helper()
		got, _ := res.Body.([]string)
		if len(got) != len(expected) {
			t.Fatalf("%s: got %v, want %v", name, got, expected)
		}
		for i := range expected {
			if got[i] != expected[i] {
				t.Fatalf("%s: got %v, want %v", name, got, expected)
			}
		}
	}

	res, _ := s.ListProjectZoneInstances(t.Context(), prj.String(), zA.String())
	check("zone A", res, iA.String())
	res, _ = s.ListProjectRegionInstances(t.Context(), prj.String(), r1.String())
	check("region 1", res, iA.String(), iB.String())
	res, _ = s.ListProjectRegionInstances(t.Context(), prj.String(), r2.String())
	check("region 2", res, iC.String())
}
