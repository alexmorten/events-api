package db

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/alexmorten/events-api/utils"

	"github.com/google/uuid"
)

var timeType = reflect.TypeOf(time.Time{})
var stringType = reflect.TypeOf("")
var uuidType = reflect.TypeOf(uuid.UUID{})

//UnmarshalNeoFields of the given interface
//interface should be a pointer to some struct
func UnmarshalNeoFields(obj interface{}, props map[string]interface{}) {
	forEachSettableNeoStructField(reflect.ValueOf(obj).Elem(), func(field reflect.Value, tag string) {
		prop := props[tag]
		propVal := reflect.ValueOf(prop)
		propType := propVal.Type()
		fieldType := field.Type()
		if propType == fieldType {
			field.Set(propVal)
		} else {
			switch fieldType {
			case uuidType:
				if propType == stringType {
					uid, err := uuid.Parse(prop.(string))
					if err == nil {
						field.Set(reflect.ValueOf(uid))
					} else {
						fmt.Println(err)
					}
				}
			case timeType:
				if propType == stringType {
					timeValue, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", prop.(string))
					if err == nil {
						field.Set(reflect.ValueOf(timeValue))
					} else {
						fmt.Println(err)
					}
				}
			}
		}
	})
}

//MarshalNeoFields returns a map that can be given to a neo4j run call
func MarshalNeoFields(obj interface{}) map[string]interface{} {
	props := map[string]interface{}{}
	forEachSettableNeoStructField(reflect.ValueOf(obj).Elem(), func(field reflect.Value, tag string) {
		fieldInterface := field.Interface()
		switch fieldInterface.(type) {
		case uuid.UUID:
			uid := fieldInterface.(uuid.UUID)
			props[tag] = uid.String()
		case time.Time:
			timeValue := fieldInterface.(time.Time)
			props[tag] = timeValue.Format("2006-01-02 15:04:05.999999999 -0700 MST")
		default:
			props[tag] = fieldInterface
		}
	})
	return props
}

//NeoFields returns the fields of the given struct tag have the tag `neo:"<something>"`
func NeoFields(obj interface{}) (neoFieldNames []string) {
	forEachSettableNeoStructField(reflect.ValueOf(obj).Elem(), func(field reflect.Value, tag string) {
		neoFieldNames = append(neoFieldNames, tag)
	})
	return
}

//NeoPropString can be used inside a neo4j query to define all props of a node represented by the given struct
func NeoPropString(obj interface{}) string {
	neoFields := NeoFields(obj)
	propParamCombinations := []string{}
	for _, propName := range neoFields {
		propParamCombination := fmt.Sprintf("%v: $%v", propName, propName)
		propParamCombinations = append(propParamCombinations, propParamCombination)
	}
	return strings.Join(propParamCombinations, ", ")
}

func forEachSettableNeoStructField(val reflect.Value, f func(field reflect.Value, tag string)) {
	utils.ForEachNestedField(val, func(field reflect.Value, structField reflect.StructField) {
		tag := structField.Tag.Get("neo")
		if tag != "" {
			f(field, tag)
		}
	})
}
