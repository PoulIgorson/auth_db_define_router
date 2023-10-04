package base

import (
	"reflect"
	"sync"

	. "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
	. "github.com/PoulIgorson/sub_engine_fiber/define"
)

type modelMap struct {
	sync.Map
}

func (m *modelMap) Load(id any) Model {
	v, ok := m.Map.Load(id)
	if !ok {
		return nil
	}
	return v.(Model)
}

func (m *modelMap) LoadOK(id any) (Model, bool) {
	v, ok := m.Map.Load(id)
	if !ok {
		return nil, false
	}
	return v.(Model), true
}

func (m *modelMap) Range(fn func(id any, model Model) bool) {
	m.Map.Range(func(idI any, modelI any) bool {
		return fn(idI, modelI.(Model))
	})
}

type Manager struct {
	isInstance bool
	table      Table

	objects modelMap
	minId   any
	maxId   any

	UseCache    bool
	OnClear     func(ManagerI)
	OnBroadcast func(ManagerI, Nexter)
	OnAll       func(ManagerI) []Model
	OnFilter    func(manager ManagerI, include Params, exclude ...Params) []Model
}

func NewManager(table Table) *Manager {
	return &Manager{
		table: table,
	}
}

func (manager *Manager) IsInstance() bool {
	return manager.isInstance
}

func (manager *Manager) Table() Table {
	return manager.table
}

func (manager *Manager) Clear() {
	if manager.OnClear != nil {
		manager.OnClear(manager)
		return
	}
	manager.objects = modelMap{}
	manager.minId = nil
	manager.maxId = nil
}

func (manager *Manager) Copy() ManagerI {
	return &Manager{
		isInstance: true,
		table:      manager.table,

		objects: manager.objects,
		minId:   manager.minId,
		maxId:   manager.maxId,

		UseCache:    manager.UseCache,
		OnClear:     manager.OnClear,
		OnBroadcast: manager.OnBroadcast,
		OnAll:       manager.OnAll,
		OnFilter:    manager.OnFilter,
	}
}

func (manager *Manager) Get(id any) Model {
	if manager.UseCache || manager.isInstance {
		m := manager.objects.Load(id)
		if m != nil {
			manager.CheckPointers(m)
			return m
		}
	}

	model, _ := manager.table.Get(id)
	if model == nil {
		manager.objects.Delete(id)
		return nil
	}

	manager.objects.Store(id, model)
	if Compare(manager.maxId, id) == -1 || manager.maxId == nil {
		manager.maxId = id
	}
	if Compare(id, manager.minId) == -1 || manager.minId == nil {
		manager.minId = id
	}
	return model
}

func (manager *Manager) Delete(id any) {
	manager.table.Delete(id)
}

func (manager *Manager) Store(id any, model Model) {
	manager.objects.Store(id, model)
}

func (manager *Manager) ClearId(id any) {
	manager.objects.Delete(id)
}

func (manager *Manager) Broadcast(next Nexter) {
	if manager.OnBroadcast != nil {
		manager.OnBroadcast(manager, next)
		return
	}
	for id := next(); id != nil; {
		model, _ := manager.table.Get(id)
		if model == nil {
			continue
		}
		if Compare(manager.maxId, id) == -1 || manager.maxId == nil {
			manager.maxId = id
		}
		if Compare(id, manager.minId) == -1 || manager.minId == nil {
			manager.minId = id
		}
		manager.objects.Store(id, model)
	}
}

func (manager *Manager) CheckModel(model Model, include Params, exclude ...Params) bool {
	if model == nil {
		return false
	}
	for key, value := range include {
		mvalue, err := Check(model, key[len(key)-getOffset(key):])
		if err != nil {
			continue
		}

		r := Compare(mvalue.Interface(), value)
		if !checkKey(key, r) {
			return false
		}
	}
	if len(exclude) > 0 {
		for key, value := range exclude[0] {
			mvalue, err := Check(model, key[len(key)-getOffset(key):])
			if err != nil {
				continue
			}

			r := Compare(mvalue.Interface(), value)
			if checkKey(key, r) {
				return false
			}
		}
	}
	return true
}

func (manager *Manager) All() []Model {
	if manager.OnAll != nil {
		return manager.OnAll(manager)
	}
	objects := []Model{}
	manager.objects.Range(func(id any, model Model) bool {
		manager.CheckPointers(model)
		objects = append(objects, model)
		return false
	})
	return objects
}

func (manager *Manager) Filter(include Params, exclude ...Params) ManagerI {
	var newObjects modelMap
	var maxId, minId any
	if manager.OnFilter != nil {
		models := manager.OnFilter(manager, include, exclude...)
		for _, model := range models {
			newObjects.Store(model.Id(), model)
			if Compare(maxId, model.Id()) == -1 || maxId == nil {
				maxId = model.Id()
			}
			if Compare(model.Id(), minId) == -1 || minId == nil {
				minId = model.Id()
			}
		}
	} else {
		manager.objects.Range(func(id any, model Model) bool {
			manager.CheckPointers(model)
			if manager.CheckModel(model, include, exclude...) {
				newObjects.Store(id, model)
				if Compare(maxId, id) == -1 || maxId == nil {
					maxId = id
				}
				if Compare(id, minId) == -1 || minId == nil {
					minId = id
				}
			}
			return false
		})
	}

	return &Manager{
		isInstance: true,
		table:      manager.table,

		objects: newObjects,
		maxId:   maxId,
		minId:   minId,

		UseCache:    manager.UseCache,
		OnClear:     manager.OnClear,
		OnBroadcast: manager.OnBroadcast,
		OnAll:       manager.OnAll,
		OnFilter:    manager.OnFilter,
	}
}

func (manager *Manager) First() Model {
	if manager.minId == nil || !manager.UseCache && !manager.isInstance {
		return nil
	}
	model := manager.objects.Load(manager.minId)
	manager.CheckPointers(model)
	return model
}

func (manager *Manager) Last() Model {
	if manager.maxId == nil || !manager.UseCache && !manager.isInstance {
		return nil
	}
	model := manager.objects.Load(manager.maxId)
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
	modelT := reflect.TypeOf(model)
	if modelT.Kind() != reflect.Pointer {
		return
	}
	modelT = modelT.Elem()
	modelV := reflect.ValueOf(model).Elem()
	for i := 0; i < modelT.NumField(); i++ {
		field := modelT.Field(i)
		if field.Tag.Get("json") != "-" {
			continue
		}
		if field.Type.Kind() == reflect.Pointer && field.IsExported() {
			manager.processCheck(modelV.Field(i))
			continue
		}
		if field.Type.Kind() == reflect.Array || field.Type.Kind() == reflect.Slice {
			list := modelV.Field(i)
			for i := 0; i < list.Len(); i++ {
				value := list.Index(i)
				valueT := reflect.TypeOf(value.Interface())
				if valueT.Kind() != reflect.Pointer {
					break
				}
				if !manager.processCheck(value) {
					break
				}
			}
			continue
		}
	}
}

func (manager *Manager) processCheck(value reflect.Value) bool {
	if !value.IsNil() {
		submodel, ok := value.Interface().(Model)
		if !ok {
			return false
		}
		manager.CheckPointers(submodel)
		table := manager.table.DB().TableOfModel(submodel)
		if table != nil {
			submodel = table.Manager().Get(submodel.Id())
			if submodel != nil {
				value.Set(reflect.ValueOf(submodel))
				return true
			}
		}
		value.SetZero()
	}
	return true
}
