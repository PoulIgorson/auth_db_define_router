package db

import (
	"encoding/json"
	"fmt"

	. "github.com/PoulIgorson/sub_engine_fiber/database/errors"
	. "github.com/PoulIgorson/sub_engine_fiber/define"

	"github.com/PoulIgorson/sub_engine_fiber/logs"

	bolt "go.etcd.io/bbolt"
)

const DELETE = "DELETE"

// Bucket implements interface simple access to read/write in bbolt db.
type Bucket struct {
	db   *DB
	name string

	model   Model
	Objects Manager
}

// DB returns pointer to DB.
func (bucket *Bucket) DB() *DB {
	return bucket.db
}

// Name returns string, name of Bucket.
func (bucket *Bucket) Name() string {
	return bucket.name
}

// Model returns string, name of Bucket.
func (bucket *Bucket) Model() Model {
	return bucket.model
}

func (bucket *Bucket) Count() uint {
	var count string
	bucket.db.boltDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket.name))
		count = string(bucket.Get([]byte("0")))
		return nil
	})
	if count == "" || count == "0" {
		bucket.set(0, "1")
		count = "1"
	}
	return ParseUint(count) - 1
}

// Get implements getting value of key in bucket.
func (bucket *Bucket) Get(key uint) (Model, Error) {
	var value string
	err := bucket.db.boltDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket.name))
		value = string(bucket.Get([]byte(fmt.Sprint(key))))
		if value == "" {
			return fmt.Errorf("key `%v` is not exists", key)
		}
		return nil
	})
	if value == DELETE {
		err = NewErrValueDelete(key)
	}
	if err != nil {
		return nil, NewErrorf("Bucket.Get: %v", err.Error())
	}
	return bucket.model.Create(bucket.db, value), nil
}

// Set implements setting value of key in bucket.
func (bucket *Bucket) set(key uint, value string, save_bucket ...string) Error {
	if _, err := bucket.Get(key); err != nil && err.Name() == NewErrValueDelete(0).Name() {
		return NewErrorf("Bucket.Set: %v", err.Error())
	}
	err := bucket.db.boltDB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket.name))
		return bucket.Put([]byte(fmt.Sprint(key)), []byte(value))
	})
	if err != nil {
		logs.Error("Bucket.Set: Error of saving bucket `%v`: %v", bucket.Name(), err.Error())
	}
	return nil
}

// Delete implements Deleting value of key in bucket.
func (bucket *Bucket) Delete(key uint) Error {
	delete(bucket.Objects.objects, key)
	return bucket.set(key, DELETE)
}

// DeleteAll implements Deleting all values in bucket.
func (bucket *Bucket) DeleteAll() Error {
	err := bucket.db.BoltDB().Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket([]byte(bucket.name)); err != nil {
			return err
		}
		_, err := tx.CreateBucketIfNotExists([]byte(bucket.name))
		return err
	})
	if err != nil {
		return NewErrorf("Bucket.DeleteAll: %v", err.Error())
	}
	bucket.Objects.objects = map[uint]Model{}
	return nil
}

// SaveModel saving bucket in db
func SaveModel(bucket *Bucket, model Model) Error {
	if bucket == nil {
		return NewErrorf("SaveModel: %v", NewErrNilBucket().Error())
	}
	field_id, err := Check(model, "ID")
	if err != nil {
		return NewErrorf("SaveModel: %v", err.Error())
	}
	idUint := field_id.Interface().(uint)
	if _, err := bucket.Get(idUint); err != nil || idUint == 0 {
		next_id := bucket.Count() + 1
		field_id.SetUint(uint64(next_id))
		idUint = uint(next_id)
		bucket.set(0, fmt.Sprint(next_id+1))
	}
	buf, err := json.Marshal(model)
	if err != nil {
		return NewErrorf("SaveModel: %v", err.Error())
	}
	if idUint == 0 {
		return NewErrorf("SaveModel: internal error")
	}

	if _, ok := bucket.Objects.objects[idUint]; !ok {
		bucket.Objects.count++
	}
	bucket.Objects.objects[idUint] = model
	if bucket.Objects.maxId < idUint {
		bucket.Objects.maxId = idUint
	}
	if bucket.Objects.minId > idUint || bucket.Objects.minId == 0 {
		bucket.Objects.minId = idUint
	}
	return bucket.set(idUint, string(buf))
}
