// Package db implements simple method of treatment to bbolt db.
package bbolt

import (
	"sync"

	bolt "go.etcd.io/bbolt"

	"github.com/PoulIgorson/sub_engine_fiber/database/base"
	. "github.com/PoulIgorson/sub_engine_fiber/database/define"
	. "github.com/PoulIgorson/sub_engine_fiber/database/errors"
	. "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
)

var _ DB = &DataBase{}

type bucketMap struct {
	sync.Map
}

func (m *bucketMap) Load(key string) *Bucket {
	v, ok := m.Map.Load(key)
	if !ok {
		return nil
	}
	return v.(*Bucket)
}

func (m *bucketMap) LoadOK(key string) (*Bucket, bool) {
	v, ok := m.Map.Load(key)
	if !ok {
		return nil, false
	}
	return v.(*Bucket), true
}

// DataBase implements interface access to bbolt db.
type DataBase struct {
	boltDB  *bolt.DB
	buckets bucketMap // map[string]Table
}

func (db *DataBase) BoltDB() *bolt.DB {
	return db.boltDB
}

// Open return pointer to DataBase,
// If DataBase does not exist then error.
func Open(path string) (*DataBase, error) {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, NewErrorf(err.Error())
	}
	return &DataBase{boltDB: db}, nil
}

// Close implements access to close DataBase.
func (db *DataBase) Close() error {
	db.buckets = bucketMap{}
	err := db.boltDB.Close()
	if err == nil {
		db.boltDB = nil
		return nil
	}
	return NewErrorf(err.Error())
}

func (db *DataBase) TableFromCache(name string) Table {
	bucket, ok := db.buckets.LoadOK(name)
	if !ok || bucket == nil {
		return nil
	}
	return bucket
}

// Table returns pointer to Bucket in db,
// Returns error if name is too long.
// name is not required
func (db *DataBase) Table(_ string, model Model) (Table, error) {
	name := GetNameModel(model)
	if bucket := db.buckets.Load(name); bucket != nil {
		return bucket, nil
	}
	_, ok := model.Id().(uint)
	if !ok && name != "user" {
		return nil, NewErrorf("bbolt: id must be uint")
	}
	err := db.boltDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		return err
	})
	if err != nil {
		return nil, NewErrorf(err.Error())
	}
	bucket := &Bucket{
		db:    db,
		name:  name,
		model: model,
	}
	bucket.Objects = base.NewManager(bucket)
	db.buckets.Store(name, bucket)
	now := uint(0)
	count := bucket.Count()
	bucket.Objects.Broadcast(func() any {
		now++
		if now >= count {
			return nil
		}
		return now
	})
	return bucket, nil
}

// ExistsBucket returns true if bucket exists.
func (db *DataBase) ExistsTable(name string) bool {
	if _, ok := db.buckets.LoadOK(name); ok {
		return true
	}
	var exists bool
	db.boltDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(name))
		exists = (bucket != nil)
		return nil
	})
	return exists
}
