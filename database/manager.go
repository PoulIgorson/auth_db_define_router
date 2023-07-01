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

var OpenChansModel = map[uint]uint{}
var uidChanModel uint = 0

func CloseChanModel(modelChan chan Model, key uint) {
	if OpenChansModel[key] > 0 {
		close(modelChan)
		OpenChansModel[key] = 0
	}
}

func (manager *Manager) FilterChan(include Params, exclude ...Params) (chan Model, uint) {
	objChan := make(chan Model, 1000)
	key := uidChanModel
	OpenChansModel[key] = 1
	uidChanModel++

	go func(objChan chan Model) {
		ok := false
		do := func(model Model) {
			ok := true
			for key, value := range include {
				mvalue, err := Check(model, key)
				if err != nil {
					continue
				}
				if mvalue.Interface() != reflect.ValueOf(value).Interface() {
					ok = false
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
						ok = false
						break
					}
				}
			}
			if ok {
				if OpenChansModel[key] == 0 {
					return
				}
				objChan <- model
			}
		}
		if manager.isInstance {
			for _, model := range manager.objects {
				do(model)
			}
		} else {
			for inc := uint(1); inc < manager.bucket.Count()+1; inc++ {
				model, err := manager.bucket.Get(inc)
				if err != nil {
					continue
				}
				do(model)
			}
		}
		OpenChansModel[key] = 2
		if !ok {
			CloseChanModel(objChan, key)
		}
	}(objChan)
	return objChan, key
}

func (manager *Manager) All() []Model {
	if manager.isInstance {
		return manager.objects
	}
	objects := []Model{}
	for inc := uint(1); inc < manager.bucket.Count()+1; inc++ {
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
