/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

/*
 * A Kowabunga Multi-Zones Resource (MZR) is a special instantiation of Kompute instances
 * existing in all of a given region's zones.
 * It provides a zone-level local service instance, allowing for local network affinity,
 * while ensuring region-level service availability through usage of cross-zones virtual IP addresses.
 */

import (
	"fmt"
	"slices"

	"github.com/kowabunga-cloud/common/klog"
)

const (
	MongoCollectionMzrSchemaVersion = 2
	MongoCollectionMzrName          = "mzr"
)

type MZR interface {
	MZR() (*MultiZonesResource, error)
}

type MultiZonesResource struct {
	// anonymous field, inheritance
	Resource `bson:"inline"`

	// parents
	ProjectID       string `bson:"project_id"`
	RegionID        string `bson:"region_id"`
	PrivateSubnetID string `bson:"private_subnet_id"`
	PublicSubnetID  string `bson:"public_subnet_id"`

	// properties
	Profile           string            `bson:"profile"`
	PrivateAdapterIDs map[string]string `bson:"private_adapter_ids"`
	PrivateVIPs       []string          `bson:"private_vips"`
	PublicAdapterIDs  map[string]string `bson:"public_adapter_ids"`
	PublicVIPs        []string          `bson:"public_vips"`
	VirtualIPs        []VirtualIP       `bson:"virtual_ips"`
	KomputeIDs        []string          `bson:"kompute_ids"`
}

