// Package db implements simple method of treatment to bbolt db.
package db

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"

	. "github.com/PoulIgorson/sub_engine_fiber/define"

	bolt "go.etcd.io/bbolt"
)

// itoa convet int to string.
var itoa = strconv.Itoa

const DELETE = "DELETE"
const sSAVE_BUCKET = "SAVE_BUCKET"

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

// SaveBucket saving bucket in db
func SaveBucket(bucket *Bucket, imodel interface{}) error {
	if bucket == nil {
		return fmt.Errorf("SaveBucket: bucket is nil")
	}
	field_id, err := Check(imodel, "ID")
	if err != nil {
		return err
	}
	idInt := field_id.Interface().(uint)
	if _, err := bucket.Get(int(idInt)); err != nil || idInt == 0 {
		k, _ := bucket.Get(0)
		next_id := Atoi(k)
		if next_id == 0 {
			next_id++
		}
		field_id.SetUint(uint64(next_id))
		idInt = uint(next_id)
		bucket.Set(0, Itoa(next_id+1), sSAVE_BUCKET)
	}
	buf, err := json.Marshal(imodel)
	if err != nil {
		return err
	}
	return bucket.Set(int(idInt), string(buf))
}

func (this *Bucket) Count() uint {
	count, _ := this.Get(0)
	if count == "" || count == "0" {
		this.Set(0, "1", sSAVE_BUCKET)
		count = "1"
	}
	return ParseUint(count) - 1
}

// Set implements setting value of key in bucket.
func (this *Bucket) Set(key int, value string, save_bucket ...string) error {
	if key == 0 && (len(save_bucket) == 0 || save_bucket[0] != sSAVE_BUCKET) {
		return fmt.Errorf("Bucket.Set: key `%v` is not available", 0)
	}
	if rvalue, _ := this.Get(key); rvalue == DELETE {
		return fmt.Errorf("Bucket.Set: value of key `%v` is delete", key)
	}
	err := this.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(this.name))
		return bucket.Put([]byte(itoa(key)), []byte(value))
	})
	if err != nil {
		fmt.Printf("Bucket.Set: Error of saving bucket `%v`: %v", this.Name(), err.Error())
	}
	return err
}

// Delete implements Deleting value of key in bucket.
func (this *Bucket) Delete(key int) error {
	return this.Set(key, DELETE)
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
	if value == DELETE {
		err = fmt.Errorf("Bucket.Set: value of key `%v` is delete", key)
	}
	return value, err
}

func (this *Bucket) GetAllStr() ([]string, error) {
	var resp []string
	var v string
	var err error
	for inc := 1; inc < int(this.Count()); inc++ {
		v, err = this.Get(inc)
		if v == DELETE {
			continue
		}
		if err != nil {
			break
		}

		var data fiber.Map
		err = json.Unmarshal([]byte(v), &data)
		if err != nil {
			break
		}

		resp = append(resp, v)
	}
	return resp, err
}

// GetOfField returns json-string of field in bucket.
func (this *Bucket) GetOfField(field string, value string) (string, error) {
	return this.GetOfFields([]string{field}, []string{value})
}

// GetOfFields returns json-string of fields in bucket.
func (this *Bucket) GetOfFields(fields []string, values []string) (string, error) {
	count := Min(len(fields), len(values))
	maxInd := int(this.Count()) + 1
	for inc := 1; inc < maxInd; inc++ {
		v, err := this.Get(inc)
		if err != nil || v == DELETE {
			continue
		}

		var data fiber.Map
		json.Unmarshal([]byte(v), &data)

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
	return "", fmt.Errorf("DB: index `%v` does not exists", maxInd)
}

// GetsOfField returns all json-strings of field in bucket.
func (this *Bucket) GetsOfField(field string, value string) ([]string, error) {
	return this.GetsOfFields([]string{field}, []string{value})
}

// GetsOfFields returns all json-strings of fields in bucket.
func (this *Bucket) GetsOfFields(fields []string, values []string) ([]string, error) {
	count := Min(len(fields), len(values))
	var resp []string
	maxInd := int(this.Count()) + 1
	var v string
	var err error
	for inc := 1; inc < maxInd; inc++ {
		v, err = this.Get(inc)
		if v == DELETE {
			continue
		}
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
	return resp, err
}
