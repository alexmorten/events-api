package actions

import (
	"errors"
	"net/http"

	"github.com/alexmorten/events-api/db"

	"github.com/alexmorten/events-api/models"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

//RegisterSportRoutes within the given router group
func (h *ActionHandler) RegisterSportRoutes(group *gin.RouterGroup) {
	group.GET("/:uid", h.getSport)
	group.GET("", h.getSports)
	group.PATCH("/:uid", h.updateSport)
	group.POST("", h.postSports)
	group.DELETE("/:uid", h.deleteSport)
}

func (h *ActionHandler) getSport(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("uid can't be empty"))
	}

	sport, err := models.FindSport(h.dbDriver, uid)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	c.JSON(200, sport)
}

func (h *ActionHandler) getSports(c *gin.Context) {
	sports := []*models.Sport{}

	dbSession, err := h.dbDriver.Session(neo4j.AccessModeRead)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	records, err := neo4j.Collect(dbSession.Run("match (n:Sport) return properties(n)", nil))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	for _, record := range records {
		propInterface, ok := record.Get("properties(n)")
		if ok {
			props, ok := propInterface.(map[string]interface{})
			if ok {
				sports = append(sports, models.SportFromProps(props))
			}
		}
	}
	c.JSON(http.StatusOK, sports)
}

func (h *ActionHandler) postSports(c *gin.Context) {
	currentUserClaim := h.currentUserClaim(c)
	if currentUserClaim == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if !currentUserClaim.Admin {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	sport := models.NewSport()
	sportAttributes := &models.SportAttributes{}
	err := c.ShouldBindJSON(sportAttributes)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	sport.SportAttributes = *sportAttributes
	props, err := db.CreateBy(h.dbDriver, sport, currentUserClaim.UID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	createdSport := models.SportFromProps(props)
	c.JSON(http.StatusCreated, createdSport)
}

type sportAttributesUpdate struct {
	Name *string `json:"name"`
}

func (h *ActionHandler) updateSport(c *gin.Context) {
	currentUserClaim := h.currentUserClaim(c)
	if currentUserClaim == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if !currentUserClaim.Admin {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	uid := c.Param("uid")
	if uid == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("uid can't be empty"))
	}
	sport, err := models.FindSport(h.dbDriver, uid)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	updateAttributes := &sportAttributesUpdate{}
	err = c.ShouldBindJSON(updateAttributes)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	models.UpdateFrom(&sport.SportAttributes, updateAttributes)

	sportProps, err := db.Save(h.dbDriver, sport)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(200, models.SportFromProps(sportProps))
}

func (h *ActionHandler) deleteSport(c *gin.Context) {
	currentUserClaim := h.currentUserClaim(c)
	if currentUserClaim == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if !currentUserClaim.Admin {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	uid := c.Param("uid")
	if uid == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("uid can't be empty"))
	}
	sport, err := models.FindSport(h.dbDriver, uid)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	err = db.DeleteNode(h.dbDriver, sport.UID.String())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
