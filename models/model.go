package models

import (
	"time"

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
