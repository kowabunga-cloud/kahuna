/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"fmt"

	"github.com/kowabunga-cloud/kahuna/internal/sdk"
)

const (
	MongoCollectionKMotionPlanSchemaVersion = 1
	MongoCollectionKMotionPlanName          = "kmotion_plan"

	KMotionDestinationManual = "manual"
	KMotionDestinationZone   = "zone"
	KMotionDestinationRegion = "region"
	KMotionDestinationAuto   = "auto"
)

// KMotionPlan is a short-lived, single-use record of a proposed instance
// migration: it is created by a "plan" call and consumed (deleted) by the
// matching "commit" call. It is not a user-facing named resource (no
// name/description are ever set), just persisted so it survives across
// requests (and kahuna replicas) between plan and commit.
type KMotionPlan struct {
	// anonymous field, inheritance
	Resource `bson:"inline"`

	InstanceID string `bson:"instance_id"`
	SourceID   string `bson:"source_id"`
	TargetID   string `bson:"target_id"`
	Live       bool   `bson:"live"`
}

func NewKMotionPlan(instanceId, sourceId, targetId string, live bool) (*KMotionPlan, error) {
	p := KMotionPlan{
		Resource:   NewResource("", "", MongoCollectionKMotionPlanSchemaVersion),
		InstanceID: instanceId,
		SourceID:   sourceId,
		TargetID:   targetId,
		Live:       live,
	}

	_, err := GetDB().Insert(MongoCollectionKMotionPlanName, p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func FindKMotionPlanByID(id string) (*KMotionPlan, error) {
	return FindResourceByID[KMotionPlan](MongoCollectionKMotionPlanName, id)
}

func (p *KMotionPlan) Delete() error {
	if p.String() == ResourceUnknown {
		return nil
	}
	return GetDB().Delete(MongoCollectionKMotionPlanName, p.ID)
}

func (p *KMotionPlan) Model() sdk.KMotionPlan {
	return sdk.KMotionPlan{
		Id:       p.String(),
		Instance: p.InstanceID,
		Source:   p.SourceID,
		Target:   p.TargetID,
		Live:     p.Live,
	}
}

// PlanMigration computes a kMotion plan for the instance: it elects a
// destination Kaktus node per the requested destination mode, honoring
// maintenance mode, the instance's own current host (never re-elected) and
// its Kwarantine anti-affinity group memberships, regardless of mode
// (including manual). The resulting plan is persisted so it can be
// referenced by a later commit call.
func (i *Instance) PlanMigration(destination, explicitKaktus string, live bool) (*KMotionPlan, error) {
	if !i.KMotionEnabled {
		return nil, fmt.Errorf("kMotion is disabled for instance %s", i.String())
	}

	source, err := i.Kaktus()
	if err != nil {
		return nil, err
	}

	if destination == "" {
		destination = KMotionDestinationAuto
	}

	excludedKaktuses, excludedZones := KwarantineExclusions(i.String())
	// never (re)elect the instance's own current host
	excludedKaktuses[source.String()] = true

	target, err := i.electKMotionTarget(destination, explicitKaktus, source, excludedKaktuses, excludedZones)
	if err != nil {
		return nil, err
	}

	return NewKMotionPlan(i.String(), source.String(), target.String(), live)
}

func (i *Instance) electKMotionTarget(destination, explicitKaktus string, source *Kaktus, excludedKaktuses, excludedZones map[string]bool) (*Kaktus, error) {
	switch destination {
	case KMotionDestinationManual:
		if explicitKaktus == "" {
			return nil, fmt.Errorf("a destination kaktus must be specified for manual kMotion")
		}
		target, err := FindKaktusByID(explicitKaktus)
		if err != nil {
			return nil, err
		}
		if target.Maintenance {
			return nil, fmt.Errorf("kaktus %s is under maintenance", target.String())
		}
		if excludedKaktuses[target.String()] || excludedZones[target.ZoneID] {
			return nil, fmt.Errorf("kaktus %s violates a Kwarantine anti-affinity policy for instance %s", target.String(), i.String())
		}
		return target, nil

	case KMotionDestinationZone:
		zone, err := source.Zone()
		if err != nil {
			return nil, err
		}
		return zone.ElectMostFavorableKaktus(i.Name, zone.Kaktuses(), excludedKaktuses, excludedZones)

	case KMotionDestinationRegion, KMotionDestinationAuto:
		zone, err := source.Zone()
		if err != nil {
			return nil, err
		}
		region, err := zone.Region()
		if err != nil {
			return nil, err
		}
		return region.ElectMostFavorableKaktus(i.Name, excludedKaktuses, excludedZones)

	default:
		return nil, fmt.Errorf("unknown kMotion destination mode: %s", destination)
	}
}

// PlanMigration is the Kompute counterpart to Instance.PlanMigration,
// delegating to its underlying instance.
func (k *Kompute) PlanMigration(destination, explicitKaktus string, live bool) (*KMotionPlan, error) {
	i, err := k.Instance()
	if err != nil {
		return nil, err
	}
	return i.PlanMigration(destination, explicitKaktus, live)
}
