package interfaces

import (
	. "github.com/PoulIgorson/sub_engine_fiber/database/errors"
)

type DB interface {
	Close() Error

	Table(name string, model Model) (Table, Error)
	ExistsTable(name string) bool
	TableFromCache(name string) Table
	TableOfModel(model Model) Table
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
	SetManager(ManagerI)
}

type Model interface {
	Create(DB, string) Model // Create(DB, jsonString) Model

	Id() any
	Save(table Table) error
	Delete(DB) error
}

type Nexter func() any
type Params map[string]any
type ManagerI interface {
	IsInstance() bool
	Table() Table
	Copy() ManagerI
	Clear()

	Get(id any) Model
	Delete(id any)
	Store(id any, model Model)
	ClearId(id any)

	Broadcast(Nexter)
	All() []Model
	Filter(include Params, exclude ...Params) ManagerI

	Count() uint
	First() Model
	Last() Model
}
