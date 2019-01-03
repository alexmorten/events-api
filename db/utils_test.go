package db_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/alexmorten/events-api/db"
	"github.com/stretchr/testify/assert"
)

type SomeBaseModel struct {
	UID uuid.UUID `neo:"uid"`
	A   string    `neo:"a"`
	B   int       `neo:"b"`
}

type SomeModel struct {
	SomeBaseModel
	C time.Time `neo:"c"`
	D string    `something:"else"`
}

func Test_UnmarshalNeoFields(t *testing.T) {
	uid := uuid.New()
	props := map[string]interface{}{
		"a":   "123",
		"b":   123,
		"c":   time.Time{}.Format("2006-01-02 15:04:05.999999999 -0700 MST"),
		"D":   "1234",
		"uid": uid.String(),
	}
	m := &SomeModel{}
	db.UnmarshalNeoFields(m, props)
	assert.Equal(t, "123", m.A)
	assert.Equal(t, 123, m.B)
	assert.Equal(t, time.Time{}, m.C)
	assert.Equal(t, uid, m.UID)
	assert.Equal(t, "", m.D)
}

func Test_MarshalNeoFields(t *testing.T) {
	uid := uuid.New()
	timeValue := time.Time{}
	m := &SomeModel{}
	m.A = "123"
	m.B = 123
	m.C = timeValue
	m.D = "1234"
	m.UID = uid

	props := db.MarshalNeoFields(m)
	assert.Equal(t, "123", props["a"])
	assert.Equal(t, 123, props["b"])
	assert.Equal(t, timeValue.Format("2006-01-02 15:04:05.999999999 -0700 MST"), props["c"])
	assert.Equal(t, uid.String(), props["uid"])
	assert.Equal(t, nil, props["d"])
}

func Test_NeoFields(t *testing.T) {
	m := &SomeModel{}
	assert.Equal(t, []string{"uid", "a", "b", "c"}, db.NeoFields(m))
}
