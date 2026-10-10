/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"encoding/hex"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/kowabunga-cloud/common/klog"
	"github.com/kowabunga-cloud/kahuna/internal/sdk"
)

// isMissingResource tells whether a resource lookup error means that the
// resource doesn't exist (as opposed to e.g. the database being unavailable).
func isMissingResource(err error) bool {
	var invalidByte hex.InvalidByteError
	return errors.Is(err, mongo.ErrNoDocuments) || errors.Is(err, bson.ErrInvalidHex) || errors.As(err, &invalidByte)
}

// ignoreMissing returns nil if err means that a looked up resource doesn't
// exist, e.g. so that deleting an already deleted child isn't an error, and
// a partially failed deletion can be resumed.
func ignoreMissing(err error) error {
	if isMissingResource(err) {
		return nil
	}
	return err
}

// resourceUpdate writes obj to the database and invalidates its cache entry.
// It is the single authoritative place that couples a DB write with cache invalidation,
// keeping db.go free of cache knowledge.
func resourceUpdate(collection string, id bson.ObjectID, obj interface{}) error {
	_, err := GetDB().Update(collection, id, obj)
	if err != nil {
		return err
	}
	GetCache().Invalidate(collection, id.Hex())
	return nil
}

// resourceUpdateFields atomically applies an update document (e.g. $set,
// $addToSet, $pull) to a resource document, refreshing its update time, and
// invalidates its cache entry. Unlike resourceUpdate, which replaces the whole
// document, it can't overwrite concurrent changes made to other fields since
// the object was read.
func resourceUpdateFields(collection string, id bson.ObjectID, update bson.D) error {
	update = append(update, bson.E{Key: "$currentDate", Value: bson.D{bson.E{Key: "updated_at", Value: true}}})
	err := GetDB().UpdateFields(collection, id, update)
	if err != nil {
		return err
	}
	GetCache().Invalidate(collection, id.Hex())
	return nil
}

// resourceSetFields atomically sets the given fields of a resource document.
func resourceSetFields(collection string, id bson.ObjectID, fields bson.D) error {
	return resourceUpdateFields(collection, id, bson.D{bson.E{Key: "$set", Value: fields}})
}

// resourceAddChildRef atomically adds childId to a resource's child
// references array field, unless already present.
func resourceAddChildRef(collection string, id bson.ObjectID, field, childId string) error {
	return resourceUpdateChildRefAndCounters(collection, id, field, childId, true, nil)
}

// resourceRemoveChildRef atomically removes childId from a resource's child
// references array field.
func resourceRemoveChildRef(collection string, id bson.ObjectID, field, childId string) error {
	return resourceUpdateChildRefAndCounters(collection, id, field, childId, false, nil)
}

// counterExpr returns an aggregation expression adding delta to a numeric
// field, floored at 0 as counters are decoded into unsigned integers.
func counterExpr(field string, delta any) bson.D {
	return bson.D{bson.E{Key: "$max", Value: bson.A{
		0,
		bson.D{bson.E{Key: "$add", Value: bson.A{bson.D{bson.E{Key: "$ifNull", Value: bson.A{"$" + field, 0}}}, delta}}},
	}}}
}

// resourceApplyPipeline atomically applies a single $set aggregation stage to
// a resource document, refreshing its update time, and invalidates its cache
// entry.
func resourceApplyPipeline(collection string, id bson.ObjectID, set bson.D) error {
	set = append(set, bson.E{Key: "updated_at", Value: "$$NOW"})
	err := GetDB().UpdateFields(collection, id, bson.A{bson.D{bson.E{Key: "$set", Value: set}}})
	if err != nil {
		return err
	}
	GetCache().Invalidate(collection, id.Hex())
	return nil
}

