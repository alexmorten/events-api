package actions_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/alexmorten/events-api/models"

	"github.com/alexmorten/events-api"
	"github.com/stretchr/testify/require"

	"github.com/alexmorten/events-api/testhelpers"

	"github.com/alexmorten/events-api/db"
)

func Test_Sports(t *testing.T) {
	dbDriver := db.Driver()
	s := api.NewServer("")
	s.Init()
	t.Run("unauthorized requests return 401", func(t *testing.T) {
		testhelpers.Clear(dbDriver)

		w := httptest.NewRecorder()
		body := `{"name":"blubbi di blup"}`
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("POST", "/sports", reader)

		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("admins can create and get a sport", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		w := httptest.NewRecorder()
		body := `{"name":"blubbi di blup"}`
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("POST", "/sports", reader)

		testhelpers.AddSomeAuthorization(dbDriver, req)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code)

		w = httptest.NewRecorder()
		user := testhelpers.CreateAdminUser(dbDriver)
		body = `{"name":"blubbi di blup"}`
		reader = bytes.NewReader([]byte(body))
		req, _ = http.NewRequest("POST", "/sports", reader)

		testhelpers.AddAuthorizationHeader(req, user)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/sports", nil)
		s.Engine.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		sports := &[]models.Sport{}
		err := json.Unmarshal(w.Body.Bytes(), sports)
		require.NoError(t, err)
		require.Len(t, *sports, 1)
		require.Equal(t, (*sports)[0].Name, "blubbi di blup")
	})

	t.Run("POSTS to /sports cannot set the uid", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		w := httptest.NewRecorder()
		body := `{"name":"blubbi di blup", "uid": "6ec69d34-2abe-4072-bf70-c423f342da73"}`
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("POST", "/sports", reader)

		user := testhelpers.CreateAdminUser(dbDriver)
		testhelpers.AddAuthorizationHeader(req, user)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		sport := &models.Sport{}
		err := json.Unmarshal(w.Body.Bytes(), sport)
		require.NoError(t, err)

		assert.NotEqual(t, "6ec69d34-2abe-4072-bf70-c423f342da73", sport.UID.String())
	})

	t.Run("can update a sport", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		user := testhelpers.CreateAdminUser(dbDriver)
		sport := models.NewSport()
		sport.Name = "Before"
		props, err := db.CreateBy(dbDriver, sport, user.UID)
		require.NoError(t, err)
		sportUID := props["uid"].(string)

		w := httptest.NewRecorder()
		body := `{"name":"Should not be accepted"}`
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("PATCH", "/sports/"+sportUID, reader)

		testhelpers.AddSomeAuthorization(dbDriver, req)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code)

		w = httptest.NewRecorder()
		body = `{"name":"After"}`
		reader = bytes.NewReader([]byte(body))
		req, _ = http.NewRequest("PATCH", "/sports/"+sportUID, reader)

		testhelpers.AddAuthorizationHeader(req, user)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		updatedSport := &models.Sport{}
		err = json.Unmarshal(w.Body.Bytes(), updatedSport)
		require.NoError(t, err)
		assert.Equal(t, "After", updatedSport.Name)
	})

	t.Run("can delete a sport", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		user := testhelpers.CreateAdminUser(dbDriver)
		sport := models.NewSport()
		sport.Name = "Before"
		props, err := db.CreateBy(dbDriver, sport, user.UID)
		require.NoError(t, err)
		sportUID := props["uid"].(string)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/sports/"+sportUID, nil)

		testhelpers.AddSomeAuthorization(dbDriver, req)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("DELETE", "/sports/"+sportUID, nil)

		testhelpers.AddAuthorizationHeader(req, user)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusNoContent, w.Code)

		time.Sleep(50 * time.Millisecond)
		foundClub, err := models.FindSport(dbDriver, sportUID)
		assert.Error(t, err)
		assert.Nil(t, foundClub)
	})
}
