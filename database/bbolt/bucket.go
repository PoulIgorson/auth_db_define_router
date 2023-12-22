package bbolt

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	bolt "go.etcd.io/bbolt"

	. "github.com/PoulIgorson/sub_engine_fiber/database/define"
	. "github.com/PoulIgorson/sub_engine_fiber/database/errors"
	. "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
	. "github.com/PoulIgorson/sub_engine_fiber/define"
)

var _ Table = &Bucket{}

const _DELETE = "DELETE"

func checkId(idI any) (uint, error) {
	id, ok := idI.(uint)
	if !ok {
		keyF, ok := idI.(float64)
		if !ok {
			return 0, NewErrorf("bbolt: key must be uint")
		}
		id = uint(int(keyF))
	}
	return id, nil
}

// Bucket implements interface simple access to read/write in bbolt db.
type Bucket struct {
	db   *DataBase
	name string

	model   Model
	Objects ManagerI
}

func (bucket *Bucket) Manager() ManagerI {
	return bucket.Objects
}

func (bucket *Bucket) SetManager(newManager ManagerI) {
	bucket.Objects = newManager
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
func (bucket *Bucket) Get(keyI any) (Model, error) {
	key, err := checkId(keyI)
	if err != nil {
		return nil, err
	}
	var value string
	err = bucket.db.boltDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket.name))
		value = string(bucket.Get([]byte(fmt.Sprint(key))))
		if value == "" {
			return fmt.Errorf("bbolt: key `%v` is not exists", key)
		}
		return nil
	})
	if value == _DELETE {
		err = NewErrValueDelete(key)
	}
	if err != nil {
		return nil, NewErrorf("bbolt: Bucket.Get: %v", err.Error())
	}
	return bucket.model.Create(bucket.db, value), nil
}

// Set implements setting value of key in bucket.
func (bucket *Bucket) set(keyI any, value string) error {
	key, err := checkId(keyI)
	if err != nil {
		return err
	}
	if _, err := bucket.Get(key); err != nil && err.(Error).Name() == NewErrValueDelete(0).Name() {
		return NewErrorf("bbolt: Bucket.Set: %v", err.Error())
	}
	err = bucket.db.boltDB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket.name))
		if value == _DELETE {
			return bucket.Delete([]byte(fmt.Sprint(key)))
		}
		return bucket.Put([]byte(fmt.Sprint(key)), []byte(value))
	})
	if err != nil {
		log.Printf("bbolt: Bucket.Set: Error of saving bucket `%v`: %v\n", bucket.Name(), err.Error())
	}
	return nil
}

// Delete implements Deleting value of key in bucket.
func (bucket *Bucket) Delete(keyI any) error {
	key, err := checkId(keyI)
	if err != nil {
		return err
	}
	bucket.Objects.ClearId(key)
	return bucket.set(key, _DELETE)
}

// DeleteAll implements Deleting all values in bucket.
func (bucket *Bucket) DeleteAll() error {
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
	bucket.Objects.Clear()
	return nil
}

func (bucket *Bucket) Save(model Model) error {
	field_id, err := Check(model, "ID")
	if err != nil {
		return NewErrorf("bbolt: " + err.Error())
	}
	idUint, err := checkId(model.Id())
	if err != nil {
		if GetNameModel(model) != "user" {
			return err
		}
		field_id.Set(reflect.ValueOf(uint(0)))
		idUint, err = 0, nil
	}

	if idUint == 0 {
		next_id := bucket.Count() + 1
		field_id.Set(reflect.ValueOf(uint(next_id)))
		idUint = next_id
		bucket.set(uint(0), fmt.Sprint(next_id+1))
	} else if _, err := bucket.Get(idUint); err != nil {
		return err
	}

	buf, err := json.Marshal(model)
	if err != nil {
		return NewErrorf("bbolt: " + err.Error())
	}

	if idUint == 0 {
		next_id := bucket.Count() + 1
		field_id.Set(reflect.ValueOf(uint(next_id)))
		idUint = uint(next_id)
		bucket.set(uint(0), fmt.Sprint(next_id+1))
		if idUint == 0 {
			return NewErrorf("bbolt: internal error")
		}
		buf, err = json.Marshal(model)
		if err != nil {
			return NewErrorf("bbolt: " + err.Error())
		}
	}

	errr := bucket.set(idUint, string(buf))
	if errr != nil {
		return errr
	}

	model, _ = bucket.Get(idUint)

	bucket.Objects.Store(idUint, model)

	return nil
}
