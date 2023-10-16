package interfaces

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	. "github.com/PoulIgorson/sub_engine_fiber/log"
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

func (pbt *PBTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(*pbt).Format("2006-01-02 15:04:05Z") + `"`), nil
}

func JSONParse(data []byte, model Model) error {
	modelV := reflect.ValueOf(model)
	if modelV.Kind() != reflect.Ptr {
		return fmt.Errorf("model must be a pointer to a struct")
	}

	modelV = modelV.Elem()
	if modelV.Kind() != reflect.Struct {
		return fmt.Errorf("model must be a pointer to a struct")
	}

	dict := map[string]any{}
	json.Unmarshal(data, &dict)

	modelT := modelV.Type()
	for i := 0; i < modelT.NumField(); i++ {
		field := modelT.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		value := dict[tag]
		if value == nil {
			continue
		}
		fieldValue := modelV.Field(i)
		if fieldValue.CanSet() {
			fieldType := fieldValue.Type()
			dataValue := reflect.ValueOf(value)
			if dataValue.Type().ConvertibleTo(fieldType) {
				convertedValue := dataValue.Convert(fieldType)
				fieldValue.Set(convertedValue)
			} else if unmarshaler, ok := dataValue.Interface().(json.Unmarshaler); ok {
				bytes, _ := json.Marshal(value)
				unmarshaler.UnmarshalJSON(bytes)
			} else if unmarshaler, ok := dataValue.Interface().(interface{ Unmarshal([]byte) error }); ok {
				bytes, _ := json.Marshal(value)
				unmarshaler.Unmarshal(bytes)
			} else {
				LogError.Printf("%s: cannot assign value for field '%s'\n", modelT.Name(), field.Name)
			}
		}
	}

	return nil
}
