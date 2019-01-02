package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

//Model is the base for all models
type Model struct {
	UID       uuid.UUID `json:"uid"`
	CreatedAt time.Time `json:"created_at"`
	created   bool
}

func modelPropsString() string {
	return "uid: $uid, created_at: $created_at"
}

func newModel() Model {
	return Model{
		UID:       uuid.New(),
		CreatedAt: time.Now(),
		created:   true,
	}
}

func modelFromProps(props map[string]interface{}) Model {
	model := Model{}
	if value, ok := props["uid"]; ok {
		uuidString, ok := value.(string)
		if ok {
			model.UID, _ = uuid.Parse(uuidString)
		}
	}

	if value, ok := props["created_at"]; ok {
		timeString, ok := value.(string)
		if ok {
			createdAt, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", timeString)
			if err == nil {
				model.CreatedAt = createdAt
			} else {
				fmt.Println(err)
			}
		}
	}
	return model
}