// resourceIncCounters atomically adds (possibly negative) deltas to a
// resource's numeric fields, flooring results at 0.
func resourceIncCounters(collection string, id bson.ObjectID, deltas bson.D) error {
	set := bson.D{}
	for _, d := range deltas {
		set = append(set, bson.E{Key: d.Key, Value: counterExpr(d.Key, d.Value)})
	}
	return resourceApplyPipeline(collection, id, set)
}

// childRefExprs returns aggregation expressions telling whether adding (or
// removing) childId to (or from) an array field changes it, and the resulting
// array. Fields stored as null (nil Go slices) are handled as empty arrays,
// which $addToSet and $pull refuse to operate on.
func childRefExprs(field string, childId any, add bool) (changes, newRefs, refs bson.D) {
	refs = bson.D{bson.E{Key: "$ifNull", Value: bson.A{"$" + field, bson.A{}}}}
	present := bson.D{bson.E{Key: "$in", Value: bson.A{childId, refs}}}
	if add {
		changes = bson.D{bson.E{Key: "$not", Value: bson.A{present}}}
		newRefs = bson.D{bson.E{Key: "$concatArrays", Value: bson.A{refs, bson.A{childId}}}}
		return changes, newRefs, refs
	}
	changes = present
	newRefs = bson.D{bson.E{Key: "$filter", Value: bson.D{
		bson.E{Key: "input", Value: refs},
		bson.E{Key: "cond", Value: bson.D{bson.E{Key: "$ne", Value: bson.A{"$$this", childId}}}},
	}}}
	return changes, newRefs, refs
}

// childRefSetExpr returns the $set stage entry adding (or removing) childId
// to (or from) an array field.
func childRefSetExpr(field string, childId any, add bool) bson.E {
	changes, newRefs, refs := childRefExprs(field, childId, add)
	return bson.E{Key: field, Value: bson.D{bson.E{Key: "$cond", Value: bson.A{changes, newRefs, refs}}}}
}

// resourceUpdateChildRefAndCounters atomically adds (or removes) childId
// to (or from) a resource's child references array field and, only if that
// actually changed the references, applies the counter deltas (negated on
// removal, floored at 0). Adding an already referenced child or removing an
// unreferenced one is thus a no-op, and counters can't drift.
func resourceUpdateChildRefAndCounters(collection string, id bson.ObjectID, field, childId string, add bool, deltas bson.D) error {
	changes, _, _ := childRefExprs(field, childId, add)
	set := bson.D{childRefSetExpr(field, childId, add)}
	for _, d := range deltas {
		delta := d.Value
		if !add {
			delta = bson.D{bson.E{Key: "$multiply", Value: bson.A{d.Value, -1}}}
		}
		set = append(set, bson.E{Key: d.Key, Value: bson.D{bson.E{Key: "$cond", Value: bson.A{
			changes,
			counterExpr(d.Key, delta),
			bson.D{bson.E{Key: "$ifNull", Value: bson.A{"$" + d.Key, 0}}},
		}}}})
	}
	return resourceApplyPipeline(collection, id, set)
}

// resourceClaim atomically applies a $set aggregation stage to a resource
// document only if it still matches cond, and tells whether it did. It's the
// building block of check-and-set allocations (IDs, subnets ...), where
// concurrent requests must never get the same resource.
func resourceClaim(collection string, id bson.ObjectID, cond, set bson.D) (bool, error) {
	filter := append(bson.D{bson.E{Key: "_id", Value: id}}, cond...)
	set = append(set, bson.E{Key: "updated_at", Value: "$$NOW"})
	claimed, err := GetDB().UpdateFieldsIf(collection, filter, bson.A{bson.D{bson.E{Key: "$set", Value: set}}})
	if err != nil {
		return false, err
	}
	GetCache().Invalidate(collection, id.Hex())
	return claimed, nil
}

// resourceDelete removes a resource from the database and invalidates its cache entry,
// so that it can no longer be looked up by ID.
func resourceDelete(collection string, id bson.ObjectID) error {
	err := GetDB().Delete(collection, id)
	if err != nil {
		return err
	}
	GetCache().Invalidate(collection, id.Hex())
	return nil
}

