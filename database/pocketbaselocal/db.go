package pocketbaselocal

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	pocketbase "github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"

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
	app         *pocketbase.PocketBase
	collections collectionMap // map[string]Table
}

func New(appp ...*pocketbase.PocketBase) *DataBase {
	db := &DataBase{}
	if len(appp) > 0 && appp[0] != nil {
		db.app = appp[0]
		return db
	} else {
		app := pocketbase.New()
		db.app = app
	}
	go func() {
		if err := db.app.Start(); err != nil {
			err = NewErrorf("pocketbaselocal: %v", err)
			log.Printf("pocketbaselocal: %v\n", err)
			os.Exit(1)
		}
	}()
	for {
		if db.app.Dao() == nil {
			continue
		}
		_, err := db.app.Dao().FindRecordsByFilter("users", `id!=""`, "-created", 1, 0)
		if err == nil {
			break
		}
	}
	return db
}

func (db *DataBase) Close() error {
	db.collections.Range(func(_ string, collection *Collection) (continue_ bool) {
		collection.Objects.Clear()
		return true
	})
	db.collections = collectionMap{}
	return db.app.DB().Close()
}

func (db *DataBase) UpdateCollection(model Model) error {
	log.Println("UpdateSchema is not available")
	return nil
	name := GetNameModel(model)
	data, err := CreateDataCollection(name, model)
	if err != nil {
		return err
	}

	typ := data["type"].(string)
	dataB, err := json.Marshal(data["schema"])
	if err != nil {
		return NewErrorf("pocketbaselocal.updateCollection.marshalData: %v", err)
	}
	schema := &schema.Schema{}
	err = schema.UnmarshalJSON(dataB)
	if err != nil {
		return NewErrorf("pocketbaselocal.updateCollection.unmarshalJSON: %v", err)
	}

	var collection *models.Collection

	if collectionPB, err := db.app.Dao().FindCollectionByNameOrId(name); err == nil {
		collection = collectionPB
		collection.MarkAsNotNew()
		log.Println("UpdateSchema is not available")
		return nil
	} else {
		collection = &models.Collection{
			Name: name,
			Type: typ,
		}
		collection.MarkAsNew()
	}
	collection.Schema = *schema
	return db.app.Dao().SaveCollection(collection)
}

func (db *DataBase) Table(_ string, model Model) (Table, error) {
	if db.app.Dao() == nil {
		panic("pockebaselocal.database: pocketbase not running")
	}

	name := GetNameModel(model)
	if table := db.collections.Load(name); table != nil {
		return table, nil
	}

	/*if err := db.UpdateCollection(model); err != nil {
		return nil, NewErrorf("pocketbaselocal: %v", err)
	}*/

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
	_, err := db.app.Dao().FindCollectionByNameOrId(name)
	return err == nil
}

func (db *DataBase) TableFromCache(name string) Table {
	return db.collections.Load(name)
}
