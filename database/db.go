// Package db implements simple method of treatment to bbolt db.
package db

import (
	bolt "go.etcd.io/bbolt"
)

// DB implements interface access to bbolt db.
type DB struct {
	db *bolt.DB
}

func (db *DB) BoltDB() *bolt.DB {
	return db.db
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
func (this *DB) Close() error {
	return this.db.Close()
}

// Bucket returns pointer to Bucket in db,
// Returns error if name is blank, or name is too long.
func (this *DB) Bucket(name string, model Model) (*Bucket, error) {
	err := this.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		return err
	})
	if err != nil {
		return nil, err
	}
	bucket := &Bucket{
		db:   this,
		name: name,
	}
	bucket.Objects = Manager{
		bucket: bucket,
		model:  model,
	}
	return bucket, nil
}

// ExistsBucket returns true if bucket exists.
func (this *DB) ExistsBucket(name string) bool {
	var exists bool
	this.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(name))
		exists = (bucket != nil)
		return nil
	})
	return exists
}
