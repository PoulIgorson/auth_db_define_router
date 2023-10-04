package pocketbase

import (
	"sync"

	"github.com/PoulIgorson/sub_engine_fiber/database/base"
	. "github.com/PoulIgorson/sub_engine_fiber/database/errors"
	. "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
)

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

func (m *collectionMap) Range(fn func(name string, collection *Collection) bool) {
	m.Map.Range(func(nameI any, collectionI any) bool {
		return fn(nameI.(string), collectionI.(*Collection))
	})
}

type DataBase struct {
	pb          *PocketBase
	collections collectionMap
}

func Open(address, identity, password string) *DataBase {
	return &DataBase{
		pb: New(address, identity, password),
	}
}

func OpenWith(pb *PocketBase) *DataBase {
	return &DataBase{
		pb: pb,
	}
}

func (db *DataBase) Close() Error {
	return nil
}

func (db *DataBase) TableFromCache(name string) Table {
	return db.collections.Load(name)
}

func (db *DataBase) TableOfModel(model Model) Table {
	if model.Id() == nil {
		return nil
	}
	if _, ok := model.Id().(string); !ok {
		return nil
	}
	var table Table
	db.collections.Range(func(_ string, collection *Collection) bool {
		_, err := collection.Get(model.Id())
		if err == nil {
			table = collection
			return true
		}
		return false
	})
	return table
}

func (db *DataBase) Table(name string, model Model) (Table, Error) {
	if collection, ok := db.collections.LoadOK(name); ok {
		return collection, nil
	}
	if !db.ExistsTable(name) {
		return nil, NewErrorf("pb: table not exists")
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
	return err == nil
}
