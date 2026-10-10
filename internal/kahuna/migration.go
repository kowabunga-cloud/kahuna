/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

/* This file is a specific collection of DB schema migration or pruner helpers
   Kowabunga DB has changed over the course of time and we need to ensure that all objects
   are properly setup with the right fields before going any further.
   Check (and update when required) will only be done once, at startup.
*/

import (
	"fmt"

	"github.com/kowabunga-cloud/common/klog"
)

// findDanglingRefs returns the child references from ids whose resource no
// longer exists. Any other lookup error aborts, so that a transient failure
// never makes valid references being dropped.
func findDanglingRefs(ids []string, find func(id string) error, owner, kind string) ([]string, error) {
	dangling := []string{}
	for _, id := range ids {
		err := find(id)
		if err == nil {
			continue
		}
		if !isMissingResource(err) {
			return nil, fmt.Errorf("unable to verify %s %s referenced by %s: %w", kind, id, owner, err)
		}
		klog.Errorf("%s is referencing %s %s, which doesn't seem to exists anymore, dropping it ...", owner, kind, id)
		dangling = append(dangling, id)
	}
	return dangling, nil
}

// lookup adapts a typed resource finder to findDanglingRefs()
func lookup[T any](find func(id string) (*T, error)) func(id string) error {
	return func(id string) error {
		_, err := find(id)
		return err
	}
}

func kawaiiCleanupDereferencedChildren() error {
	// projects
	for _, prj := range FindProjects() {
		owner := "Project " + prj.Name
		refs := []struct {
			ids  *[]string
			find func(id string) error
			kind string
		}{
			{&prj.InstanceIDs, lookup(FindInstanceByID), "instance"},
			{&prj.VolumeIDs, lookup(FindVolumeByID), "volume"},
			{&prj.KomputeIDs, lookup(FindKomputeByID), "Kompute"},
			{&prj.KawaiiIDs, lookup(FindKawaiiByID), "Kawaii"},
			{&prj.KyloIDs, lookup(FindKyloByID), "Kylo"},
		}

		updated := false
		for _, ref := range refs {
			dangling, err := findDanglingRefs(*ref.ids, ref.find, owner, ref.kind)
			if err != nil {
				return err
			}
			for _, id := range dangling {
				RemoveChildRef(ref.ids, id)
				updated = true
			}
		}

		if updated {
			err := prj.Save()
			if err != nil {
				return err
			}
		}
	}

	// subnets
	for _, subnet := range FindSubnets() {
		dangling, err := findDanglingRefs(subnet.AdapterIDs, lookup(FindAdapterByID), "Subnet "+subnet.Name, "adapter")
		if err != nil {
			return err
		}
		if len(dangling) == 0 {
			continue
		}
		for _, id := range dangling {
			RemoveChildRef(&subnet.AdapterIDs, id)
		}
		err = subnet.Save()
		if err != nil {
			return err
		}
	}

	return nil
}

func migrateInstances() error {
	for _, i := range FindInstances() {
		updated := false
		if i.LocalIP == "" {
			device := fmt.Sprintf("%s%d", AdapterOsNicLinuxPrefix, AdapterOsNicLinuxStartIndex+1)
			if i.OS == TemplateOsWindows {
				device = fmt.Sprintf("%s%d", AdapterOsNicWindowsPrefix, AdapterOsNicWindowsStartIndex+1)
			}

			adapterId, ok := i.Interfaces[device]
			if !ok {
				continue
			}

			adapter, err := FindAdapterByID(adapterId)
			if err != nil {
				return err
			}
			if len(adapter.Addresses) > 0 {
				i.LocalIP = adapter.Addresses[0]
				updated = true
			}
		}

		if updated {
			klog.Infof("DB Schema migration of instance %s done.", i.Name)
			err := i.Save()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func dbSchemaMigration() error {
	var err error

	err = AdapterMigrateSchema()
	if err != nil {
		return err
	}

	err = AgentMigrateSchema()
	if err != nil {
		return err
	}

	err = DnsRecordMigrateSchema()
	if err != nil {
		return err
	}

	err = HarMigrateSchema()
	if err != nil {
		return err
	}

	err = InstanceMigrateSchema()
	if err != nil {
		return err
	}

	err = KaktusMigrateSchema()
	if err != nil {
		return err
	}

	err = KawaiiMigrateSchema()
	if err != nil {
		return err
	}

	err = KiwiMigrateSchema()
	if err != nil {
		return err
	}

	err = KomputeMigrateSchema()
	if err != nil {
		return err
	}

	err = KonveyMigrateSchema()
	if err != nil {
		return err
	}

	err = KyloMigrateSchema()
	if err != nil {
		return err
	}

	err = MzrMigrateSchema()
	if err != nil {
		return err
	}

	err = NfsMigrateSchema()
	if err != nil {
		return err
	}

	err = ProjectMigrateSchema()
	if err != nil {
		return err
	}

	err = RegionMigrateSchema()
	if err != nil {
		return err
	}

	err = StoragePoolMigrateSchema()
	if err != nil {
		return err
	}

	err = SubnetMigrateSchema()
	if err != nil {
		return err
	}

	err = TeamMigrateSchema()
	if err != nil {
		return err
	}

	err = TemplateMigrateSchema()
	if err != nil {
		return err
	}

	err = TokenMigrateSchema()
	if err != nil {
		return err
	}

	err = UserMigrateSchema()
	if err != nil {
		return err
	}

	err = VNetMigrateSchema()
	if err != nil {
		return err
	}

	err = VolumeMigrateSchema()
	if err != nil {
		return err
	}

	err = ZoneMigrateSchema()
	if err != nil {
		return err
	}

	return nil
}

func MigrateDatabaseSchema() error {
	klog.Infof("Migrating DB schema ...")

	err := dbSchemaMigration()
	if err != nil {
		return err
	}

	err = migrateInstances()
	if err != nil {
		return err
	}

	err = kawaiiCleanupDereferencedChildren()
	if err != nil {
		return err
	}

	err = GetDB().EnsureIndexes()
	if err != nil {
		return err
	}

	return nil
}
