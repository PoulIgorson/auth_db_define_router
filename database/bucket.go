package db

import (
	"encoding/json"
	"fmt"

	. "github.com/PoulIgorson/sub_engine_fiber/database/errors"
	. "github.com/PoulIgorson/sub_engine_fiber/define"

	bolt "go.etcd.io/bbolt"
)

const DELETE = "DELETE"
const sSAVE = "SAVE"

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
		bucket.Set(0, "1", sSAVE)
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
		return nil, NewErrorf("Bucket.Set: %v", err.Error())
	}
	return bucket.model.Create(bucket.db, value), nil
}

// Set implements setting value of key in bucket.
func (bucket *Bucket) Set(key uint, value string, save_bucket ...string) Error {
	if key == 0 && (len(save_bucket) == 0 || save_bucket[0] != sSAVE) {
		return NewErrorf("Bucket.Set: %v", NewErrValueNotAvaiable(0).Error())
	}
	if _, err := bucket.Get(key); err.Name() == NewErrValueDelete(0).Name() {
		return NewErrorf("Bucket.Set: %v", err.Error())
	}
	err := bucket.db.boltDB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket.name))
		return bucket.Put([]byte(fmt.Sprint(key)), []byte(value))
	})
	if err != nil {
		fmt.Printf("Bucket.Set: Error of saving bucket `%v`: %v", bucket.Name(), err.Error())
	}
	return NewErrorf("Bucket.Set: %v", err.Error())
}

// Delete implements Deleting value of key in bucket.
func (bucket *Bucket) Delete(key uint) Error {
	return bucket.Set(key, DELETE)
}

// SaveModel saving bucket in db
func SaveModel(bucket *Bucket, imodel interface{}) Error {
	if bucket == nil {
		return NewErrorf("SaveModel: %v", NewErrNilBucket().Error())
	}
	field_id, err := Check(imodel, "ID")
	if err != nil {
		return NewErrorf("SaveModel: %v", err.Error())
	}
	idInt := field_id.Interface().(uint)
	if _, err := bucket.Get(idInt); err != nil || idInt == 0 {
		next_id := bucket.Count()
		field_id.SetUint(uint64(next_id))
		idInt = uint(next_id)
		bucket.Set(0, fmt.Sprint(next_id+1), sSAVE)
	}
	buf, err := json.Marshal(imodel)
	if err != nil {
		return NewErrorf("SaveModel: %v", err.Error())
	}
	return bucket.Set(idInt, string(buf))
}

func (bucket *Bucket) GetAllModels() []Model {
	var resp []Model
	for inc := uint(1); inc < bucket.Count(); inc++ {
		model, err := bucket.Get(inc)
		if err != nil {
			continue
		}
		resp = append(resp, model)
	}
	return resp
}
