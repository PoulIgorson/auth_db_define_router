package pocketbaselocal

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/pocketbase/pocketbase/models"

	"github.com/PoulIgorson/sub_engine_fiber/database/base"
	. "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
	. "github.com/PoulIgorson/sub_engine_fiber/define"
)

func ManagerFilter(manager ManagerI, include Params, _ ...Params) []Model {
	objects := []Model{}
	filter := ""
	for key, value := range nameFieldsToJSONTags(manager.Table().Model(), include) {
		if strings.IndexByte("=<>~", key[len(key)-1]) == -1 {
			key += "="
		}
		if len(filter) != 0 {
			filter += "&&"
		}
		if str, ok := value.(string); ok {
			value = `"` + str + `"`
		}
		filter += key + fmt.Sprint(value)
	}

	records, err := manager.Table().DB().(*DataBase).app.Dao().FindRecordsByFilter(manager.Table().Name(), filter, "-created", 0, 0)
	if err != nil {
		log.Printf("pocketbaselocal.managerFilter: %v\n", err)
		return nil
	}

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
		records, err := manager.Table().DB().(*DataBase).app.Dao().FindRecordsByFilter(manager.Table().Name(), `id!=""`, "-created", 0, 0)
		if err != nil {
			log.Printf("pocketbaselocal.managerAll: %v\n", err)
			return nil
		}
		for _, record := range records {
			model := recordToModel(record, manager.Table().DB(), manager.Table().Model())
			objects = append(objects, model)
		}
	}
	return objects
}

func recordToModel(record *models.Record, db DB, model Model) Model {
	dataByte, _ := json.Marshal(record.PublicExport())
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
