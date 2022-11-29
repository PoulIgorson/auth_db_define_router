// Package db implements simple method of treatment to bbolt db.
package db

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"

	. "auth_db_define_router/define"

	bolt "go.etcd.io/bbolt"
)

// itoa convet int to string.
var itoa = strconv.Itoa

// DB implements interface access to bbolt db.
type DB struct {
	db *bolt.DB
}

// Bucket implements interface simple access to read/write in bbolt db.
type Bucket struct {
	db   *bolt.DB
	name string
}

// DB returns pointer to bbolt.DB.
func (b *Bucket) DB() *bolt.DB {
	return b.db
}

// Name returns string, name of Bucket.
func (b *Bucket) Name() string {
	return b.name
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
func (this *DB) Bucket(name string) (*Bucket, error) {
	err := this.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		return err
	})
	if err != nil {
		return nil, err
	}
	return &Bucket{this.db, name}, nil
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

// Set implements setting value of key in bucket.
func (this *Bucket) Set(key int, value string) error {
	return this.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(this.name))
		return bucket.Put([]byte(itoa(key)), []byte(value))
	})
}

// Get implements getting value of key in bucket.
func (this *Bucket) Get(key int) (string, error) {
	var value string
	err := this.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(this.name))
		value = string(bucket.Get([]byte(itoa(key))))
		if value == "" {
			return fmt.Errorf("key `%v` is not exists", key)
		}
		return nil
	})
	return value, err
}

// GetOfField returns json-string of field in bucket.
func (this *Bucket) GetOfField(field string, value string) (string, error) {
	return this.GetOfFields([]string{field}, []string{value})
}

// GetOfFields returns json-string of fields in bucket.
func (this *Bucket) GetOfFields(fields []string, values []string) (string, error) {
	count := Min(len(fields), len(values))
	for inc := 1; true; inc++ {
		v, err := this.Get(inc)
		if err != nil {
			return "", err
		}

		var data fiber.Map
		err = json.Unmarshal([]byte(v), &data)
		if err != nil {
			return "ErrorJSON", err
		}

		finded := 0
		for i := 0; i < count; i++ {
			if data[fields[i]] == nil {
				continue
			}

			find_val := fmt.Sprint(data[fields[i]])
			if find_val == values[i] {
				finded++
			}
		}
		if finded == count {
			return v, nil
		}
	}
	return "", nil
}

// GetsOfField returns json-strings of field in bucket.
func (this *Bucket) GetsOfField(field string, value string) ([]string, error) {
	return this.GetsOfFields([]string{field}, []string{value})
}

// GetsOfFields returns json-strings of fields in bucket.
func (this *Bucket) GetsOfFields(fields []string, values []string) ([]string, error) {
	count := Min(len(fields), len(values))
	var resp []string
	for inc := 1; true; inc++ {
		v, err := this.Get(inc)
		if err != nil {
			break
		}

		var data fiber.Map
		err = json.Unmarshal([]byte(v), &data)
		if err != nil {
			break
		}

		finded := 0
		for i := 0; i < count; i++ {
			if data[fields[i]] == nil {
				continue
			}

			find_val := fmt.Sprint(data[fields[i]])
			if find_val == values[i] {
				finded++
			}
		}
		if finded == count {
			resp = append(resp, v)
		}
	}
	return resp, nil
}
