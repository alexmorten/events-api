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
	group.PATCH("/:uid", h.updateEvent)
	group.POST("", h.postEvents)
	group.DELETE("/:uid", h.deleteEvent)
}

func (h *ActionHandler) getEvent(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("uid can't be empty"))
	}

	event, err := models.FindEvent(h.dbDriver, uid)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	c.JSON(200, event)
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
	eventAttributes := &models.EventAttributes{}
	err := c.ShouldBindJSON(eventAttributes)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	event.EventAttributes = *eventAttributes
	props, err := db.CreateBy(h.dbDriver, event, currentUserClaim.UID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	createdEvent := models.EventFromProps(props)
	c.JSON(http.StatusCreated, createdEvent)
}

type eventAttributesUpdate struct {
	Name *string `json:"name"`
}

func (h *ActionHandler) updateEvent(c *gin.Context) {
	currentUserClaim := h.currentUserClaim(c)
	if currentUserClaim == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	uid := c.Param("uid")
	if uid == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("uid can't be empty"))
	}
	event, err := models.FindEvent(h.dbDriver, uid)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	canEdit := event.CanBeEditedBy(h.dbDriver, currentUserClaim.UID)
	if !canEdit {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	updateAttributes := &eventAttributesUpdate{}
	err = c.ShouldBindJSON(updateAttributes)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	models.UpdateFrom(&event.EventAttributes, updateAttributes)

	eventProps, err := db.Save(h.dbDriver, event)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(200, models.EventFromProps(eventProps))
}

func (h *ActionHandler) deleteEvent(c *gin.Context) {
	currentUserClaim := h.currentUserClaim(c)
	if currentUserClaim == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	uid := c.Param("uid")
	if uid == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("uid can't be empty"))
	}
	event, err := models.FindEvent(h.dbDriver, uid)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	canEdit := event.CanBeEditedBy(h.dbDriver, currentUserClaim.UID)
	if !canEdit {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	err = db.DeleteNode(h.dbDriver, uid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
