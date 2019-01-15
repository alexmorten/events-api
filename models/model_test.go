package models_test

import (
	"testing"

	"github.com/alexmorten/events-api/models"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
)

type SomeAttributes struct {
	UID uuid.UUID
	A   string
	B   int
}

type SomeUpdate struct {
	UID *uuid.UUID
	A   *string
	B   *int
}

func Test_UpdateFrom(t *testing.T) {
	uid := uuid.New()
	someAttributes := &SomeAttributes{
		A:   "A",
		B:   2,
		UID: uid,
	}
	b := 4
	someUpdate := &SomeUpdate{
		B: &b,
	}

	models.UpdateFrom(someAttributes, someUpdate)

	assert.Equal(t, "A", someAttributes.A)
	assert.Equal(t, uid, someAttributes.UID)
	assert.Equal(t, 4, someAttributes.B)
}
