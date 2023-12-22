package pocketbase

import (
	"log"
	"strings"
	"sync"

	"github.com/PoulIgorson/sub_engine_fiber/database/base"
	. "github.com/PoulIgorson/sub_engine_fiber/database/define"
	. "github.com/PoulIgorson/sub_engine_fiber/database/errors"
	. "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
)

var _ DB = &DataBase{}

type collectionMap struct {
	sync.Map
}

func (m *collectionMap) Load(name string) *Collection {
	v, ok := m.Map.Load(name)
	if !ok {
		return nil
	}
	return v.(*Collection)
}

func (m *collectionMap) LoadOK(name string) (*Collection, bool) {
	v, ok := m.Map.Load(name)
	if !ok {
		return nil, false
	}
	return v.(*Collection), true
}

func (m *collectionMap) Range(fn func(name string, collection *Collection) (continue_ bool)) {
	m.Map.Range(func(nameI any, collectionI any) bool {
		return fn(nameI.(string), collectionI.(*Collection))
	})
}

type DataBase struct {
	pb          *PocketBase
	collections collectionMap // map[string]Table
}

func Open(address, identity, password string, isAdmin bool, updateCollections ...bool) *DataBase {
	return OpenWith(New(address, identity, password, isAdmin, updateCollections...))
}

func OpenWith(pb *PocketBase) *DataBase {
	db := &DataBase{
		pb: pb,
	}
	return db
}
func (db *DataBase) DB() *PocketBase {
	return db.pb
}

func (db *DataBase) Close() error {
	return nil
}

func (db *DataBase) TableFromCache(name string) Table {
	return db.collections.Load(name)
}

func (db *DataBase) CreateCollection(name string, model Model) error {
	data, err := CreateDataCollection(name, model)
	if err != nil {
		return err
	}

	if db.ExistsTable(name) {
		log.Println("UpdateSchema is not available")
		return nil
		return ToError(db.pb.UpdateCollection(data))
	}

	return ToError(db.pb.CreateCollection(data))
}

func (db *DataBase) Table(name string, model Model) (Table, error) {
	if collection, ok := db.collections.LoadOK(name); ok {
		return collection, nil
	}

	if err := db.CreateCollection(name, model); err != nil {
		if !strings.Contains(err.Error(), "The request requires valid admin authorization token to be set.") {
			return nil, err
		}
	}

	if _, ok := model.Id().(string); !ok && name != "user" {
		return nil, NewErrorf("pb: id must be string")
	}

	collection := &Collection{
		db:    db,
		name:  name,
		model: model,
	}
	manager := base.NewManager(collection)
	manager.OnAll = ManagerAll
	manager.OnFilter = ManagerFilter
	collection.Objects = manager
	db.collections.Store(name, collection)
	return collection, nil
}

func (db *DataBase) ExistsTable(name string) bool {
	_, err := db.pb.Filter(name, map[string]any{})
	if err != nil && strings.Contains(err.Error(), "refused") {
		panic(err)
	}
	return err == nil
}
