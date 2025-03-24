package utils

import (
	"database/sql"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type DataUtil struct{}

func (d DataUtil) StructToKVMap(s interface{}) map[string]interface{} {
	v := reflect.ValueOf(s)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	kvMap := make(map[string]interface{})

	if v.Kind() != reflect.Struct {
		return kvMap
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if !field.IsExported() {
			continue
		}

		if fieldValue.Kind() == reflect.Slice {
			sliceValue := []interface{}{}
			for j := 0; j < fieldValue.Len(); j++ {
				elem := fieldValue.Index(j)
				if elem.Kind() == reflect.Struct {
					sliceValue = append(sliceValue, d.StructToKVMap(elem.Interface()))
				} else if elem.Kind() == reflect.Ptr && elem.Elem().Kind() == reflect.Struct {
					sliceValue = append(sliceValue, d.StructToKVMap(elem.Elem().Interface()))
				} else {
					sliceValue = append(sliceValue, elem.Interface())
				}
			}
			kvMap[field.Name] = sliceValue
		} else if fieldValue.Kind() == reflect.Ptr {
			if !fieldValue.IsNil() {
				if fieldValue.Elem().Kind() == reflect.String ||
					fieldValue.Elem().Kind() == reflect.Int ||
					fieldValue.Elem().Kind() == reflect.Bool ||
					fieldValue.Elem().Kind() == reflect.Float64 {
					kvMap[field.Name] = fieldValue.Elem().Interface()
				} else {
					kvMap[field.Name] = d.StructToKVMap(fieldValue.Interface())
				}
			} else {
				kvMap[field.Name] = nil
			}
		} else if fieldValue.Kind() == reflect.Struct {
			if fieldValue.Type().String() == "time.Time" {
				kvMap[field.Name] = fieldValue.Interface().(time.Time).Format(time.RFC3339)
			} else if fieldValue.Type().String() == "sql.NullTime" {
				nullTime := fieldValue.Interface().(sql.NullTime)
				if nullTime.Valid {
					kvMap[field.Name] = nullTime.Time.Format(time.RFC3339)
				} else {
					kvMap[field.Name] = nil
				}
			} else {
				kvMap[field.Name] = d.StructToKVMap(fieldValue.Interface())
			}
		} else if fieldValue.Kind() == reflect.Interface {
			if !fieldValue.IsNil() {
				underlyingValue := fieldValue.Elem().Interface()
				if reflect.TypeOf(underlyingValue).Kind() == reflect.Struct {
					kvMap[field.Name] = d.StructToKVMap(underlyingValue)
				} else {
					kvMap[field.Name] = underlyingValue
				}
			} else {
				kvMap[field.Name] = nil
			}
		} else {
			kvMap[field.Name] = fieldValue.Interface()
		}
	}

	return kvMap
}

func (d DataUtil) CamelCaseToSnakeCase(str string) string {
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := re.ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(snake)
}

func (d DataUtil) GetModelFields(model interface{}) []string {
	var fields []string
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	t := val.Type()
	if t.Kind() != reflect.Struct {
		return fields
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fields = append(fields, d.CamelCaseToSnakeCase(field.Name))
	}

	return fields
}

var (
	Data DataUtil = DataUtil{}
)
