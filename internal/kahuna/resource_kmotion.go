/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"time"

	"github.com/kowabunga-cloud/kahuna/internal/sdk"
)

const (
	kMotionShutdownTimeout      = 60 * time.Second
	kMotionShutdownPollInterval = 2 * time.Second
)

const (
	MongoCollectionKMotionPlanSchemaVersion = 1
	MongoCollectionKMotionPlanName          = "kmotion_plan"

	KMotionDestinationManual = "manual"
	KMotionDestinationZone   = "zone"
	KMotionDestinationRegion = "region"
	KMotionDestinationAuto   = "auto"

	// ErrKMotionLiveNotSupported is returned (and specifically matched on by
	// the route handlers, to map it to a distinct HTTP status) when a plan
	// asks for a live migration: only cold (live=false) migrations are
	// currently supported, since the Kaktus agent RPC protocol has no live
	// migration call yet.
	ErrKMotionLiveNotSupported = "live kMotion is not yet supported; only cold (non-live) migrations are currently available"
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
