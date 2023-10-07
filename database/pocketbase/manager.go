package pocketbase

import (
	"encoding/json"

	"github.com/PoulIgorson/sub_engine_fiber/database/base"
	. "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
	. "github.com/PoulIgorson/sub_engine_fiber/define"
)

func ManagerFilter(manager ManagerI, include Params, _ ...Params) []Model {
	objects := []Model{}
	records, _ := manager.Table().DB().(*DataBase).pb.Filter(manager.Table().Name(), nameFieldsToJSONTags(manager.Table().Model(), include))
	for _, record := range records {
		model := recordToModel(record, manager.Table().DB(), manager.Table().Model())
		objects = append(objects, model)
	}
	return objects
}

func ManagerAll(manager ManagerI) []Model {
	objects := []Model{}
	if manager.IsInstance() {
		oldAll := manager.(*base.Manager).OnAll
		manager.(*base.Manager).OnAll = nil
		objects = manager.All()
		manager.(*base.Manager).OnAll = oldAll
	} else {
		records, _ := manager.Table().DB().(*DataBase).pb.Filter(manager.Table().Name(), map[string]any{})
		for _, record := range records {
			model := recordToModel(record, manager.Table().DB(), manager.Table().Model())
			objects = append(objects, model)
		}
	}
	return objects
}

func recordToModel(record *Record, db DB, model Model) Model {
	dataByte, _ := json.Marshal(record.data)
	return model.Create(db, string(dataByte))
}

func nameFieldsToJSONTags(model Model, params Params) Params {
	tagParams := Params{}
	for nameField, value := range params {
		tag := GetTagField(model, nameField, "json")
		if tag != "" {
			tagParams[tag] = value
		} else {
			tagParams[nameField] = value
		}
	}
	return tagParams
}
