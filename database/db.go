// Package db implements simple method of treatment to bbolt db.
package db

import (
	bolt "go.etcd.io/bbolt"
)

// DB implements interface access to bbolt db.
type DB struct {
	boltDB *bolt.DB
}

func (db *DB) BoltDB() *bolt.DB {
	return db.boltDB
}

// Open return pointer to DB,
// If DB does not exist then error.
func Open(path string) (*DB, error) {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

// Close implements access to close DB.
func (db *DB) Close() error {
	return db.boltDB.Close()
}

// Bucket returns pointer to Bucket in db,
// Returns error if name is blank, or name is too long.
func (db *DB) Bucket(name string, model Model) (*Bucket, error) {
	err := db.boltDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		return err
	})
	if err != nil {
		return nil, err
	}
	bucket := &Bucket{
		db:   db,
		name: name,
	}
	bucket.Objects = Manager{
		bucket: bucket,
		model:  model,
	}
	return bucket, nil
}

// ExistsBucket returns true if bucket exists.
func (db *DB) ExistsBucket(name string) bool {
	var exists bool
	db.boltDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(name))
		exists = (bucket != nil)
		return nil
	})
	return exists
}
