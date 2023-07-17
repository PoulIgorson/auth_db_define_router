package db

import (
	"reflect"

	. "github.com/PoulIgorson/sub_engine_fiber/define"
)

type Model interface {
	Create(*DB, string) Model
	Id() uint
}

type Params map[string]any

type Manager struct {
	isInstance bool
	bucket     *Bucket

	count     uint
	objects   map[uint]Model
	rwObjects bool
	minId     uint
	maxId     uint
}

func (manager *Manager) IsInstance() bool {
	return manager.isInstance
}

func (manager *Manager) Bucket() *Bucket {
	return manager.bucket
}

func (manager *Manager) Copy() *Manager {
	return &Manager{
		isInstance: true,
		bucket:     manager.bucket,
		objects:    manager.objects,
		count:      manager.count,
		minId:      manager.minId,
		maxId:      manager.maxId,
	}
}

func (manager *Manager) Get(id uint) Model {
	for manager.rwObjects {
	}
	manager.rwObjects = true
	if m := manager.objects[id]; m != nil {
		manager.rwObjects = false
		return m
	}
	manager.rwObjects = false
	model, _ := manager.bucket.Get(id)
	if model != nil {
		for manager.rwObjects {
		}
		manager.rwObjects = true
		manager.objects[id] = model
		manager.rwObjects = false
		manager.count++
		if manager.maxId < id {
			manager.maxId = id
		}
		if manager.minId > id || manager.minId == 0 {
			manager.minId = id
		}
	}
	return model
}

func (manager *Manager) Delete(id uint) {
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

func (manager *Manager) Filter(include Params, exclude ...Params) *Manager {
	newObjects := map[uint]Model{}
	var maxId, minId uint
	be := false
	for manager.rwObjects {
	}
	manager.rwObjects = true
	for id, model := range manager.objects {
		be = true
		if manager.CheckModel(model, include, exclude...) {
			newObjects[id] = model
			if id > maxId {
				maxId = id
			}
			if minId > id || minId == 0 {
				minId = id
			}
		}
	}
	manager.rwObjects = false
	if !be && !manager.isInstance {
		start := manager.minId
		if start == 0 {
			start = 1
		}
		for inc := start; inc < manager.bucket.Count()+1; inc++ {
			model := manager.Get(inc)
			if manager.CheckModel(model, include, exclude...) {
				newObjects[model.Id()] = model
				if model.Id() > maxId {
					maxId = model.Id()
				}
				if minId > model.Id() || minId == 0 {
					minId = model.Id()
				}
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
	be := false
	for id := range manager.objects {
		for manager.rwObjects {
		}
		manager.rwObjects = true
		model := manager.objects[id]
		manager.rwObjects = false
		be = true
		objects = append(objects, model)
	}
	if !be && !manager.isInstance {
		start := manager.minId
		if start == 0 {
			start = 1
		}
		for inc := start; inc < manager.bucket.Count()+1; inc++ {
			model := manager.Get(inc)
			if model == nil {
				continue
			}
			objects = append(objects, model)
		}
	}
	return objects
}

func (manager *Manager) First() Model {
	for manager.rwObjects {
	}
	manager.rwObjects = true
	model := manager.objects[manager.minId]
	manager.rwObjects = false
	return model
}

func (manager *Manager) Last() Model {
	for manager.rwObjects {
	}
	manager.rwObjects = true
	model := manager.objects[manager.maxId]
	manager.rwObjects = false
	return model
}

func (manager *Manager) Count() uint {
	return uint(len(manager.All()))
}
