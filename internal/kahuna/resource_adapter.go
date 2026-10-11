/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/netip"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/kowabunga-cloud/common/klog"
	"github.com/kowabunga-cloud/kahuna/internal/sdk"
)

const (
	MongoCollectionAdapterSchemaVersion = 2
	MongoCollectionAdapterName          = "adapter"

	AdapterOsNicLinuxPrefix       = "ens"
	AdapterOsNicLinuxStartIndex   = 3
	AdapterOsNicWindowsPrefix     = "interface"
	AdapterOsNicWindowsStartIndex = 0
)

type Adapter struct {
	// anonymous field, inheritance
	Resource `bson:"inline"`

	// parents
	SubnetID string `bson:"subnet_id"`

	// properties
	MAC       string   `bson:"mac"`
	Addresses []string `bson:"addresses"`
	Reserved  bool     `bson:"reserved"`

	// children references
}

func AdapterMigrateSchema() error {
	// rename collection
	err := GetDB().RenameCollection("adapters", MongoCollectionAdapterName)
	if err != nil {
		return err
	}

	for _, adapter := range FindAdapters() {
		if adapter.SchemaVersion == 0 || adapter.SchemaVersion == 1 {
			err := adapter.migrateSchemaV2()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func RandomAdapterMAC() string {
	// QEMU vendor prefix - 52:54:00
	// 52:54:00:00:00:00 to 52:54:00:FF:FF:FF
	// Kowabunga vendor prefix - da:1e:70
	// da:1e:70:00:00:00 to da:1e:70:FF:FF:FF
	// This is still 2^24 (= 16 777 216) possibilities.
	for {
		var hw = make(net.HardwareAddr, 6)
		hw[0] = 0xda
		hw[1] = 0x1e
		hw[2] = 0x70

		// read is safe for concurrent use
		_, err := rand.Read(hw[3:])
		if err != nil {
			continue
		}

		mac := hw.String()

		// ensure we're not out of luck ...
		inUse, err := macInUse(mac, "")
		if err != nil || inUse {
			continue
		}

		return mac
	}
}

func broadcastAddr(p netip.Prefix) netip.Addr {
	if !p.Addr().Is4() {
		return netip.Addr{}
	}
	b := p.Addr().As4()
	ones := p.Bits()
	mask := ^uint32(0) >> ones
	val := binary.BigEndian.Uint32(b[:]) | mask
	binary.BigEndian.PutUint32(b[:], val)
	return netip.AddrFrom4(b)
}

func findFreeIP(s *Subnet) (string, error) {
	ips, err := findFreeIPs(s, 1)
	if err != nil {
		return "", err
	}
	return ips[0], nil
}

// findFreeIPs returns up to limit assignable IP addresses from the subnet,
// lowest first.
func findFreeIPs(s *Subnet, limit int) ([]string, error) {
	p, err := netip.ParsePrefix(s.CIDR)
	if err != nil {
		return nil, err
	}
	p = p.Masked() // a.b.c.d/mask => a.b.c.0/mask

	// broadcast and gateway addresses
	bcastAddr := broadcastAddr(p)
	gwAddr, _ := netip.ParseAddr(s.Gateway)

	// find all subnet IPs
	registeredIPs := s.FindIPs()
	registeredSet := make(map[string]struct{}, len(registeredIPs))
	for _, rip := range registeredIPs {
		registeredSet[rip] = struct{}{}
	}

	free := []string{}
	addr := p.Addr()
	addr = addr.Next() // skip the first IP from the range (.0)
	for p.Contains(addr) {
		// ensure it's not part of the subnet's reserved pool ranges or zone-local gateway pool ranges
		if s.IsInReservedPoolAddr(addr) || s.IsInGwPoolAddr(addr) {
			addr = addr.Next()
			continue
		}

		// ensure it's not gateway or broadcast address
		if addr == gwAddr || addr == bcastAddr {
			addr = addr.Next()
			continue
		}

		ip := addr.String()

		// ensure it's not already assigned to other adapters in the subnet
		if _, exists := registeredSet[ip]; exists {
			addr = addr.Next()
			continue
		}

		// seems unused, go for it
		free = append(free, ip)
		if len(free) >= limit {
			break
		}
		addr = addr.Next()
	}

	if len(free) == 0 {
		return nil, fmt.Errorf("no IP can be assigned")
	}
	return free, nil
}

func (a *Adapter) AssignIP() (string, error) {
	s, err := a.Subnet()
	if err != nil {
		return "", err
	}
	return findFreeIP(s)
}

// adapterIPRetryCandidates is the number of lowest free IP addresses a
// conflicting IP assignment randomly picks from on retry.
const adapterIPRetryCandidates = 32

func pickFreeIP(subnetId string, random bool) (string, error) {
	s, err := FindSubnetByID(subnetId)
	if err != nil {
		return "", err
	}
	if !random {
		return findFreeIP(s)
	}
	ips, err := findFreeIPs(s, adapterIPRetryCandidates)
	if err != nil {
		return "", err
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(ips))))
	if err != nil {
		return ips[0], nil
	}
	return ips[n.Int64()], nil
}

func verifyAdapterSettings(subnetId, selfId, mac string, addresses []string) error {
	s, err := FindSubnetByID(subnetId)
	if err != nil {
		return err
	}

	_, cidr, err := net.ParseCIDR(s.CIDR)
	if err != nil {
		return err
	}

	// ensure it's a correctly formatted IEEE 802 MAC-48 address
	_, err = net.ParseMAC(mac)
	if err != nil {
		return err
	}

	// ensure we don't have any MAC duplicates
	inUse, err := macInUse(mac, selfId)
	if err != nil {
		return err
	}
	if inUse {
		return fmt.Errorf("duplicated MAC address (%s) has been found. We can't authorize that unless you flag is as reserved", mac)
	}

	// ensure requested IPv4 addresses are not already assigned to other adapters
	if len(addresses) > 0 {
		used := map[string]struct{}{}
		for _, ip := range s.FindIPsExcept(selfId) {
			used[ip] = struct{}{}
		}
		for _, ip := range addresses {
			if _, ok := used[ip]; ok {
				return fmt.Errorf("%w: %s", ErrAdapterIPInUse, ip)
			}
		}
	}

	// ensure assigned IPv4 are in the associated subnet's CIDR range
	for _, i := range addresses {
		ip := net.ParseIP(i)
		if ip == nil {
			return fmt.Errorf("invalid IP: %s", i)
		}
		if !cidr.Contains(ip) {
			return fmt.Errorf("IP %s is not part of subnet's CIDR %s", i, cidr)
		}
	}

	return nil
}

var ErrAdapterIPInUse = errors.New("IP address is already assigned to another adapter")

// adapterIPAssignmentAttempts bounds automatic IP assignment retries, when
// concurrent requests happen to pick the same free IP address.
const adapterIPAssignmentAttempts = 10

func NewAdapter(subnetId, name, desc, mac string, addresses []string, reserved, autoAssign bool) (*Adapter, error) {
	if mac == "" {
		// no MAc address was provided, i.e. user requested one to be auto-generated
		mac = RandomAdapterMAC()
	}

	autoAssignIP := autoAssign && len(addresses) == 0
	for attempt := 1; ; attempt++ {
		// assign the first free IP address from the subnet if requested, or a
		// random one amongst the first free ones when retrying after a
		// conflict, so that concurrent requests don't keep colliding
		if autoAssignIP {
			ip, err := pickFreeIP(subnetId, attempt > 1)
			if err != nil {
				return nil, err
			}
			klog.Debugf("IP address %s is being assigned.", ip)
			addresses = []string{ip}
		}

		a, err := newAdapter(subnetId, name, desc, mac, addresses, reserved)
		ipTaken := errors.Is(err, ErrAdapterIPInUse) || mongo.IsDuplicateKeyError(err)
		if autoAssignIP && ipTaken && attempt < adapterIPAssignmentAttempts {
			// IP address got assigned to another adapter in the meantime
			klog.Debugf("IP address %s got concurrently assigned, retrying ...", addresses[0])
			continue
		}
		return a, err
	}
}

func newAdapter(subnetId, name, desc, mac string, addresses []string, reserved bool) (*Adapter, error) {
	// ensure the requested adapter is correctly flagged
	err := verifyAdapterSettings(subnetId, "", mac, addresses)
	if err != nil {
		klog.Errorf("verifyAdapterSettings: %s", err)
		return nil, err
	}

	a := Adapter{
		Resource:  NewResource(name, desc, MongoCollectionAdapterSchemaVersion),
		SubnetID:  subnetId,
		MAC:       mac,
		Addresses: addresses,
		Reserved:  reserved,
	}

	s, err := a.Subnet()
	if err != nil {
		return nil, err
	}

	// the database ensures an IP address can't be assigned twice in a subnet,
	// even by concurrent requests
	_, err = GetDB().Insert(MongoCollectionAdapterName, a)
	if err != nil {
		return nil, err
	}

	klog.Debugf("Created new adapter %s", a.String())

	// add adapter to subnet
	err = s.AddAdapter(a.String())
	if err != nil {
		return nil, err
	}

	return &a, nil
}

// macInUse tells whether a non-reserved adapter, other than selfId, already
// uses this MAC address.
func macInUse(mac, selfId string) (bool, error) {
	adapters, err := FindResourcesByKey[Adapter](MongoCollectionAdapterName, "mac", mac)
	if err != nil {
		return false, err
	}
	for _, a := range adapters {
		if a.String() != selfId && !a.Reserved {
			return true, nil
		}
	}
	return false, nil
}

func FindAdapters() []Adapter {
	return FindResources[Adapter](MongoCollectionAdapterName)
}

func FindAdaptersBySubnet(subnetId string) ([]Adapter, error) {
	return FindResourcesByKey[Adapter](MongoCollectionAdapterName, "subnet_id", subnetId)
}

func FindAdapterByID(id string) (*Adapter, error) {
	return FindResourceByID[Adapter](MongoCollectionAdapterName, id)
}

func FindAdapterByName(name string) (*Adapter, error) {
	return FindResourceByName[Adapter](MongoCollectionAdapterName, name)
}

func (a *Adapter) renameDbField(from, to string) error {
	return GetDB().Rename(MongoCollectionAdapterName, a.ID, from, to)
}

func (a *Adapter) setSchemaVersion(version int) error {
	return GetDB().SetSchemaVersion(MongoCollectionAdapterName, a.ID, version)
}

func (a *Adapter) migrateSchemaV2() error {
	err := a.renameDbField("subnet", "subnet_id")
	if err != nil {
		return err
	}

	err = a.setSchemaVersion(2)
	if err != nil {
		return err
	}

	return nil
}

func (a *Adapter) Subnet() (*Subnet, error) {
	return FindSubnetByID(a.SubnetID)
}

// Update applies a partial update: an empty MAC address and a nil (i.e.
// omitted) addresses list leave the current settings untouched.
func (a *Adapter) Update(name, desc, mac string, addresses []string, reserved bool) error {
	if mac == "" {
		mac = a.MAC
	}
	if addresses == nil {
		addresses = a.Addresses
	}

	// ensure the requested adapter is correctly flagged
	err := verifyAdapterSettings(a.SubnetID, a.String(), mac, addresses)
	if err != nil {
		klog.Error(err)
		return err
	}

	a.UpdateResourceDefaults(name, desc)
	a.MAC = mac
	a.Addresses = addresses
	a.Reserved = reserved
	return a.Save()
}

func (a *Adapter) Save() error {
	a.Updated()
	return resourceUpdate(MongoCollectionAdapterName, a.ID, a)
}

func (a *Adapter) Delete() error {
	klog.Debugf("Deleting adapter %s", a.String())

	if a.String() == ResourceUnknown {
		return nil
	}

	// remove adapter's reference from parent, if still around
	s, err := a.Subnet()
	if err == nil {
		err = s.RemoveAdapter(a.String())
	}
	if ignoreMissing(err) != nil {
		return err
	}

	return resourceDelete(MongoCollectionAdapterName, a.ID)
}

func (a *Adapter) Model() sdk.Adapter {
	return sdk.Adapter{
		Id:          a.String(),
		Name:        a.Name,
		Description: a.Description,
		Mac:         a.MAC,
		Addresses:   a.Addresses,
		Reserved:    ptr(a.Reserved),
	}
}
