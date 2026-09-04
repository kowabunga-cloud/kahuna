/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"

	"github.com/kowabunga-cloud/common/klog"
)

var ErrDbNotConnected = errors.New("database not connected")

type KowabungaDB struct {
	Client *mongo.Client
	DB     *mongo.Database
}

type KowabungaDbEvent struct {
	DocumentKey KowabungaDocumentKey `bson:"documentKey"`
	Operation   string               `bson:"operationType"`
}
type KowabungaDocumentKey struct {
	ID bson.ObjectID `bson:"_id"`
}

// database singleton
var dbOnce sync.Once
var kDB *KowabungaDB

func GetDB() *KowabungaDB {
	dbOnce.Do(func() {
		klog.Debugf("Creating Kowabunga DB instance")
		kDB = &KowabungaDB{}
	})

	return kDB
}

func (db *KowabungaDB) Open(uri, database string) error {

	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetWriteConcern(writeconcern.Majority()))
	if err != nil {
		return err
	}
	db.Client = client

	// look for a primary server
	err = db.Client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		return err
	}

	db.DB = db.Client.Database(database)

	return nil
}

func (db *KowabungaDB) Admin() *mongo.Database {
	return db.Client.Database("admin")
}

func (db *KowabungaDB) Close() error {
	if db.Client != nil {
		return db.Client.Disconnect(context.TODO())
	}
	return nil
}

func (db *KowabungaDB) HasCollection(collection string) bool {
	coll, _ := db.DB.ListCollectionNames(context.Background(), bson.D{bson.E{Key: "name", Value: collection}})
	return len(coll) == 1
}

func (db *KowabungaDB) RenameCollection(from, to string) error {
	dbAdmin := db.Admin()
	if !db.HasCollection(from) {
		return nil
	}
	_from := fmt.Sprintf("%s.%s", db.DB.Name(), from)
	_to := fmt.Sprintf("%s.%s", db.DB.Name(), to)
	klog.Infof("Renaming MongoDB collection '%s' into '%s'", _from, _to)
	return dbAdmin.RunCommand(context.Background(), bson.D{bson.E{Key: "renameCollection", Value: _from}, bson.E{Key: "to", Value: _to}}).Err()
}

func (db *KowabungaDB) Insert(collection string, obj interface{}) (interface{}, error) {
	c := db.DB.Collection(collection)
	return c.InsertOne(context.TODO(), obj)
}

func (db *KowabungaDB) Update(collection string, id bson.ObjectID, obj interface{}) (interface{}, error) {
	c := db.DB.Collection(collection)
	return c.ReplaceOne(context.TODO(), bson.D{bson.E{Key: "_id", Value: id}}, obj)
}

func (db *KowabungaDB) Rename(collection string, id bson.ObjectID, from, to string) error {
	// cleanup cache data, if any
	defer func() {
		_ = GetCache().Delete(collection, id.Hex())
	}()

	c := db.DB.Collection(collection)

	// check if document has such a field
	result := c.FindOne(context.TODO(), bson.D{bson.E{Key: "_id", Value: id}, bson.E{Key: from, Value: bson.D{bson.E{Key: "$exists", Value: true}}}})
	if result.Err() == nil {
		// it does: rename field and update document
		klog.Debugf("Renaming document '%s' from '%s' field '%s' into '%s'", id.Hex(), collection, from, to)
		filter := bson.D{bson.E{Key: "_id", Value: id}}
		update := bson.D{bson.E{Key: "$rename", Value: bson.D{bson.E{Key: from, Value: to}}}}
		_, err := c.UpdateOne(context.TODO(), filter, update)
		return err
	}

	return nil
}

func (db *KowabungaDB) SetSchemaVersion(collection string, id bson.ObjectID, schemaVersion int) error {
	// cleanup cache data, if any
	defer func() {
		_ = GetCache().Delete(collection, id.Hex())
	}()

	c := db.DB.Collection(collection)
	filter := bson.D{bson.E{Key: "_id", Value: id}}
	update := bson.D{bson.E{Key: "$set", Value: bson.D{bson.E{Key: "schema_version", Value: schemaVersion}}}}

	// check if document has such a schemaVersion
	result1 := c.FindOne(context.TODO(), bson.D{bson.E{Key: "_id", Value: id}, bson.E{Key: "schema_version", Value: bson.D{bson.E{Key: "$exists", Value: false}}}})
	if result1.Err() == nil {
		// it does not: adds initial schema version
		klog.Debugf("Updating document '%s' from '%s', initializing 'schemaVersion' field to '%d'", id.Hex(), collection, schemaVersion)
		_, err := c.UpdateOne(context.TODO(), filter, update)
		return err
	}

	// check if document has outdated schemaVersion
	result2 := c.FindOne(context.TODO(), bson.D{bson.E{Key: "_id", Value: id}, bson.E{Key: "schema_version", Value: bson.D{bson.E{Key: "$ne", Value: schemaVersion}}}})
	if result2.Err() == nil {
		// it does: upates schema version
		klog.Debugf("Updating document '%s' from '%s', setting 'schemaVersion' field to '%d'", id.Hex(), collection, schemaVersion)
		_, err := c.UpdateOne(context.TODO(), filter, update)
		return err
	}

	return nil
}

func (db *KowabungaDB) FindAll(collection string, results interface{}) error {
	if db.DB == nil {
		return ErrDbNotConnected
	}
	c := db.DB.Collection(collection)
	cursor, err := c.Find(context.TODO(), bson.D{})
	if err != nil {
		return err
	}

	return cursor.All(context.TODO(), results)
}

