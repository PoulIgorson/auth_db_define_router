package db

import (
	"encoding/json"
	"fmt"

	. "github.com/PoulIgorson/sub_engine_fiber/define"

	bolt "go.etcd.io/bbolt"
)

const DELETE = "DELETE"
const sSAVE = "SAVE"

// Bucket implements interface simple access to read/write in bbolt db.
type Bucket struct {
	db   *DB
	name string

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
		bucket.Set(0, Itoa(next_id+1), sSAVE)
	}
	buf, err := json.Marshal(imodel)
	if err != nil {
		return err
	}
	return bucket.Set(int(idInt), string(buf))
}

func (bucket *Bucket) Count() uint {
	count, _ := bucket.Get(0)
	if count == "" || count == "0" {
		bucket.Set(0, "1", sSAVE)
		count = "1"
	}
	return ParseUint(count) - 1
}

// Set implements setting value of key in bucket.
func (bucket *Bucket) Set(key int, value string, save_bucket ...string) error {
	if key == 0 && (len(save_bucket) == 0 || save_bucket[0] != sSAVE) {
		return fmt.Errorf("Bucket.Set: key `%v` is not available", 0)
	}
	if rvalue, _ := bucket.Get(key); rvalue == DELETE {
		return fmt.Errorf("Bucket.Set: value of key `%v` is delete", key)
	}
	err := bucket.db.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket.name))
		return bucket.Put([]byte(Itoa(key)), []byte(value))
	})
	if err != nil {
		fmt.Printf("Bucket.Set: Error of saving bucket `%v`: %v", bucket.Name(), err.Error())
	}
	return err
}

// Delete implements Deleting value of key in bucket.
func (bucket *Bucket) Delete(key int) error {
	return bucket.Set(key, DELETE)
}

// Get implements getting value of key in bucket.
func (bucket *Bucket) Get(key int) (string, error) {
	var value string
	err := bucket.db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket.name))
		value = string(bucket.Get([]byte(Itoa(key))))
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

func (bucket *Bucket) GetAllStr() ([]string, error) {
	var resp []string
	var v string
	var err error
	for inc := 1; inc < int(bucket.Count()); inc++ {
		v, err = bucket.Get(inc)
		if v == DELETE {
			continue
		}
		if err != nil {
			break
		}

		var data map[string]any
		err = json.Unmarshal([]byte(v), &data)
		if err != nil {
			break
		}

		resp = append(resp, v)
	}
	return resp, err
}

// GetOfField returns json-string of field in bucket.
func (bucket *Bucket) GetOfField(field string, value string) (string, error) {
	return bucket.GetOfFields([]string{field}, []string{value})
}

// GetOfFields returns json-string of fields in bucket.
func (bucket *Bucket) GetOfFields(fields []string, values []string) (string, error) {
	count := Min(len(fields), len(values))
	maxInd := int(bucket.Count()) + 1
	for inc := 1; inc < maxInd; inc++ {
		v, err := bucket.Get(inc)
		if err != nil || v == DELETE {
			continue
		}

		var data map[string]any
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
func (bucket *Bucket) GetsOfField(field string, value string) ([]string, error) {
	return bucket.GetsOfFields([]string{field}, []string{value})
}

// GetsOfFields returns all json-strings of fields in bucket.
func (bucket *Bucket) GetsOfFields(fields []string, values []string) ([]string, error) {
	count := Min(len(fields), len(values))
	var resp []string
	maxInd := int(bucket.Count()) + 1
	var v string
	var err error
	for inc := 1; inc < maxInd; inc++ {
		v, err = bucket.Get(inc)
		if v == DELETE {
			continue
		}
		if err != nil {
			break
		}

		var data map[string]any
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

func (bucket *Bucket) First() string {
	modelsStr, err := bucket.GetAllStr()
	if err != nil || len(modelsStr) == 0 {
		return ""
	}

	return modelsStr[0]
}

func (bucket *Bucket) Last() string {
	modelsStr, err := bucket.GetAllStr()
	if err != nil || len(modelsStr) == 0 {
		return ""
	}

	return modelsStr[len(modelsStr)-1]
}
