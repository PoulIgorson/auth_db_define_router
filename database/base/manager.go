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

func (m *modelMap) Range(fn func(id any, model Model) (continue_ bool)) {
	m.Map.Range(func(idI any, modelI any) bool {
		return fn(idI, modelI.(Model))
	})
}

type Manager struct {
	isInstance bool
	table      Table

	objects modelMap // map[uint]Model
	minId   any
	maxId   any

	UseCache bool
	OnCount  func(ManagerI) uint
	OnAll    func(ManagerI) []Model
	OnFilter func(manager ManagerI, include Params, exclude ...Params) []Model
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

		UseCache: manager.UseCache,
		OnCount:  manager.OnCount,
		OnAll:    manager.OnAll,
		OnFilter: manager.OnFilter,
	}
}

func (manager *Manager) Get(id any) Model {
	if manager.UseCache || manager.isInstance {
		model := manager.objects.Load(id)
		if model != nil {
			manager.Store(model.Id(), model)
			manager.CheckPointers(model)
			return model
		}
	}

	model, _ := manager.table.Get(id)
	if model == nil {
		manager.objects.Delete(id)
		return nil
	}

	manager.Store(model.Id(), model)
	return model
}

func (manager *Manager) Delete(id any) {
	manager.table.Delete(id)
	manager.ClearId(id)
}

func (manager *Manager) Store(id any, model Model) {
	if id == nil || model == nil {
		return
	}
	manager.objects.Store(id, model)
	manager.compareAndSetMinMaxId(id)
}

func (manager *Manager) compareAndSetMinMaxId(id any, setNil ...bool) {
	if len(setNil) > 0 && setNil[0] {
		if Compare(manager.maxId, id) == 0 {
			manager.maxId = nil
		}
		if Compare(id, manager.minId) == 0 {
			manager.minId = nil
		}
		return
	}
	if manager.maxId == nil || Compare(manager.maxId, id) == -1 {
		manager.maxId = id
	}
	if manager.minId == nil || Compare(id, manager.minId) == -1 {
		manager.minId = id
	}
}

func (manager *Manager) ClearId(id any) {
	manager.objects.Delete(id)
	manager.compareAndSetMinMaxId(id, true)
}

func (manager *Manager) Broadcast(next Nexter) {
	if next == nil {
		return
	}
	id := next()
	for id != nil {
		model, _ := manager.table.Get(id)
		manager.Store(id, model)
		id = next()
	}
}

func (manager *Manager) CheckModel(model Model, include Params, exclude ...Params) bool {
	if model == nil {
		return false
	}
	check := func(params Params, invert bool) bool {
		for key, value := range params {
			mvalue, err := Check(model, key[:len(key)-getOffset(key)])
			if err != nil {
				continue
			}

			r := Compare(mvalue.Interface(), value)
			if checkKey(key, r) == invert {
				return false
			}
		}
		return true
	}

	return check(include, false) && (len(exclude) == 0 || check(exclude[0], true))
}

func (manager *Manager) All() []Model {
	if manager.OnAll != nil {
		return manager.OnAll(manager)
	}
	objects := []Model{}
	manager.objects.Range(func(id any, model Model) bool {
		manager.CheckPointers(model)
		objects = append(objects, model)
		return true
	})
	return objects
}

func (manager *Manager) Filter(include Params, exclude ...Params) ManagerI {
	newManager := &Manager{
		isInstance: true,
		table:      manager.table,

		UseCache: manager.UseCache,
		OnCount:  manager.OnCount,
		OnAll:    manager.OnAll,
		OnFilter: manager.OnFilter,
	}

	if manager.OnFilter != nil {
		models := manager.OnFilter(manager, include, exclude...)
		for _, model := range models {
			newManager.Store(model.Id(), model)
		}
	} else {
		manager.objects.Range(func(id any, model Model) bool {
			manager.CheckPointers(model)
			if manager.CheckModel(model, include, exclude...) {
				newManager.Store(model.Id(), model)
			}
			return true
		})
	}

	return newManager
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
	if manager.OnCount != nil {
		return manager.OnCount(manager)
	}
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
