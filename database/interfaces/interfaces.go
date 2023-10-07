package interfaces

import (
	"time"
)

type DB interface {
	Close() error

	Table(name string, model Model) (Table, error)
	ExistsTable(name string) bool
	TableFromCache(name string) Table
	TableOfModel(model Model) Table
}

type Table interface {
	Name() string
	DB() DB
	Model() Model

	Get(id any) (Model, error)
	Save(model Model) error
	Delete(id any) error

	DeleteAll() error
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

// Pocketbase return time in `2006-01-02 15:04:05Z` format.
// And package time parsing in `2006-01-02T15:04:05Z07:00` format.
type PBTime time.Time

func (pbt *PBTime) Unmarshal(data []byte) error {
	t, err := time.Parse("2006-01-02 15:04:05Z", string(data))
	if err != nil {
		return err
	}
	*pbt = PBTime(t)
	return nil
}