func MzrMigrateSchema() error {
	for _, mzr := range FindMZRs() {
		if mzr.SchemaVersion == 0 || mzr.SchemaVersion == 1 {
			err := mzr.migrateSchemaV2()
			if err != nil {
				return err
			}

			// migrate data
			mzrReloaded, err := FindMZRByID(mzr.String())
			if err != nil {
				return err
			}
			if mzrReloaded.Profile == "kgw" {
				mzrReloaded.Profile = CloudinitProfileKawaii
				err = mzrReloaded.Save()
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func FindMZRs() []MultiZonesResource {
	return FindResources[MultiZonesResource](MongoCollectionMzrName)
}

func FindMZRByID(id string) (*MultiZonesResource, error) {
	return FindResourceByID[MultiZonesResource](MongoCollectionMzrName, id)
}

func FindMZRByName(name string) (*MultiZonesResource, error) {
	return FindResourceByName[MultiZonesResource](MongoCollectionMzrName, name)
}
func (mzr *MultiZonesResource) FindLocalPrivateIPs() ([]string, error) {
	localPrivateIps := []string{}
	for _, id := range mzr.KomputeIDs {
		k, err := FindKomputeByID(id)
		if err != nil {
			return localPrivateIps, err
		}
		i, err := FindInstanceByID(k.InstanceID)
		if err != nil {
			return localPrivateIps, err
		}
		localPrivateIps = append(localPrivateIps, i.LocalIP)
	}
	return localPrivateIps, nil
}

func NewMultiZonesResource(projectId, regionId, namePrefix, desc, profile, profileId string, cpu, mem, disk, data int64, publicSubnetId string, subnetPeerings []string) (*MultiZonesResource, error) {

	// find parent objects, allows to bail before creating anything
	prj, err := FindProjectByID(projectId)
	if err != nil {
		return nil, err
	}

	r, err := FindRegionByID(regionId)
	if err != nil {
		return nil, err
	}

	// use region's default pool
	poolId := r.Defaults.StoragePoolID

	// ensure storage pool exists
	p, err := FindStoragePoolByID(poolId)
	if err != nil {
		return nil, err
	}

	// find private subnet
	privateSubnetId, err := prj.GetPrivateSubnet(regionId)
	if err != nil {
		return nil, err
	}

	// use default linux template for resources
	templateId := p.Defaults.TemplateIDs.OS

	mzr := MultiZonesResource{
		Resource:          NewResource(namePrefix, desc, MongoCollectionMzrSchemaVersion),
		ProjectID:         projectId,
		RegionID:          regionId,
		PrivateSubnetID:   privateSubnetId,
		PublicSubnetID:    publicSubnetId,
		Profile:           profile,
		PrivateAdapterIDs: make(map[string]string),
		PrivateVIPs:       []string{},
		PublicAdapterIDs:  make(map[string]string),
		PublicVIPs:        []string{},
		VirtualIPs:        []VirtualIP{},
	}

	// on failure, release whatever has been allocated so far (VIP adapters,
	// VRRP IDs, Komputes), as they get recorded into the object
	rb := newRollback("Multi-Zones Resource " + mzr.Name)
	defer rb.runUnlessDone()
	rb.add("allocated resources", mzr.releaseResources)

	if mzr.Profile == CloudinitProfileKawaii {
		// reserve public adapters for virtual IPs
		// NOTE: NAT rules public IPs are not allocated here,
		// they should have first be registered by API calls
		err := mzr.RequestPublicVIPs()
		if err != nil {
			return nil, err
		}
	} else {
		// MZR must reserve (and bind) private virtual IPs.
		// Note: kawaii doesn't need to, as it's using fixed zone-local gateways instead
		err := mzr.RequestPrivateVIPs()
		if err != nil {
			return nil, err
		}
	}

	err = mzr.GetVirtualIPs()
	if err != nil {
		return nil, err
	}

	// create a Kompute instance in each zone
	for _, zoneId := range r.Zones() {
		z, err := FindZoneByID(zoneId)
		if err != nil {
			return nil, err
		}

		// pick best host from zone
		mzrName := fmt.Sprintf("%s-%s", namePrefix, z.Name)
		h, err := z.ElectMostFavorableKaktus(mzrName, z.Kaktuses(), map[string]bool{}, map[string]bool{})
		if err != nil {
			return nil, err
		}

		name := fmt.Sprintf("%s-1", mzrName)

		// spin-up instances
		kompute, err := NewKompute(projectId, zoneId, h.String(), poolId, templateId,
			name, desc, mzr.Profile, profileId, cpu, mem, disk, 0, false, subnetPeerings, true)
		if err != nil {
			return nil, err
		}

		mzr.KomputeIDs = append(mzr.KomputeIDs, kompute.String())
	}

	klog.Debugf("Created new Multi-Zones Resource %s", mzr.String())
	_, err = GetDB().Insert(MongoCollectionMzrName, mzr)
	if err != nil {
		return nil, err
	}

	rb.done()
	return &mzr, nil
}

func (mzr *MultiZonesResource) renameDbField(from, to string) error {
	return GetDB().Rename(MongoCollectionMzrName, mzr.ID, from, to)
}

func (mzr *MultiZonesResource) setSchemaVersion(version int) error {
	return GetDB().SetSchemaVersion(MongoCollectionMzrName, mzr.ID, version)
}

func (mzr *MultiZonesResource) migrateSchemaV2() error {
	err := mzr.renameDbField("project", "project_id")
	if err != nil {
		return err
	}

	err = mzr.renameDbField("region", "region_id")
	if err != nil {
		return err
	}

	err = mzr.renameDbField("private_subnet", "private_subnet_id")
	if err != nil {
		return err
	}

	err = mzr.renameDbField("public_subnet", "public_subnet_id")
	if err != nil {
		return err
	}

	err = mzr.renameDbField("private_adapters", "private_adapter_ids")
	if err != nil {
		return err
	}

	err = mzr.renameDbField("public_adapters", "public_adapter_ids")
	if err != nil {
		return err
	}

	err = mzr.renameDbField("kces", "kompute_ids")
	if err != nil {
		return err
	}

	err = mzr.setSchemaVersion(2)
	if err != nil {
		return err
	}

	return nil
}

func (mzr *MultiZonesResource) Project() (*Project, error) {
	return FindProjectByID(mzr.ProjectID)
}

func (mzr *MultiZonesResource) RequestVIP(subnetId, prefix string, id int, public bool) (*Adapter, error) {
	tp := "private"
	if public {
		tp = "public"
	}
	adapterName := fmt.Sprintf("%s-VIP-%s-adapter-%d", prefix, tp, id)
	adapterDesc := fmt.Sprintf("%s network adapter for %s", prefix, tp)
	klog.Debugf("Creating %s VIP adapter for %s on subnet %s", tp, prefix, subnetId)
	return NewAdapter(subnetId, adapterName, adapterDesc, "", []string{}, false, true)
}

func (mzr *MultiZonesResource) RequestVIPs(public bool) error {
	r, err := FindRegionByID(mzr.RegionID)
	if err != nil {
		return err
	}

	for _, zoneId := range r.Zones() {
		z, err := FindZoneByID(zoneId)
		if err != nil {
			return err
		}

		prefix := fmt.Sprintf("%s-%s", mzr.Name, z.Name)

		var id int
		var subnetId string
		if public {
			id = len(mzr.PublicAdapterIDs)
			subnetId = mzr.PublicSubnetID
		} else {
			id = len(mzr.PrivateAdapterIDs)
			subnetId = mzr.PrivateSubnetID
		}

		adapter, err := mzr.RequestVIP(subnetId, prefix, id, public)
		if err != nil {
			return err
		}

		if public {
			mzr.PublicAdapterIDs[z.Name] = adapter.String()
			mzr.PublicVIPs = append(mzr.PublicVIPs, adapter.Addresses[0])
		} else {
			mzr.PrivateAdapterIDs[z.Name] = adapter.String()
			mzr.PrivateVIPs = append(mzr.PrivateVIPs, adapter.Addresses[0])
		}
	}

	return nil
}

func (mzr *MultiZonesResource) RequestPrivateVIPs() error {
	return mzr.RequestVIPs(false)
}

func (mzr *MultiZonesResource) RequestPublicVIPs() error {
	return mzr.RequestVIPs(true)
}

func (mzr *MultiZonesResource) GetVirtualIPs() error {
	// perform various checks before going any further ...
	publicInterface := fmt.Sprintf("%s%d", AdapterOsNicLinuxPrefix, AdapterOsNicLinuxStartIndex)
	publicSubnet, err := FindSubnetByID(mzr.PublicSubnetID)
	if err != nil {
		return err
	}

	privateInterface := fmt.Sprintf("%s%d", AdapterOsNicLinuxPrefix, AdapterOsNicLinuxStartIndex+1)
	privateSubnet, err := FindSubnetByID(mzr.PrivateSubnetID)
	if err != nil {
		return err
	}

	// public VIPs, if any ...
	slices.Sort(mzr.PublicVIPs)

	// Reserving [zoneName]vrIDs for kawaii private/public map
	vrIDs := map[string]int{}

	for zoneName, adapterId := range mzr.PublicAdapterIDs {
		adapter, err := FindAdapterByID(adapterId)
		if err != nil {
			return err
		}

		if slices.Contains(mzr.PublicVIPs, adapter.Addresses[0]) {
			// project will be updated at each iteration so we need updated object
			prj, err := mzr.Project()
			if err != nil {
				return err
			}
			// mzr.PublicAdapterIDs
			vrid, err := prj.AllocateVRID()
			if err != nil {
				return err
			}
			vrIDs[zoneName] = vrid

			vip := VirtualIP{
				VRRP:        vrid,
				Interface:   publicInterface,
				VIP:         adapter.Addresses[0],
				Priority:    VrrpPriorityBackup,
				NetMaskSize: publicSubnet.Size(),
				Public:      true,
			}
			mzr.VirtualIPs = append(mzr.VirtualIPs, vip)
		}
	}
	// private VIPs
	slices.Sort(mzr.PrivateVIPs)
	for _, v := range mzr.PrivateVIPs {
		// project will be updated at each iteration so we need updated object
		prj, err := mzr.Project()
		if err != nil {
			return err
		}

		vrid, err := prj.AllocateVRID()
		if err != nil {
			return err
		}

		vip := VirtualIP{
			VRRP:        vrid,
			Interface:   privateInterface,
			VIP:         v,
			Priority:    VrrpPriorityBackup,
			NetMaskSize: privateSubnet.Size(),
			Public:      false,
		}
		mzr.VirtualIPs = append(mzr.VirtualIPs, vip)
	}

	// Kawaii specifics
	if mzr.Profile == CloudinitProfileKawaii {
		prj, err := FindProjectByID(mzr.ProjectID)
		if err != nil {
			return err
		}

		// private local-zone gateways
		// ensure we sort out gateways by zone name, reflecting correct insertion order
		keys := make([]string, 0, len(prj.ZoneGateways))

		if len(prj.ZoneGateways) != len(vrIDs) {
			klog.Errorf("the number of public reserved IPs does not " +
				"match to the number of Zone Gateways.")
		}

		for k := range prj.ZoneGateways {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		for _, k := range keys {
			gw, ok := prj.ZoneGateways[k]
			if !ok {
				continue
			}
			vrid, ok := vrIDs[k]
			if !ok {
				return fmt.Errorf("no public virtual IP was reserved for zone %s gateway", k)
			}

			vip := VirtualIP{
				VRRP:        vrid,
				Interface:   privateInterface,
				VIP:         gw,
				Priority:    VrrpPriorityBackup,
				NetMaskSize: privateSubnet.Size(),
				Public:      false,
			}
			mzr.VirtualIPs = append(mzr.VirtualIPs, vip)
		}
	}
	return nil
}

func (mzr *MultiZonesResource) Save() error {
	mzr.Updated()
	return resourceUpdate(MongoCollectionMzrName, mzr.ID, mzr)
}

func (mzr *MultiZonesResource) Delete() error {
	err := mzr.releaseResources()
	if err != nil {
		return err
	}
	return resourceDelete(MongoCollectionMzrName, mzr.ID)
}

// releaseResources destroys the MZR's Komputes and VIP adapters and releases
// its VRRP IDs. Already released ones are skipped, so that it can be resumed
// after a partial failure, or used to roll back a partial creation.
func (mzr *MultiZonesResource) releaseResources() error {
	klog.Debugf("Deleting MZR Kompute instances %s", mzr.String())

	// Destroy underlying Komputes
	for _, komputeId := range mzr.KomputeIDs {
		kompute, err := FindKomputeByID(komputeId)
		if err == nil {
			err = kompute.Delete()
		}
		if ignoreMissing(err) != nil {
			klog.Error(err)
			return err
		}
	}

	// Destroy adapters
	adapters := []string{}
	for _, privateAdapterId := range mzr.PrivateAdapterIDs {
		adapters = append(adapters, privateAdapterId)
	}
	for _, publicAdapterId := range mzr.PublicAdapterIDs {
		adapters = append(adapters, publicAdapterId)
	}
	for _, adapterId := range adapters {
		a, err := FindAdapterByID(adapterId)
		if err == nil {
			err = a.Delete()
		}
		if ignoreMissing(err) != nil {
			return err
		}
	}

	// drop VRRP IDs from project
	prj, err := mzr.Project()
	if err != nil {
		return ignoreMissing(err)
	}
	for _, vip := range mzr.VirtualIPs {
		if vip.VRRP == 0 {
			// not an allocated VRRP ID
			continue
		}
		err = prj.RemoveVRID(vip.VRRP)
		if err != nil {
			return err
		}
	}

	return nil
}