const (
	CostCurrencyDefault = "EUR"
	CostErrorUnknown    = "Unknown Cost"

	ResourceUnknown = "000000000000000000000000"
)

type Resources struct {
	IDs []string
}

type Resource struct {
	ID            bson.ObjectID `bson:"_id"`
	CreatedAt     time.Time     `bson:"created_at"`
	UpdatedAt     time.Time     `bson:"updated_at"`
	SchemaVersion int           `bson:"schema_version"`

	Name        string `bson:"name"`
	Description string `bson:"description"`
}

func NewResource(name, desc string, schemaVersion int) Resource {
	now := time.Now()
	return Resource{
		ID:            bson.NewObjectIDFromTimestamp(now),
		CreatedAt:     now,
		UpdatedAt:     now,
		SchemaVersion: schemaVersion,
		Name:          name,
		Description:   desc,
	}
}

func (r *Resource) String() string {
	return r.ID.Hex()
}

func (r *Resource) Updated() {
	r.UpdatedAt = time.Now()
}

func (r *Resource) UpdateResourceDefaults(name, desc string) {
	SetFieldStr(&r.Name, name)
	SetFieldStr(&r.Description, desc)
}

func FindResources[T any](collection string) []T {
	var res []T
	err := GetDB().FindAll(collection, &res)
	if err != nil {
		klog.Error(err)
	}
	return res
}

func FindResourceIDs(collection string) []string {
	ids, err := GetDB().FindAllIDs(collection)
	if err != nil {
		klog.Error(err)
		return []string{}
	}
	return ids
}

func FindResourceByID[T any](collection, id string) (*T, error) {
	var res T
	err := GetDB().FindByID(collection, id, &res)
	return &res, err
}

func FindResourceByName[T any](collection, name string) (*T, error) {
	var res T
	err := GetDB().FindByName(collection, name, &res)
	return &res, err
}

func FindResourceByType[T any](collection, tp string) (*T, error) {
	var res T
	err := GetDB().FindByType(collection, tp, &res)
	return &res, err
}

func FindResourceByEmail[T any](collection, email string) (*T, error) {
	var res T
	err := GetDB().FindByEmail(collection, email, &res)
	return &res, err
}

func FindResourceByIP[T any](collection, ip string) (*T, error) {
	var res T
	err := GetDB().FindByIP(collection, ip, &res)
	return &res, err
}

func FindResourcesByKey[T any](collection, key, value string) ([]T, error) {
	var res []T
	err := GetDB().FindAllByKey(collection, key, value, &res)
	return res, err
}

func FindResourcesByFilter[T any](collection string, filter bson.D) ([]T, error) {
	var res []T
	err := GetDB().FindAllByFilter(collection, filter, &res)
	return res, err
}

// Cost

type CostStructure struct {
	Capacity int64
	Cost     ResourceCost
}

type ResourceCost struct {
	Price    float32 `bson:"cost,truncate"`
	Currency string  `bson:"currency"`
}

func NewResourceCost(price float32, currency string) ResourceCost {
	c := CostCurrencyDefault
	if currency != "" {
		c = currency
	}
	return ResourceCost{
		Price:    price,
		Currency: c,
	}
}

func (r *ResourceCost) Model() sdk.Cost {
	return sdk.Cost{
		Price:    r.Price,
		Currency: r.Currency,
	}
}

// Metadata

type ResourceMetadata struct {
	Key   string `bson:"key"`
	Value string `bson:"value"`
}

func NewResourceMetadata(k, v string) ResourceMetadata {
	return ResourceMetadata{
		Key:   k,
		Value: v,
	}
}

func (r *ResourceMetadata) Model() sdk.Metadata {
	return sdk.Metadata{
		Key:   r.Key,
		Value: r.Value,
	}
}
