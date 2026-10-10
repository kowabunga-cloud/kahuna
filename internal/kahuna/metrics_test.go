/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestSetMetricsPoolsAndStaleSeries(t *testing.T) {
	e := &KowabungaExporter{}
	e.SetupMetrics()

	m := &KowabungaMetrics{
		Regions: map[string]KowabungaRegionMetrics{
			"eu-west": {
				Pools: map[string]*KowabungaRegionPools{
					"ssd": {Capacity: 100},
					"hdd": {Capacity: 1000},
				},
			},
		},
		Projects: map[string]KowabungaProjectMetrics{
			"deleted-soon": {VCPUs: 4},
		},
	}
	e.SetMetrics(m)

	if n := testutil.CollectAndCount(e.Metrics[MetricPoolsCapacity]); n != 2 {
		t.Errorf("expected one capacity series per pool, got %d", n)
	}
	if v := testutil.ToFloat64(e.Metrics[MetricPoolsCapacity].WithLabelValues("eu-west", "hdd")); v != 1000 {
		t.Errorf("hdd pool capacity = %v, want 1000", v)
	}

	// project is deleted: its series must disappear on next scrape
	m.Projects = map[string]KowabungaProjectMetrics{}
	e.SetMetrics(m)
	if n := testutil.CollectAndCount(e.Metrics[MetricProjectCPU]); n != 0 {
		t.Errorf("deleted project series still exported (%d)", n)
	}
}
