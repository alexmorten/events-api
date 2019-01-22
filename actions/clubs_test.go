package actions_test

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func Test_Clubs(t *testing.T) {
	dbDriver := db.NewDB()
	s := api.NewServer("")
	s.Init()
	t.Run("unauthorized requests return 401", func(t *testing.T) {
		testhelpers.Clear(dbDriver)

		w := httptest.NewRecorder()
		body := `{"name":"blubbi di blup"}`
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("POST", "/clubs", reader)

		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("admins can create and get a club", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		w := httptest.NewRecorder()
		body := `{"name":"blubbi di blup"}`
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("POST", "/clubs", reader)

		testhelpers.AddSomeAuthorization(dbDriver, req)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code)

		w = httptest.NewRecorder()
		user := testhelpers.CreateAdminUser(dbDriver)
		body = `{"name":"blubbi di blup"}`
		reader = bytes.NewReader([]byte(body))
		req, _ = http.NewRequest("POST", "/clubs", reader)

		testhelpers.AddAuthorizationHeader(req, user)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/clubs", nil)
		s.Engine.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		clubs := &[]models.Club{}
		err := json.Unmarshal(w.Body.Bytes(), clubs)
		require.NoError(t, err)
		require.Len(t, *clubs, 1)
		require.Equal(t, (*clubs)[0].Name, "blubbi di blup")
	})

	t.Run("POSTS to /clubs cannot set the uid", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		w := httptest.NewRecorder()
		body := `{"name":"blubbi di blup", "uid": "6ec69d34-2abe-4072-bf70-c423f342da73"}`
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("POST", "/clubs", reader)

		user := testhelpers.CreateAdminUser(dbDriver)
		testhelpers.AddAuthorizationHeader(req, user)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		club := &models.Club{}
		err := json.Unmarshal(w.Body.Bytes(), club)
		require.NoError(t, err)

		assert.NotEqual(t, "6ec69d34-2abe-4072-bf70-c423f342da73", club.UID.String())
	})

	t.Run("can update a club", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		user := testhelpers.CreateAdminUser(dbDriver)
		club := models.NewClub()
		club.Name = "Before"
		props, err := db.CreateBy(dbDriver, club, user.UID)
		require.NoError(t, err)
		clubUID := props["uid"].(string)

		w := httptest.NewRecorder()
		body := `{"name":"Should not be accepted"}`
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("PATCH", "/clubs/"+clubUID, reader)

		testhelpers.AddSomeAuthorization(dbDriver, req)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code)

		w = httptest.NewRecorder()
		body = `{"name":"After"}`
		reader = bytes.NewReader([]byte(body))
		req, _ = http.NewRequest("PATCH", "/clubs/"+clubUID, reader)

		testhelpers.AddAuthorizationHeader(req, user)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		updatedClub := &models.Club{}
		err = json.Unmarshal(w.Body.Bytes(), updatedClub)
		require.NoError(t, err)
		assert.Equal(t, "After", updatedClub.Name)
	})

	t.Run("can delete a club", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		user := testhelpers.CreateAdminUser(dbDriver)
		club := models.NewClub()
		club.Name = "Before"
		props, err := db.CreateBy(dbDriver, club, user.UID)
		require.NoError(t, err)
		clubUID := props["uid"].(string)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/clubs/"+clubUID, nil)

		testhelpers.AddSomeAuthorization(dbDriver, req)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("DELETE", "/clubs/"+clubUID, nil)

		testhelpers.AddAuthorizationHeader(req, user)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusNoContent, w.Code)

		time.Sleep(50 * time.Millisecond)
		foundClub, err := models.FindEvent(dbDriver, clubUID)
		assert.Error(t, err)
		assert.Nil(t, foundClub)
	})

	t.Run("global admins can add a club admin", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		userToPromoted := testhelpers.CreateSomeUser(dbDriver)
		nonAdminUser := testhelpers.CreateSomeUser(dbDriver)
		adminUser := testhelpers.CreateAdminUser(dbDriver)
		club := models.NewClub()
		_, err := db.Save(dbDriver, club)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		body := fmt.Sprintf(`{"uid": "%v"}`, userToPromoted.UID.String())
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("POST", fmt.Sprintf("/clubs/%v/admins", club.UID.String()), reader)

		testhelpers.AddAuthorizationHeader(req, nonAdminUser)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code)

		w = httptest.NewRecorder()
		reader = bytes.NewReader([]byte(body))
		req, _ = http.NewRequest("POST", fmt.Sprintf("/clubs/%v/admins", club.UID.String()), reader)

		testhelpers.AddAuthorizationHeader(req, adminUser)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		//GET /admins

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", fmt.Sprintf("/clubs/%v/admins", club.UID.String()), nil)
		testhelpers.AddAuthorizationHeader(req, nonAdminUser)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code)

		time.Sleep(50 * time.Millisecond)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", fmt.Sprintf("/clubs/%v/admins", club.UID.String()), nil)
		testhelpers.AddAuthorizationHeader(req, adminUser)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		users := &[]models.PublicUserAttributes{}
		err = json.Unmarshal(w.Body.Bytes(), users)
		require.NoError(t, err)
		require.Len(t, *users, 1)
		require.Equal(t, userToPromoted.UID, (*users)[0].UID)
	})
}
