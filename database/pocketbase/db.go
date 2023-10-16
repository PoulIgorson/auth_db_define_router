package pocketbase

import (
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/PoulIgorson/sub_engine_fiber/database/base"
	. "github.com/PoulIgorson/sub_engine_fiber/database/errors"
	. "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
	. "github.com/PoulIgorson/sub_engine_fiber/log"
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

func (m *collectionMap) Range(fn func(name string, collection *Collection) (continue_ bool)) {
	m.Map.Range(func(nameI any, collectionI any) bool {
		return fn(nameI.(string), collectionI.(*Collection))
	})
}

type DataBase struct {
	pb          *PocketBase
	collections collectionMap // map[string]Table
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
func (db *DataBase) DB() *PocketBase {
	return db.pb
}

func (db *DataBase) Close() error {
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
	db.collections.Range(func(_ string, collection *Collection) (continue_ bool) {
		_, err := collection.Get(model.Id())
		if err == nil {
			table = collection
			return false
		}
		return true
	})
	return table
}

func getType(modelV reflect.Value) string {
	switch modelV.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "number"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return "number"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "text"
	case reflect.Pointer:
		return getType(modelV.Elem())
	}
	return ""
}

func getPBField(fieldT reflect.StructField, modelV reflect.Value) map[string]any {
	if fieldT.Name == "ID" {
		return nil
	}
	data := map[string]any{
		"name": fieldT.Tag.Get("json"),
	}
	if fieldT.Tag.Get("typePB") != "" {
		data["type"] = fieldT.Tag.Get("typePB")
		return data
	}
	typ := getType(modelV)
	if typ == "nil" {
		return nil
	}
	if typ != "" {
		data["type"] = typ
	} else {
		valueI := modelV.Interface()
		if _, ok := valueI.(time.Time); ok {
			data["type"] = "date"
		} else if _, ok := valueI.(*url.URL); ok {
			data["type"] = "url"
		} else if _, ok := valueI.(url.URL); ok {
			data["type"] = "url"
		} else {
			data["type"] = "json"
		}
	}
	return data
}

func (db *DataBase) CreateCollection(name string, model Model) error {
	if model == nil {
		return NewErrorf("model is nil")
	}
	if _, ok := model.Id().(string); !ok && name != "user" {
		return NewErrorf("pb: id must be string")
	}
	data := map[string]any{
		"type": "base",
		"name": name,
	}
	schema := []map[string]any{}

	modelT := reflect.TypeOf(model)
	if modelT.Kind() != reflect.Pointer {
		return NewErrorf("pb: invalid type: expected %v, getted %v", reflect.Pointer, modelT.Kind())
	}
	modelT = modelT.Elem()
	modelV := reflect.ValueOf(model).Elem()
	for i := 0; i < modelT.NumField(); i++ {
		fieldT := modelT.Field(i)
		if !fieldT.IsExported() || fieldT.Tag.Get("json") == "-" || fieldT.Tag.Get("json") == "" {
			continue
		}

		if field := getPBField(fieldT, modelV.Field(i)); field != nil {
			schema = append(schema, field)
		}
	}
	data["schema"] = schema

	if db.ExistsTable(name) {
		return ToError(db.pb.UpdateCollection(data))
	}

	return ToError(db.pb.CreateCollection(data))
}

func (db *DataBase) Table(name string, model Model) (Table, error) {
	if collection, ok := db.collections.LoadOK(name); ok {
		return collection, nil
	}

	if err := db.CreateCollection(name, model); err != nil {
		return nil, err
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
		LogError.Panicln(err)
	}
	return err == nil
}
