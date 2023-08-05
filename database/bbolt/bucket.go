package bbolt

import (
	"encoding/json"
	"fmt"

	. "github.com/PoulIgorson/sub_engine_fiber/database/errors"
	. "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
	. "github.com/PoulIgorson/sub_engine_fiber/define"

	"github.com/PoulIgorson/sub_engine_fiber/logs"

	bolt "go.etcd.io/bbolt"
)

const DELETE = "DELETE"

// Bucket implements interface simple access to read/write in bbolt db.
type Bucket struct {
	db   *DataBase
	name string

	model   Model
	Objects *Manager
}

func (bucket *Bucket) Manager() ManagerI {
	return bucket.Objects
}

// DB returns pointer to DB.
func (bucket *Bucket) DB() DB {
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
		bucket.set(uint(0), "1")
		count = "1"
	}
	return ParseUint(count) - 1
}

// Get implements getting value of key in bucket.
func (bucket *Bucket) Get(keyI any) (Model, Error) {
	key, ok := keyI.(uint)
	if !ok {
		keyF, ok := keyI.(float64)
		if !ok {
			return nil, NewErrorf("bbolt: key must be uint")
		}
		key = uint(int(keyF))
	}
	var value string
	err := bucket.db.boltDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket.name))
		value = string(bucket.Get([]byte(fmt.Sprint(key))))
		if value == "" {
			return fmt.Errorf("bbolt: key `%v` is not exists", key)
		}
		return nil
	})
	if value == DELETE {
		err = NewErrValueDelete(key)
	}
	if err != nil {
		return nil, NewErrorf("bbolt: Bucket.Get: %v", err.Error())
	}
	return bucket.model.Create(bucket.db, value), nil
}

// Set implements setting value of key in bucket.
func (bucket *Bucket) set(keyI any, value string) Error {
	key, ok := keyI.(uint)
	if !ok {
		keyF, ok := keyI.(float64)
		if !ok {
			return NewErrorf("bbolt: key must be uint")
		}
		key = uint(int(keyF))
	}
	if _, err := bucket.Get(key); err != nil && err.Name() == NewErrValueDelete(0).Name() {
		return NewErrorf("bbolt: Bucket.Set: %v", err.Error())
	}
	err := bucket.db.boltDB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket.name))
		return bucket.Put([]byte(fmt.Sprint(key)), []byte(value))
	})
	if err != nil {
		logs.Error("bbolt: Bucket.Set: Error of saving bucket `%v`: %v", bucket.Name(), err.Error())
	}
	return nil
}

// Delete implements Deleting value of key in bucket.
func (bucket *Bucket) Delete(keyI any) Error {
	key, ok := keyI.(uint)
	if !ok {
		keyF, ok := keyI.(float64)
		if !ok {
			return NewErrorf("bbolt: key must be uint")
		}
		key = uint(int(keyF))
	}
	for bucket.Objects.rwObjects {
	}
	bucket.Objects.rwObjects = true
	delete(bucket.Objects.objects, key)
	bucket.Objects.rwObjects = false
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
		return NewErrorf("bbolt: Bucket.DeleteAll: %v", err.Error())
	}
	for bucket.Objects.rwObjects {
	}
	bucket.Objects.rwObjects = true
	bucket.Objects.objects = map[uint]Model{}
	bucket.Objects.rwObjects = false
	return nil
}

func (bucket *Bucket) Save(model Model) Error {
	return SaveModel(bucket, model)
}

// SaveModel saving bucket in db
func SaveModel(bucket *Bucket, model Model) Error {
	if bucket == nil {
		return NewErrorf("bbolt: SaveModel: %v", NewErrNilBucket().Error())
	}
	field_id, err := Check(model, "ID")
	if err != nil {
		return NewErrorf("bbolt: SaveModel: %v", err.Error())
	}
	idUint, ok := field_id.Interface().(uint)
	if !ok {
		idF, ok := field_id.Interface().(float64)
		if !ok {
			return NewErrorf("bbolt: key must be uint")
		}
		idUint = uint(int(idF))
	}
	if _, err := bucket.Get(idUint); err != nil || idUint == 0 {
		next_id := bucket.Count() + 1
		field_id.SetUint(uint64(next_id))
		idUint = uint(next_id)
		bucket.set(uint(0), fmt.Sprint(next_id+1))
	}
	buf, err := json.Marshal(model)
	if err != nil {
		return NewErrorf("bbolt: SaveModel: %v", err.Error())
	}
	if idUint == 0 {
		next_id := bucket.Count() + 1
		field_id.SetUint(uint64(next_id))
		idUint = uint(next_id)
		bucket.set(uint(0), fmt.Sprint(next_id+1))
		if idUint == 0 {
			return NewErrorf("bbolt: SaveModel: internal error")
		}
		buf, err = json.Marshal(model)
		if err != nil {
			return NewErrorf("bbolt: SaveModel: %v", err.Error())
		}
	}

	for bucket.Objects.rwObjects {
	}
	bucket.Objects.rwObjects = true
	if _, ok := bucket.Objects.objects[idUint]; !ok {
		bucket.Objects.count++
	}
	bucket.Objects.rwObjects = false

	errr := bucket.set(idUint, string(buf))

	if err != nil {
		return errr
	}

	model, _ = bucket.Get(idUint)

	for bucket.Objects.rwObjects {
	}
	bucket.Objects.rwObjects = true
	bucket.Objects.objects[idUint] = model
	bucket.Objects.rwObjects = false

	if bucket.Objects.maxId < idUint {
		bucket.Objects.maxId = idUint
	}
	if bucket.Objects.minId > idUint || bucket.Objects.minId == 0 {
		bucket.Objects.minId = idUint
	}
	return nil
}
