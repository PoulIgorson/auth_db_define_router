package bbolt

import (
	"reflect"

	. "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
	. "github.com/PoulIgorson/sub_engine_fiber/define"
)

type Manager struct {
	isInstance bool
	bucket     *Bucket

	objects   map[uint]Model
	rwObjects bool
	minId     uint
	maxId     uint
}

func (manager *Manager) IsInstance() bool {
	return manager.isInstance
}

func (manager *Manager) Table() Table {
	return manager.bucket
}

func (manager *Manager) Copy() ManagerI {
	return &Manager{
		isInstance: true,
		bucket:     manager.bucket,
		objects:    manager.objects,
		minId:      manager.minId,
		maxId:      manager.maxId,
	}
}

func (manager *Manager) Get(idI any) Model {
	id, ok := idI.(uint)
	if !ok {
		return nil
	}
	for manager.rwObjects {
	}
	manager.rwObjects = true
	m := manager.objects[id]
	manager.rwObjects = false
	if m != nil {
		manager.CheckPointers(m)
		return m
	}

	model, _ := manager.bucket.Get(id)
	if model == nil {
		for manager.rwObjects {
		}
		manager.rwObjects = true
		delete(manager.objects, id)
		manager.rwObjects = false
		return nil
	}

	for manager.rwObjects {
	}
	manager.rwObjects = true
	manager.objects[id] = model
	manager.rwObjects = false
	if manager.maxId < id {
		manager.maxId = id
	}
	if manager.minId > id || manager.minId == 0 {
		manager.minId = id
	}
	return model
}

func (manager *Manager) Delete(id any) {
	manager.bucket.Delete(id)
}

func (manager *Manager) CheckModel(model Model, include Params, exclude ...Params) bool {
	if model == nil {
		return false
	}
	for key, value := range include {
		mvalue, err := Check(model, key)
		if err != nil {
			continue
		}
		if mvalue.Interface() != reflect.ValueOf(value).Interface() {
			return false
		}
	}
	if len(exclude) > 0 {
		for key, value := range exclude[0] {
			mvalue, err := Check(model, key)
			if err != nil {
				continue
			}
			if mvalue.Interface() == reflect.ValueOf(value).Interface() {
				return false
			}
		}
	}
	return true
}

func (manager *Manager) Filter(include Params, exclude ...Params) ManagerI {
	newObjects := map[uint]Model{}
	var maxId, minId uint
	for id, model := range manager.objects {
		manager.CheckPointers(model)
		if manager.CheckModel(model, include, exclude...) {
			for manager.rwObjects {
			}
			manager.rwObjects = true
			newObjects[id] = model
			manager.rwObjects = false
			if maxId < id {
				maxId = id
			}
			if id < minId || minId == 0 {
				minId = id
			}
		}
	}

	return &Manager{
		isInstance: true,
		bucket:     manager.bucket,
		objects:    newObjects,
		maxId:      maxId,
		minId:      minId,
	}
}

func (manager *Manager) All() []Model {
	objects := []Model{}
	for id := range manager.objects {
		for manager.rwObjects {
		}
		manager.rwObjects = true
		model := manager.objects[id]
		manager.rwObjects = false
		manager.CheckPointers(model)
		objects = append(objects, model)
	}
	return objects
}

func (manager *Manager) First() Model {
	for manager.rwObjects {
	}
	manager.rwObjects = true
	model := manager.objects[manager.minId]
	manager.rwObjects = false
	manager.CheckPointers(model)
	return model
}

func (manager *Manager) Last() Model {
	for manager.rwObjects {
	}
	manager.rwObjects = true
	model := manager.objects[manager.maxId]
	manager.rwObjects = false
	manager.CheckPointers(model)
	return model
}

func (manager *Manager) Count() uint {
	return uint(len(manager.All()))
}

func (manager *Manager) CheckPointers(model Model) {
	if model == nil {
		return
	}
	if _, ok := model.Id().(uint); !ok {
		return
	}
	modelT := reflect.TypeOf(model)
	if modelT.Kind() != reflect.Pointer {
		return
	}
	modelT = modelT.Elem()
	modelV := reflect.ValueOf(model).Elem()
	for i := 0; i < modelT.NumField(); i++ {
		field := modelT.Field(i)
		if field.Type.Kind() == reflect.Array || field.Type.Kind() == reflect.Slice {
			list := modelV.Field(i)
			for i := 0; i < list.Len(); i++ {
				elem := list.Index(i)
				elemT := reflect.TypeOf(elem.Interface())
				if elemT.Kind() != reflect.Pointer {
					break
				}
				if !elem.IsNil() {
					submodel, ok := elem.Interface().(Model)
					if !ok {
						break
					}
					manager.CheckPointers(submodel)
					bucket := manager.bucket.db.buckets[GetNameBucket(submodel)]
					if bucket != nil {
						submodel = bucket.Objects.Get(submodel.Id())
						if submodel != nil {
							elem.Set(reflect.ValueOf(submodel))
							continue
						}
					}
					elem.SetZero()
				}
			}
			continue
		}
		if field.Type.Kind() != reflect.Pointer || !field.IsExported() {
			continue
		}
		if field.Tag.Get("json") != "-" {
			continue
		}
		value := modelV.Field(i)
		if !value.IsNil() {
			submodel, ok := value.Interface().(Model)
			if !ok {
				continue
			}
			manager.CheckPointers(submodel)
			bucket := manager.bucket.db.buckets[GetNameBucket(submodel)]
			if bucket != nil {
				submodel = bucket.Objects.Get(submodel.Id())
				if submodel != nil {
					value.Set(reflect.ValueOf(submodel))
					continue
				}
			}
			value.SetZero()
		}
	}
}
