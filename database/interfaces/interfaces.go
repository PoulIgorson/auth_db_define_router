package interfaces

import (
	. "github.com/PoulIgorson/sub_engine_fiber/database/errors"
)

type DB interface {
	Close() Error

	Table(name string, model Model) (Table, Error)
	ExistsTable(name string) bool
}

type Table interface {
	Name() string
	DB() DB
	Model() Model

	Get(id any) (Model, Error)
	Save(model Model) Error
	Delete(id any) Error
	DeleteAll() Error
	Count() uint
	Manager() ManagerI
}

type Model interface {
	Create(DB, string) Model // Create(DB, jsonString) Model

	Id() any
	Save(table Table) error
	Delete(DB) error
}

type Params map[string]any
type ManagerI interface {
	IsInstance() bool
	Copy() ManagerI
	Table() Table

	Get(id any) Model
	Delete(id any)
	All() []Model
	Filter(include Params, exclude ...Params) ManagerI

	Count() uint
	First() Model
	Last() Model
}
