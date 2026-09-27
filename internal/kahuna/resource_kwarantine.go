/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"github.com/kowabunga-cloud/common/klog"
	"github.com/kowabunga-cloud/kahuna/internal/sdk"
)

const (
	MongoCollectionKwarantineSchemaVersion = 1
	MongoCollectionKwarantineName          = "kwarantine"

	KwarantinePolicyHost = "host"
	KwarantinePolicyZone = "zone"
)

type Kwarantine struct {
	// anonymous field, inheritance
	Resource `bson:"inline"`

	// parents
	ProjectID string `bson:"project_id"`

	// properties
	Policy string `bson:"policy"`

	// members: loose references to existing instances/komputes, not owned
	InstanceIDs []string `bson:"instance_ids"`
	KomputeIDs  []string `bson:"kompute_ids"`
}

func NewKwarantine(projectId, name, desc, policy string) (*Kwarantine, error) {
	k := Kwarantine{
		Resource:    NewResource(name, desc, MongoCollectionKwarantineSchemaVersion),
		ProjectID:   projectId,
		Policy:      policy,
		InstanceIDs: []string{},
		KomputeIDs:  []string{},
	}

	prj, err := k.Project()
	if err != nil {
		return nil, err
	}

	_, err = GetDB().Insert(MongoCollectionKwarantineName, k)
	if err != nil {
		return nil, err
	}

	klog.Debugf("Created new Kwarantine %s (%s)", k.String(), k.Name)

	// add Kwarantine to project
	err = prj.AddKwarantine(k.String())
	if err != nil {
		return nil, err
	}

	return &k, nil
}

func FindKwarantines() []Kwarantine {
	return FindResources[Kwarantine](MongoCollectionKwarantineName)
}

func FindKwarantinesByProject(projectId string) ([]Kwarantine, error) {
	return FindResourcesByKey[Kwarantine](MongoCollectionKwarantineName, "project_id", projectId)
}

func FindKwarantineByID(id string) (*Kwarantine, error) {
	return FindResourceByID[Kwarantine](MongoCollectionKwarantineName, id)
}

func FindKwarantineByName(name string) (*Kwarantine, error) {
	return FindResourceByName[Kwarantine](MongoCollectionKwarantineName, name)
}

// FindKwarantinesByInstance returns the Kwarantine groups a given instance is a member of.
func FindKwarantinesByInstance(instanceId string) []Kwarantine {
	groups := []Kwarantine{}
	for _, k := range FindKwarantines() {
		if HasChildRef(&k.InstanceIDs, instanceId) {
			groups = append(groups, k)
		}
	}
	return groups
}

// FindKwarantinesByKompute returns the Kwarantine groups a given Kompute is a member of.
func FindKwarantinesByKompute(komputeId string) []Kwarantine {
	groups := []Kwarantine{}
	for _, k := range FindKwarantines() {
		if HasChildRef(&k.KomputeIDs, komputeId) {
			groups = append(groups, k)
		}
	}
	return groups
}

// RemoveInstanceFromKwarantines detaches instanceId from every Kwarantine
// group referencing it, so deleting an instance never leaves a dangling
// membership behind.
func RemoveInstanceFromKwarantines(instanceId string) error {
	for _, k := range FindKwarantinesByInstance(instanceId) {
		err := k.RemoveInstance(instanceId)
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveKomputeFromKwarantines detaches komputeId from every Kwarantine
// group referencing it, so deleting a Kompute never leaves a dangling
// membership behind.
func RemoveKomputeFromKwarantines(komputeId string) error {
	for _, k := range FindKwarantinesByKompute(komputeId) {
		err := k.RemoveKompute(komputeId)
		if err != nil {
			return err
		}
	}
	return nil
}

// kwarantineMemberKaktus resolves the Kaktus node currently hosting a
// Kwarantine member (an instance or Kompute ID), or nil if it can't be
// resolved (e.g. the member no longer exists).
func kwarantineMemberKaktus(memberId string) *Kaktus {
	if i, err := FindInstanceByID(memberId); err == nil {
		if h, err := i.Kaktus(); err == nil {
			return h
		}
		return nil
	}

	if k, err := FindKomputeByID(memberId); err == nil {
		if i, err := k.Instance(); err == nil {
			if h, err := i.Kaktus(); err == nil {
				return h
			}
		}
	}

	return nil
}

// KwarantineExclusions returns the set of Kaktus node IDs and the set of
// availability zone IDs that memberId (an instance or Kompute ID) must not
// be (re)scheduled onto/into, due to its Kwarantine anti-affinity group
// memberships. memberId may be empty (e.g. a brand new instance that hasn't
// joined any group yet), in which case both sets are empty.
func KwarantineExclusions(memberId string) (excludedKaktuses, excludedZones map[string]bool) {
	excludedKaktuses = map[string]bool{}
	excludedZones = map[string]bool{}
	if memberId == "" {
		return
	}

	groups := append(FindKwarantinesByInstance(memberId), FindKwarantinesByKompute(memberId)...)
	for _, g := range groups {
		members := append(append([]string{}, g.InstanceIDs...), g.KomputeIDs...)
		for _, m := range members {
			if m == memberId {
				continue
			}

			h := kwarantineMemberKaktus(m)
			if h == nil {
				continue
			}

			if g.Policy == KwarantinePolicyZone {
				excludedZones[h.ZoneID] = true
			} else {
				excludedKaktuses[h.String()] = true
			}
		}
	}

	return
}

func (k *Kwarantine) Project() (*Project, error) {
	return FindProjectByID(k.ProjectID)
}

func (k *Kwarantine) AddInstance(id string) error {
	klog.Debugf("Adding instance %s to Kwarantine %s", id, k.String())
	AddChildRef(&k.InstanceIDs, id)
	return k.Save()
}

func (k *Kwarantine) RemoveInstance(id string) error {
	klog.Debugf("Removing instance %s from Kwarantine %s", id, k.String())
	RemoveChildRef(&k.InstanceIDs, id)
	return k.Save()
}

func (k *Kwarantine) AddKompute(id string) error {
	klog.Debugf("Adding Kompute %s to Kwarantine %s", id, k.String())
	AddChildRef(&k.KomputeIDs, id)
	return k.Save()
}

func (k *Kwarantine) RemoveKompute(id string) error {
	klog.Debugf("Removing Kompute %s from Kwarantine %s", id, k.String())
	RemoveChildRef(&k.KomputeIDs, id)
	return k.Save()
}

func (k *Kwarantine) Update(name, desc, policy string) error {
	k.UpdateResourceDefaults(name, desc)
	k.Policy = policy
	return k.Save()
}

func (k *Kwarantine) Save() error {
	k.Updated()
	return resourceUpdate(MongoCollectionKwarantineName, k.ID, k)
}

func (k *Kwarantine) Delete() error {
	klog.Debugf("Deleting Kwarantine %s", k.String())

	if k.String() == ResourceUnknown {
		return nil
	}

	// remove Kwarantine's reference from parent
	prj, err := k.Project()
	if err != nil {
		return err
	}
	err = prj.RemoveKwarantine(k.String())
	if err != nil {
		return err
	}

	return GetDB().Delete(MongoCollectionKwarantineName, k.ID)
}

func (k *Kwarantine) Model() sdk.Kwarantine {
	return sdk.Kwarantine{
		Id:          k.String(),
		Name:        k.Name,
		Description: k.Description,
		Policy:      k.Policy,
		Instances:   k.InstanceIDs,
		Komputes:    k.KomputeIDs,
	}
}
