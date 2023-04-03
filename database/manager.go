package db

import (
	"reflect"

	. "github.com/PoulIgorson/sub_engine_fiber/define"
)

type Model interface {
	Create(*DB, string) Model
}

type Params map[string]any

type Manager struct {
	isInstance bool
	bucket     *Bucket

	objects []Model
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
	}
}

func (manager *Manager) Get(id uint) Model {
	modelStr, _ := manager.bucket.Get(id)
	return modelStr
}

func (manager *Manager) Filter(include Params, exclude ...Params) *Manager {
	objects := manager.All()

	newObjects := []Model(objects)
	for i, model := range objects {
		for key, value := range include {
			mvalue, err := Check(model, key)
			if err != nil {
				continue
			}
			if mvalue.Interface() != reflect.ValueOf(value).Interface() {
				newObjects[i] = nil
				break
			}
		}
		if len(exclude) > 0 {
			for key, value := range exclude[0] {
				mvalue, err := Check(model, key)
				if err != nil {
					continue
				}
				if mvalue.Interface() == reflect.ValueOf(value).Interface() {
					newObjects[i] = nil
					break
				}
			}
		}
	}
	objects = []Model{}
	for _, obj := range newObjects {
		if obj != nil {
			objects = append(objects, obj)
		}
	}
	return &Manager{
		isInstance: true,
		bucket:     manager.bucket,
		objects:    objects,
	}
}

func (manager *Manager) All() []Model {
	if manager.isInstance {
		return manager.objects
	}
	objects := []Model{}
	for inc := uint(1); inc < manager.bucket.Count(); inc++ {
		model, err := manager.bucket.Get(inc)
		if err != nil {
			continue
		}
		objects = append(objects, model)
	}
	return objects
}

func (manager *Manager) First() Model {
	objects := manager.All()
	if len(objects) == 0 {
		return nil
	}
	return objects[0]
}

func (manager *Manager) Last() Model {
	objects := manager.All()
	if len(objects) == 0 {
		return nil
	}
	return objects[len(objects)-1]
}

func (manager *Manager) Count() uint {
	if manager.isInstance {
		return uint(len(manager.objects))
	}
	return manager.bucket.Count()
}
