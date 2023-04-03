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

	model Model

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
		model:      manager.model,
		objects:    manager.objects,
	}
}

func (manager *Manager) Get(id uint) Model {
	modelStr, err := manager.bucket.Get(int(id))
	if err != nil {
		return nil
	}
	return manager.model.Create(manager.bucket.db, modelStr)
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
		model:      manager.model,
		objects:    newObjects,
	}
}

func (manager *Manager) All() []Model {
	if manager.isInstance {
		return manager.objects
	}

	modelsStr, err := manager.bucket.GetAllStr()
	if err != nil {
		return nil
	}

	objects := []Model{}
	for _, modelStr := range modelsStr {
		objects = append(objects, manager.model.Create(manager.bucket.db, modelStr))
	}
	return objects
}

func (manager *Manager) First() Model {
	if manager.isInstance {
		return manager.objects[0]
	}

	return manager.model.Create(manager.bucket.db, manager.bucket.First())
}

func (manager *Manager) Last() Model {
	if manager.isInstance {
		return manager.objects[len(manager.objects)-1]
	}

	return manager.model.Create(manager.bucket.db, manager.bucket.Last())
}

func (manager *Manager) Count() uint {
	if manager.isInstance {
		return uint(len(manager.objects))
	}

	return manager.bucket.Count()
}
