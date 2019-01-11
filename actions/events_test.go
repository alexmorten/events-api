package actions_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alexmorten/events-api/models"

	"github.com/alexmorten/events-api"
	"github.com/stretchr/testify/require"

	"github.com/alexmorten/events-api/testhelpers"

	"github.com/alexmorten/events-api/db"
)

func Test_PostEvent(t *testing.T) {
	dbDriver := db.NewDB()
	s := api.NewServer("")
	s.Init()
	t.Run("unauthorized requests return 401", func(t *testing.T) {
		testhelpers.Clear(dbDriver)

		w := httptest.NewRecorder()
		body := `{"name":"blubbi di blup"}`
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("POST", "/events", reader)

		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("requests successful with the Authorization header", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		w := httptest.NewRecorder()
		body := `{"name":"blubbi di blup"}`
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("POST", "/events", reader)

		testhelpers.AddSomeAuthorization(dbDriver, req)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/events", nil)
		testhelpers.AddSomeAuthorization(dbDriver, req)
		s.Engine.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		events := &[]models.Event{}
		err := json.Unmarshal(w.Body.Bytes(), events)
		require.NoError(t, err)
		require.Len(t, *events, 1)
		require.Equal(t, (*events)[0].Name, "blubbi di blup")
	})

	t.Run("POSTS to /events cannot set the uid", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		w := httptest.NewRecorder()
		body := `{"name":"blubbi di blup", "uid": "6ec69d34-2abe-4072-bf70-c423f342da73"}`
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("POST", "/events", reader)

		testhelpers.AddSomeAuthorization(dbDriver, req)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		event := &models.Event{}
		err := json.Unmarshal(w.Body.Bytes(), event)
		require.NoError(t, err)

		assert.NotEqual(t, "6ec69d34-2abe-4072-bf70-c423f342da73", event.UID.String())
	})
}
