// Package db implements simple method of treatment to bbolt db.
package bbolt

import (
	bolt "go.etcd.io/bbolt"

	. "github.com/PoulIgorson/sub_engine_fiber/database/errors"
	. "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
)

// DataBase implements interface access to bbolt db.
type DataBase struct {
	boltDB  *bolt.DB
	buckets map[string]*Bucket
}

func (db *DataBase) BoltDB() *bolt.DB {
	return db.boltDB
}

// Open return pointer to DataBase,
// If DataBase does not exist then error.
func Open(path string) (*DataBase, Error) {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, NewErrorf(err.Error())
	}
	return &DataBase{db, map[string]*Bucket{}}, nil
}

// Close implements access to close DataBase.
func (db *DataBase) Close() Error {
	db.buckets = nil
	err := db.boltDB.Close()
	if err == nil {
		return nil
	}
	return NewErrorf(err.Error())
}

// Table returns pointer to Bucket in db,
// Returns error if name is blank, or name is too long.
func (db *DataBase) Table(name string, model Model) (Table, Error) {
	if db.buckets[name] != nil {
		return db.buckets[name], nil
	}
	_, ok := model.Id().(uint)
	if !ok && name != "user" {
		_, ok = model.Id().(float64)
		if !ok {
			return nil, NewErrorf("bbolt: id must be uint")
		}
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
	bucket.Objects = &Manager{
		bucket:  bucket,
		objects: map[uint]Model{},
	}
	db.buckets[name] = bucket
	return bucket, nil
}

// ExistsBucket returns true if bucket exists.
func (db *DataBase) ExistsTable(name string) bool {
	if db.buckets[name] != nil {
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
