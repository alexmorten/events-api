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

	api "github.com/alexmorten/events-api"
	"github.com/stretchr/testify/require"

	"github.com/alexmorten/events-api/testhelpers"

	"github.com/alexmorten/events-api/db"
)

func Test_Groups(t *testing.T) {
	config := api.DefaultServerConfig()
	dbDriver := db.Driver(config.Neo4jAddress)
	s := api.NewServer(config)
	s.Init()

	t.Run("admins can create and get a group inside a club", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		club := models.NewClub()
		_, err := db.Save(dbDriver, club)
		require.NoError(t, err)
		endpoint := fmt.Sprintf("/clubs/%s/groups", club.UID)

		w := httptest.NewRecorder()
		body := `{"name":"blubbi di blup"}`
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("POST", endpoint, reader)

		testhelpers.AddSomeAuthorization(dbDriver, req)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code)

		w = httptest.NewRecorder()
		user := testhelpers.CreateAdminUser(dbDriver)
		body = `{"name":"blubbi di blup"}`
		reader = bytes.NewReader([]byte(body))
		req, _ = http.NewRequest("POST", endpoint, reader)

		testhelpers.AddAuthorizationHeader(req, user)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", endpoint, nil)
		s.Engine.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		groups := &[]models.Group{}
		err = json.Unmarshal(w.Body.Bytes(), groups)
		require.NoError(t, err)
		require.Len(t, *groups, 1)
		require.Equal(t, (*groups)[0].Name, "blubbi di blup")
	})

	t.Run("POSTS to <club>/groups cannot set the uid", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		club := models.NewClub()
		_, err := db.Save(dbDriver, club)
		require.NoError(t, err)
		endpoint := fmt.Sprintf("/clubs/%s/groups", club.UID)

		w := httptest.NewRecorder()
		body := `{"name":"blubbi di blup", "uid": "6ec69d34-2abe-4072-bf70-c423f342da73"}`
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("POST", endpoint, reader)

		user := testhelpers.CreateAdminUser(dbDriver)
		testhelpers.AddAuthorizationHeader(req, user)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		group := &models.Group{}
		err = json.Unmarshal(w.Body.Bytes(), group)
		require.NoError(t, err)

		assert.NotEqual(t, "6ec69d34-2abe-4072-bf70-c423f342da73", club.UID.String())
	})

	t.Run("can update a group", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		user := testhelpers.CreateAdminUser(dbDriver)
		group := models.NewGroup()
		group.Name = "Before"
		props, err := db.CreateBy(dbDriver, group, user.UID)
		require.NoError(t, err)
		groupUID := props["uid"].(string)

		w := httptest.NewRecorder()
		body := `{"name":"Should not be accepted"}`
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("PATCH", "/groups/"+groupUID, reader)

		testhelpers.AddSomeAuthorization(dbDriver, req)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code)

		w = httptest.NewRecorder()
		body = `{"name":"After"}`
		reader = bytes.NewReader([]byte(body))
		req, _ = http.NewRequest("PATCH", "/groups/"+groupUID, reader)

		testhelpers.AddAuthorizationHeader(req, user)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		updatedClub := &models.Club{}
		err = json.Unmarshal(w.Body.Bytes(), updatedClub)
		require.NoError(t, err)
		assert.Equal(t, "After", updatedClub.Name)
	})

	t.Run("can delete a group", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		user := testhelpers.CreateAdminUser(dbDriver)
		group := models.NewGroup()
		group.Name = "Before"
		props, err := db.CreateBy(dbDriver, group, user.UID)
		require.NoError(t, err)
		groupUID := props["uid"].(string)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/groups/"+groupUID, nil)

		testhelpers.AddSomeAuthorization(dbDriver, req)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("DELETE", "/groups/"+groupUID, nil)

		testhelpers.AddAuthorizationHeader(req, user)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusNoContent, w.Code)

		time.Sleep(50 * time.Millisecond)
		foundClub, err := models.FindEvent(dbDriver, groupUID)
		assert.Error(t, err)
		assert.Nil(t, foundClub)
	})

	t.Run("global admins can add a group admin", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		userToPromote := testhelpers.CreateSomeUser(dbDriver)
		nonAdminUser := testhelpers.CreateSomeUser(dbDriver)
		adminUser := testhelpers.CreateAdminUser(dbDriver)
		group := models.NewGroup()
		_, err := db.Save(dbDriver, group)
		require.NoError(t, err)
		endpoint := fmt.Sprintf("/groups/%v/admins", group.UID.String())

		w := httptest.NewRecorder()
		body := fmt.Sprintf(`{"uid": "%v"}`, userToPromote.UID.String())
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("POST", endpoint, reader)

		testhelpers.AddAuthorizationHeader(req, nonAdminUser)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code)

		w = httptest.NewRecorder()
		reader = bytes.NewReader([]byte(body))
		req, _ = http.NewRequest("POST", endpoint, reader)

		testhelpers.AddAuthorizationHeader(req, adminUser)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		//GET /admins

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", endpoint, nil)
		testhelpers.AddAuthorizationHeader(req, adminUser)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		users := &[]models.PublicUserAttributes{}
		err = json.Unmarshal(w.Body.Bytes(), users)
		require.NoError(t, err)
		require.Len(t, *users, 1)
		require.Equal(t, userToPromote.UID, (*users)[0].UID)
	})

	t.Run("group admins can add another group admin", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		userToPromote := testhelpers.CreateSomeUser(dbDriver)
		nonAdminUser := testhelpers.CreateSomeUser(dbDriver)
		adminUser := testhelpers.CreateSomeUser(dbDriver)
		group := models.NewGroup()
		_, err := db.Save(dbDriver, group)
		require.NoError(t, err)
		endpoint := fmt.Sprintf("/groups/%v/admins", group.UID.String())
		_, err = db.CreateRelation(dbDriver, adminUser.UID, group.UID, models.UserAdministersGroupOrClub)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		body := fmt.Sprintf(`{"uid": "%v"}`, userToPromote.UID.String())
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("POST", endpoint, reader)

		testhelpers.AddAuthorizationHeader(req, nonAdminUser)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code)

		w = httptest.NewRecorder()
		reader = bytes.NewReader([]byte(body))
		req, _ = http.NewRequest("POST", endpoint, reader)

		testhelpers.AddAuthorizationHeader(req, adminUser)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		//GET /admins

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", endpoint, nil)
		testhelpers.AddAuthorizationHeader(req, adminUser)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		users := &[]models.PublicUserAttributes{}
		err = json.Unmarshal(w.Body.Bytes(), users)
		require.NoError(t, err)
		require.Len(t, *users, 2)
	})

	t.Run("parent club admins can add another group admin", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		userToPromote := testhelpers.CreateSomeUser(dbDriver)
		nonAdminUser := testhelpers.CreateSomeUser(dbDriver)
		adminUser := testhelpers.CreateSomeUser(dbDriver)

		group := models.NewGroup()
		_, err := db.Save(dbDriver, group)
		require.NoError(t, err)

		club := models.NewClub()
		_, err = db.Save(dbDriver, club)
		require.NoError(t, err)

		_, err = db.CreateRelation(dbDriver, group.UID, club.UID, models.GroupBelongsToGroupOrClub)
		require.NoError(t, err)

		endpoint := fmt.Sprintf("/groups/%v/admins", group.UID.String())
		_, err = db.CreateRelation(dbDriver, adminUser.UID, club.UID, models.UserAdministersGroupOrClub)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		body := fmt.Sprintf(`{"uid": "%v"}`, userToPromote.UID.String())
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("POST", endpoint, reader)

		testhelpers.AddAuthorizationHeader(req, nonAdminUser)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code)

		w = httptest.NewRecorder()
		reader = bytes.NewReader([]byte(body))
		req, _ = http.NewRequest("POST", endpoint, reader)

		testhelpers.AddAuthorizationHeader(req, adminUser)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		//GET /admins

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", endpoint, nil)
		testhelpers.AddAuthorizationHeader(req, adminUser)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		users := &[]models.PublicUserAttributes{}
		err = json.Unmarshal(w.Body.Bytes(), users)
		require.NoError(t, err)
		require.Len(t, *users, 1)
	})

	t.Run("parent group admins can add another group admin", func(t *testing.T) {
		testhelpers.Clear(dbDriver)
		userToPromote := testhelpers.CreateSomeUser(dbDriver)
		nonAdminUser := testhelpers.CreateSomeUser(dbDriver)
		adminUser := testhelpers.CreateSomeUser(dbDriver)

		group := models.NewGroup()
		_, err := db.Save(dbDriver, group)
		require.NoError(t, err)

		parentGroup := models.NewGroup()
		_, err = db.Save(dbDriver, parentGroup)
		require.NoError(t, err)

		_, err = db.CreateRelation(dbDriver, group.UID, parentGroup.UID, models.GroupBelongsToGroupOrClub)
		require.NoError(t, err)

		endpoint := fmt.Sprintf("/groups/%v/admins", group.UID.String())
		_, err = db.CreateRelation(dbDriver, adminUser.UID, parentGroup.UID, models.UserAdministersGroupOrClub)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		body := fmt.Sprintf(`{"uid": "%v"}`, userToPromote.UID.String())
		reader := bytes.NewReader([]byte(body))
		req, _ := http.NewRequest("POST", endpoint, reader)

		testhelpers.AddAuthorizationHeader(req, nonAdminUser)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code)

		w = httptest.NewRecorder()
		reader = bytes.NewReader([]byte(body))
		req, _ = http.NewRequest("POST", endpoint, reader)

		testhelpers.AddAuthorizationHeader(req, adminUser)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		//GET /admins

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", endpoint, nil)
		testhelpers.AddAuthorizationHeader(req, adminUser)
		s.Engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		users := &[]models.PublicUserAttributes{}
		err = json.Unmarshal(w.Body.Bytes(), users)
		require.NoError(t, err)
		require.Len(t, *users, 1)
	})
}
