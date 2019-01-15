package utils

import "reflect"

//ForEachDirectField in val
func ForEachDirectField(val reflect.Value, f func(field reflect.Value, structField reflect.StructField)) {
	valType := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if !field.CanSet() {
			continue
		}

		structField := valType.Field(i)
		f(field, structField)
	}
}

//ForEachNestedField in the struct
func ForEachNestedField(val reflect.Value, f func(reflect.Value, reflect.StructField)) {
	ForEachDirectField(val, func(field reflect.Value, structField reflect.StructField) {
		if field.Kind() == reflect.Struct && structField.Anonymous {
			ForEachNestedField(field, f)
			return
		}
		f(field, structField)
	})
}
