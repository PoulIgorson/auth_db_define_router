package define

import (
	"net/url"
	"reflect"
	"time"

	. "github.com/PoulIgorson/sub_engine_fiber/database/errors"
	. "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
)

func GetNameModel(model Model) string {
	typeName := ""
	if t := reflect.TypeOf(model); t.Kind() == reflect.Pointer {
		typeName = t.Elem().Name()
	} else {
		typeName = t.Name()
	}
	var name []rune
	for i, ch := range typeName {
		if i == 0 {
			if 'A' <= ch && ch <= 'Z' {
				ch += 0x20
			}
			name = append(name, ch)
			continue
		}
		if 'A' <= ch && ch <= 'Z' {
			if typeName[i-1] != '_' {
				name = append(name, '_')
			}
			name = append(name, ch+0x20)
			continue
		}
		name = append(name, ch)
	}
	return string(name)
}

func GetType(valueV reflect.Value) string {
	valueI := valueV.Interface()
	switch valueV.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "number"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return "number"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "text"
	case reflect.Pointer:
		return GetType(valueV.Elem())
	}

	if _, ok := valueI.(time.Time); ok {
		return "date"
	}
	if _, ok := valueI.(PBTime); ok {
		return "date"
	}
	if _, ok := valueI.(url.URL); ok {
		return "url"
	}

	if valueV.Kind() == reflect.Struct {
		return "json"
	}
	return ""
}

func getPBField(fieldT reflect.StructField, modelV reflect.Value) map[string]any {
	if fieldT.Name == "ID" {
		return nil
	}
	data := map[string]any{
		"name": fieldT.Tag.Get("json"),
	}
	if fieldT.Tag.Get("typePB") != "" {
		data["type"] = fieldT.Tag.Get("typePB")
		return data
	}
	typ := GetType(modelV)
	if typ != "" {
		data["type"] = typ
	} else {
		data["type"] = "json"
	}
	return data
}

func CreateDataCollection(name string, model Model) (map[string]any, error) {
	if model == nil {
		return nil, NewErrorf("model is nil")
	}
	if _, ok := model.Id().(string); !ok && name != "user" {
		return nil, NewErrorf("id must be string")
	}
	data := map[string]any{
		"type": "base",
		"name": name,
	}
	schema := []map[string]any{}

	modelT := reflect.TypeOf(model)
	if modelT.Kind() != reflect.Pointer {
		return nil, NewErrorf("invalid type: expected %v, got %v", reflect.Pointer, modelT.Kind())
	}
	modelT = modelT.Elem()
	if modelT.Kind() != reflect.Struct {
		return nil, NewErrorf("invalid type: expected %v, got %v", reflect.Struct, modelT.Kind())
	}
	modelV := reflect.ValueOf(model).Elem()
	for i := 0; i < modelT.NumField(); i++ {
		fieldT := modelT.Field(i)
		if !fieldT.IsExported() || fieldT.Tag.Get("json") == "-" || fieldT.Tag.Get("json") == "" {
			continue
		}

		if field := getPBField(fieldT, modelV.Field(i)); field != nil {
			schema = append(schema, field)
		}
	}
	data["schema"] = schema

	return data, nil
}
