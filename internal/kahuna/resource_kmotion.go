/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"fmt"
	"time"

	"github.com/kowabunga-cloud/common/klog"
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

// CommitMigration executes a previously computed kMotion plan. Only cold
// (live=false) migrations are currently supported; a live plan is rejected
// with ErrKMotionLiveNotSupported. The target is re-validated (still
// exists, not under maintenance, still Kwarantine-compliant) since cluster
// state may have changed since the plan was computed. The plan is consumed
// (deleted) on success.
func (i *Instance) CommitMigration(plan *KMotionPlan) error {
	if plan.InstanceID != i.String() {
		return fmt.Errorf("kMotion plan %s does not apply to instance %s", plan.String(), i.String())
	}

	if plan.Live {
		return fmt.Errorf("%s", ErrKMotionLiveNotSupported)
	}

	target, err := FindKaktusByID(plan.TargetID)
	if err != nil {
		return err
	}

	if target.Maintenance {
		return fmt.Errorf("kaktus %s is under maintenance, plan %s is no longer valid", target.String(), plan.String())
	}

	excludedKaktuses, excludedZones := KwarantineExclusions(i.String())
	if excludedKaktuses[target.String()] || excludedZones[target.ZoneID] {
		return fmt.Errorf("kaktus %s now violates a Kwarantine anti-affinity policy for instance %s, plan %s is no longer valid", target.String(), i.String(), plan.String())
	}

	if err := i.Migrate(target.String()); err != nil {
		return err
	}

	return plan.Delete()
}

// Migrate performs a cold migration of the instance to a different Kaktus
// computing node: it powers the instance off, undefines its libvirt domain
// from the current host (attached volumes are never touched, only
// re-attached to the redefined domain), moves it to the new host and
// redefines (and, if it was running, restarts) it there. If redefining on
// the target fails, it attempts a best-effort rollback onto the original
// source host.
func (i *Instance) Migrate(targetKaktusId string) error {
	source, err := i.Kaktus()
	if err != nil {
		return err
	}

	target, err := FindKaktusByID(targetKaktusId)
	if err != nil {
		return err
	}

	if source.String() == target.String() {
		return fmt.Errorf("instance %s is already hosted on kaktus %s", i.String(), target.String())
	}

	klog.Infof("Migrating instance %s from kaktus %s to kaktus %s", i.String(), source.String(), target.String())

	wasRunning := i.IsRunning()
	if wasRunning {
		if err := i.shutdownForMigration(); err != nil {
			return fmt.Errorf("unable to power off instance %s for migration: %w", i.String(), err)
		}
	}

	// undefine the domain from the source host; volumes/adapters are
	// untouched, only the compute definition moves
	if err := i.delete(); err != nil {
		return fmt.Errorf("unable to undefine instance %s on source kaktus %s: %w", i.String(), source.String(), err)
	}

	// move the instance to its new host
	i.KaktusID = target.String()
	if err := i.Save(); err != nil {
		return err
	}

	// redefine (and, if it was running, restart) on the target host
	if err := i.CreateInstance(); err != nil {
		klog.Errorf("unable to redefine instance %s on target kaktus %s, attempting rollback onto %s: %v", i.String(), target.String(), source.String(), err)

		i.KaktusID = source.String()
		if saveErr := i.Save(); saveErr != nil {
			klog.Errorf("rollback failed to restore instance %s kaktus assignment to %s: %v", i.String(), source.String(), saveErr)
			return err
		}
		if rollbackErr := i.CreateInstance(); rollbackErr != nil {
			klog.Errorf("rollback failed to redefine instance %s on source kaktus %s: %v", i.String(), source.String(), rollbackErr)
			return fmt.Errorf("migration of instance %s failed and rollback also failed, it may now be undefined on every host: %w", i.String(), err)
		}
		return err
	}

	// move Kaktus-side bookkeeping (usage counters, instance-list membership)
	if err := source.RemoveInstance(i.String()); err != nil {
		klog.Error(err)
	}
	if err := target.AddInstance(i.String()); err != nil {
		klog.Error(err)
	}

	return nil
}

// shutdownForMigration attempts a graceful ACPI shutdown, falling back to a
// hard power-off if the guest doesn't cooperate within the grace period, so
// a migration commit never hangs indefinitely on an unresponsive guest.
func (i *Instance) shutdownForMigration() error {
	if err := i.Shutdown(); err != nil {
		return err
	}
	if i.waitUntilStopped(kMotionShutdownTimeout) {
		return nil
	}

	klog.Warningf("instance %s did not shut down gracefully within %s, forcing power-off", i.String(), kMotionShutdownTimeout)
	if err := i.Stop(); err != nil {
		return err
	}
	if i.waitUntilStopped(kMotionShutdownTimeout) {
		return nil
	}

	return fmt.Errorf("instance did not power off within %s", kMotionShutdownTimeout)
}

func (i *Instance) waitUntilStopped(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !i.IsRunning() {
			return true
		}
		time.Sleep(kMotionShutdownPollInterval)
	}
	return !i.IsRunning()
}

// CommitMigration is the Kompute counterpart to Instance.CommitMigration.
func (k *Kompute) CommitMigration(plan *KMotionPlan) error {
	i, err := k.Instance()
	if err != nil {
		return err
	}
	return i.CommitMigration(plan)
}