func (db *KowabungaDB) FindAllByKey(collection, key, value string, results interface{}) error {
	if db.DB == nil {
		return ErrDbNotConnected
	}
	c := db.DB.Collection(collection)
	cursor, err := c.Find(context.TODO(), bson.D{bson.E{Key: key, Value: value}})
	if err != nil {
		return err
	}

	return cursor.All(context.TODO(), results)
}

func (db *KowabungaDB) Find(collection, k, v string, result interface{}) error {
	if db.DB == nil {
		return ErrDbNotConnected
	}
	c := db.DB.Collection(collection)
	return c.FindOne(context.TODO(), bson.D{bson.E{Key: k, Value: v}}, nil).Decode(result)
}

func (db *KowabungaDB) FindByArrayContains(collection, k, v string, result interface{}) error {
	if db.DB == nil {
		return ErrDbNotConnected
	}
	c := db.DB.Collection(collection)
	return c.FindOne(context.TODO(), bson.D{bson.E{Key: k, Value: "{$all: [" + v + "]}"}}, nil).Decode(result)
}

func (db *KowabungaDB) FindByID(collection, id string, result interface{}) error {
	// look into cache first
	err := GetCache().Get(collection, id, result)
	if err == nil {
		return nil
	}

	if db.DB == nil {
		return ErrDbNotConnected
	}

	// failover: look into DB
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	c := db.DB.Collection(collection)
	err = c.FindOne(context.TODO(), bson.D{bson.E{Key: "_id", Value: oid}}, nil).Decode(result)
	if err == nil {
		// cache back data
		GetCache().Set(collection, id, result)
	}

	return err
}

func (db *KowabungaDB) FindByName(collection, name string, result interface{}) error {
	return db.Find(collection, "name", name, result)
}

func (db *KowabungaDB) FindByType(collection, tp string, result interface{}) error {
	return db.Find(collection, "type", tp, result)
}

func (db *KowabungaDB) FindByEmail(collection, email string, result interface{}) error {
	return db.Find(collection, "email", email, result)
}

func (db *KowabungaDB) FindByIP(collection, ip string, result interface{}) error {
	return db.Find(collection, "local_ip", ip, result)
}

func (db *KowabungaDB) Delete(collection string, id bson.ObjectID) error {
	c := db.DB.Collection(collection)
	_, err := c.DeleteOne(context.TODO(), bson.D{bson.E{Key: "_id", Value: id}})
	return err
}

type resourceIDOnly struct {
	ID bson.ObjectID `bson:"_id"`
}

func (db *KowabungaDB) FindAllIDs(collection string) ([]string, error) {
	if db.DB == nil {
		return nil, ErrDbNotConnected
	}
	c := db.DB.Collection(collection)
	opts := options.Find().SetProjection(bson.D{bson.E{Key: "_id", Value: 1}})
	cursor, err := c.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		return nil, err
	}

	var docs []resourceIDOnly
	if err := cursor.All(context.TODO(), &docs); err != nil {
		return nil, err
	}

	ids := make([]string, len(docs))
	for i, d := range docs {
		ids[i] = d.ID.Hex()
	}
	return ids, nil
}

func (db *KowabungaDB) EnsureIndexes() error {
	if db.DB == nil {
		return ErrDbNotConnected
	}

	collectionIndexes := map[string][]mongo.IndexModel{
		MongoCollectionInstanceName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "project_id", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "kaktus_id", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "local_ip", Value: 1}}},
		},
		MongoCollectionVolumeName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "project_id", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "pool_id", Value: 1}}},
		},
		MongoCollectionSubnetName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "vnet_id", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "project_id", Value: 1}}},
		},
		MongoCollectionAdapterName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "subnet_id", Value: 1}}},
		},
		MongoCollectionTokenName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "parent_type", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "agent_id", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "user_id", Value: 1}}},
		},
		MongoCollectionUserName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "email", Value: 1}}},
		},
		MongoCollectionProjectName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
		},
		MongoCollectionRegionName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
		},
		MongoCollectionZoneName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "region_id", Value: 1}}},
		},
		MongoCollectionKaktusName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "zone_id", Value: 1}}},
		},
		MongoCollectionStoragePoolName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "region_id", Value: 1}}},
		},
		MongoCollectionKomputeName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "project_id", Value: 1}}},
		},
		MongoCollectionKawaiiName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "project_id", Value: 1}}},
		},
		MongoCollectionKyloName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "project_id", Value: 1}}},
		},
		MongoCollectionKonveyName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "project_id", Value: 1}}},
		},
		MongoCollectionNfsName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "region_id", Value: 1}}},
		},
		MongoCollectionKiwiName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "region_id", Value: 1}}},
		},
		MongoCollectionVNetName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
			{Keys: bson.D{bson.E{Key: "region_id", Value: 1}}},
		},
		MongoCollectionTemplateName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
		},
		MongoCollectionTeamName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
		},
		MongoCollectionAgentName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
		},
		MongoCollectionDnsRecordName: {
			{Keys: bson.D{bson.E{Key: "name", Value: 1}}},
		},
	}

	for collName, indexes := range collectionIndexes {
		coll := db.DB.Collection(collName)
		_, err := coll.Indexes().CreateMany(context.Background(), indexes)
		if err != nil {
			klog.Errorf("Failed to ensure indexes for collection %s: %v", collName, err)
		}
	}

	return nil
}
