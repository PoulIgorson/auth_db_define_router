// Package db implements simple method of treatment to bbolt db.
package bbolt

import (
	"reflect"

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
		db.boltDB = nil
		return nil
	}
	return NewErrorf(err.Error())
}

func GetNameBucket(model Model) string {
	typeName := ""
	if t := reflect.TypeOf(model); t.Kind() == reflect.Pointer {
		typeName = t.Elem().Name()
	} else {
		typeName = t.Name()
	}
	var name []rune
	for i, ch := range typeName {
		if i == 0 {
			if 'A' <= ch && ch <= 'Z' {
				ch += 0x20
			}
			name = append(name, ch)
			continue
		}
		if 'A' <= ch && ch <= 'Z' {
			if typeName[i-1] != '_' {
				name = append(name, '_')
			}
			name = append(name, ch+0x20)
			continue
		}
		name = append(name, ch)
	}
	return string(name)
}

// Table returns pointer to Bucket in db,
// Returns error if name is too long.
// name is not required
func (db *DataBase) Table(_ string, model Model) (Table, Error) {
	name := GetNameBucket(model)
	if db.buckets[name] != nil {
		return db.buckets[name], nil
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
	bucket.Objects = &Manager{
		bucket:  bucket,
		objects: map[uint]Model{},
	}
	for inc := uint(1); inc < bucket.Count()+1; inc++ {
		model, _ := bucket.Get(inc)
		if model == nil {
			continue
		}
		if bucket.Objects.maxId < inc {
			bucket.Objects.maxId = inc
		}
		if bucket.Objects.minId > inc || bucket.Objects.minId == 0 {
			bucket.Objects.minId = inc
		}
		bucket.Objects.objects[inc] = model
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
