package pocketbaselocal

import (
	"encoding/json"
	"reflect"

	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"

	. "github.com/PoulIgorson/sub_engine_fiber/database/define"
	. "github.com/PoulIgorson/sub_engine_fiber/database/errors"
	. "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
	. "github.com/PoulIgorson/sub_engine_fiber/define"
)

var _ Table = &Collection{}

type Collection struct {
	db   *DataBase
	name string

	model   Model
	Objects ManagerI
}

func (collection Collection) DB() DB {
	return collection.db
}

func (collection Collection) Name() string {
	return collection.name
}

func (collection Collection) Model() Model {
	return collection.model
}

func (collection *Collection) Get(idI any) (Model, error) {
	id, ok := idI.(string)
	if !ok {
		return nil, NewErrorf("pb: id must be string")
	}
	record, err := collection.db.app.Dao().FindRecordById(collection.name, id)
	if err != nil {
		return nil, NewErrorf("pocketbaselocal.table.get: model not found")
	}
	dataByte, _ := json.Marshal(record.PublicExport())
	model := collection.model.Create(collection.db, string(dataByte))
	collection.Objects.Store(model.Id().(string), model)
	return model, nil
}

func (collection *Collection) Save(model Model) error {
	if model.Id() == nil {
		return NewErrorf("pocketbaselocal.collection.save: id experted string, got nil")
	}
	if _, ok := model.Id().(string); !ok {
		return NewErrorf("pocketbaselocal.collection.save: id experted string, got %v", reflect.TypeOf(model.Id()))
	}
	dataByte, _ := json.Marshal(model)
	data := map[string]any{}
	json.Unmarshal(dataByte, &data)

	record, err := collection.db.app.Dao().FindRecordById(collection.name, model.Id().(string))
	if err != nil {
		name := GetNameModel(model)
		collectionPB, err := collection.db.app.Dao().FindCollectionByNameOrId(name)
		if err != nil {
			return NewErrorf("pocketbaselocal.collection.save.findCollection: %v", err)
		}
		record = models.NewRecord(collectionPB)
	}
	delete(data, "id")

	for field, value := range data {
		typ := GetType(reflect.ValueOf(value))
		if typ == "" || typ == "json" {
			valueB, _ := json.Marshal(value)
			value = string(valueB)
		}
		record.Set(field, value)
	}

	if err := collection.db.app.Dao().SaveRecord(record); err != nil {
		return NewErrorf("pocketbaselocal.collection.save.saveRecord: %v", err)
	}

	fieldId, err := Check(model, "ID")
	if err != nil {
		return NewErrorf("pocketbaselocal.getFieldID: %v", err)
	}
	fieldId.Set(reflect.ValueOf(record.Id))
	return nil
}

func (collection *Collection) Delete(idI any) error {
	id, ok := idI.(string)
	if !ok {
		return NewErrorf("pb: id must be string")
	}
	if id == "" {
		return nil
	}
	record, err := collection.db.app.Dao().FindRecordById(collection.name, id)
	if err != nil {
		return NewErrorf("pocketbaselocal.table.delete.findRecord: %v", err)
	}

	if err := collection.db.app.Dao().DeleteRecord(record); err != nil {
		return NewErrorf("pocketbaselocal.table.delete.deleteRecord: %v", err)
	}
	return nil
}

// Pocketbase does not support DeleteAll
// All models will be deletting of one
func (collection *Collection) DeleteAll() error {
	err := collection.db.app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
		records, err := collection.db.app.Dao().FindRecordsByFilter(collection.name, `id!=""`, "-created", 0, 0)
		if err != nil {
			return err
		}
		for _, record := range records {
			txDao.DeleteRecord(record)
		}
		return nil
	})
	if err != nil {
		return NewErrorf("pocketbaselocal.table.deleteAll: %v", err)
	}
	return nil
}

func (collection Collection) Count() uint {
	return 0
}

func (collection Collection) Manager() ManagerI {
	return collection.Objects
}

func (collection *Collection) SetManager(newManager ManagerI) {
	collection.Objects = newManager
}
