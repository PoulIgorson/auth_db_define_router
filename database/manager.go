package db

import (
	"encoding/json"
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
		modelBytes, _ := json.Marshal(model)
		modelMap := map[string]any{}
		json.Unmarshal([]byte(modelBytes), &modelMap)
		for key, value := range include {
			if modelMap[key] != value {
				newObjects[i] = nil
				break
			}
		}
		if len(exclude) > 0 {
			for key, value := range exclude[0] {
				if modelMap[key] == value {
					newObjects[i] = nil
					break
				}
			}
		}
	}
	return &Manager{
		isInstance: true,
		bucket:     manager.bucket,
		objects:    newObjects,
	}
}

func (manager *Manager) All() []Model {
	if manager.isInstance {
		return manager.objects
	}
	return manager.bucket.GetAllModels()
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
