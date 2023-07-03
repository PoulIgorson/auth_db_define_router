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

	count   uint
	objects map[uint]Model
	minId   uint
	maxId   uint
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
	if m := manager.objects[id]; m != nil {
		return m
	}
	model, _ := manager.bucket.Get(id)
	if model != nil {
		if _, ok := manager.objects[id]; !ok {
			manager.count++
			if manager.maxId < id {
				manager.maxId = id
			}
			if manager.minId > id {
				manager.minId = id
			}
		}
		manager.objects[id] = model
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
	for id, model := range manager.objects {
		be = true
		if manager.CheckModel(model, include, exclude...) {
			newObjects[id] = model
			if id > maxId {
				maxId = id
			}
			if minId > id {
				minId = id
			}
		}
	}
	if !be && !manager.isInstance {
		start := manager.minId
		if start == 0 {
			start = 1
		}
		for inc := start; inc < manager.bucket.Count(); inc++ {
			model := manager.Get(inc)
			if manager.CheckModel(model, include, exclude...) {
				newObjects[model.Id()] = model
				if model.Id() > maxId {
					maxId = model.Id()
				}
				if minId > model.Id() {
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
	for _, model := range manager.Filter(Params{}).objects {
		objects = append(objects, model)
	}
	return objects
}

func (manager *Manager) First() Model {
	return manager.objects[manager.minId]
}

func (manager *Manager) Last() Model {
	return manager.objects[manager.maxId]
}

func (manager *Manager) Count() uint {
	return uint(len(manager.All()))
}
