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
