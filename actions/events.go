package actions

import (
	"errors"
	"net/http"

	"github.com/alexmorten/events-api/db"

	"github.com/alexmorten/events-api/models"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

//RegisterEventRoutes within the given router group
func (h *ActionHandler) RegisterEventRoutes(group *gin.RouterGroup) {
	group.GET("/:uid", h.getEvent)
	group.GET("", h.getEvents)
	group.POST("", h.postEvents)
}

func (h *ActionHandler) getEvent(c *gin.Context) {

	dbSession, err := h.dbDriver.Session(neo4j.AccessModeRead)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	uid := c.Param("uid")
	if uid == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("uid can't be empty"))
	}
	result, err := dbSession.Run("match(n:Event {uid: $uid}) return properties(n)", map[string]interface{}{"uid": uid})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if result.Next() {
		record := result.Record()
		propInterface, ok := record.Get("properties(n)")
		if ok {
			props, ok := propInterface.(map[string]interface{})
			if ok {
				event := models.EventFromProps(props)
				c.JSON(http.StatusOK, event)
				return
			}
		}
	} else {
		c.AbortWithStatus(http.StatusNotFound)
	}
}

func (h *ActionHandler) getEvents(c *gin.Context) {
	events := []*models.Event{}

	dbSession, err := h.dbDriver.Session(neo4j.AccessModeRead)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	result, err := dbSession.Run("match (n:Event) return properties(n)", nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	for result.Next() {
		record := result.Record()
		propInterface, ok := record.Get("properties(n)")
		if ok {
			props, ok := propInterface.(map[string]interface{})
			if ok {
				events = append(events, models.EventFromProps(props))
			}
		}
	}
	c.JSON(http.StatusOK, events)
}

func (h *ActionHandler) postEvents(c *gin.Context) {
	currentUserClaim := h.currentUserClaim(c)
	if currentUserClaim == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	event := models.NewEvent()
	err := c.ShouldBindJSON(event)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	props, err := db.CreateBy(h.dbDriver, event, currentUserClaim.UID.String())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	createdEvent := models.EventFromProps(props)
	c.JSON(http.StatusCreated, createdEvent)
}
