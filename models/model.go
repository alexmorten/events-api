package models

import (
	"reflect"
	"time"

	"github.com/alexmorten/events-api/utils"

	"github.com/google/uuid"
)

//Model is the base for all models
type Model struct {
	UID       uuid.UUID `json:"uid" neo:"uid"`
	CreatedAt time.Time `json:"created_at" neo:"created_at"`
	created   bool
}

func newModel() Model {
	return Model{
		UID:       uuid.New(),
		CreatedAt: time.Now(),
		created:   true,
	}
}

//Created shows if the model was just created
func (m *Model) Created() bool {
	return m.created
}

//UpdateFrom some struct
func UpdateFrom(obj, updateFrom interface{}) {
	objVal := reflect.ValueOf(obj).Elem()
	updateFromVal := reflect.ValueOf(updateFrom).Elem()
	utils.ForEachDirectField(updateFromVal, func(field reflect.Value, structField reflect.StructField) {
		objField := objVal.FieldByName(structField.Name)
		if !objField.IsValid() {
			return
		}

		if field.Kind() == reflect.Ptr {
			if !field.IsNil() && field.Elem().Type() == objField.Type() && objField.CanSet() {
				objField.Set(field.Elem())
			}
		} else {
			if field.Type() == objField.Type() && objField.CanSet() {
				objField.Set(field)
			}
		}
	})
}
