package actions

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/alexmorten/events-api/db"

	"github.com/alexmorten/events-api/models"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

//RegisterClubRoutes within the given router group
func (h *ActionHandler) RegisterClubRoutes(group *gin.RouterGroup) {
	group.GET("/:uid", h.getClub)
	group.GET("", h.getClubs)
	group.PATCH("/:uid", h.updateClub)
	group.POST("", h.postClubs)
	group.DELETE("/:uid", h.deleteClub)

	group.POST("/:uid/groups", h.postGroup)
	group.GET("/:uid/groups", h.getGroups)

	group.GET("/:uid/admins", h.getAdmins)
	group.POST("/:uid/admins", h.postAdmins)
}

func (h *ActionHandler) getClub(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("uid can't be empty"))
	}

	club, err := models.FindClub(h.dbDriver, uid)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	c.JSON(200, club)
}

func (h *ActionHandler) getClubs(c *gin.Context) {
	clubs := []*models.Club{}

	dbSession, err := h.dbDriver.Session(neo4j.AccessModeRead)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	records, err := neo4j.Collect(dbSession.Run("match (n:Club) return properties(n)", nil))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	for _, record := range records {
		propInterface, ok := record.Get("properties(n)")
		if ok {
			props, ok := propInterface.(map[string]interface{})
			if ok {
				clubs = append(clubs, models.ClubFromProps(props))
			}
		}
	}
	c.JSON(http.StatusOK, clubs)
}

func (h *ActionHandler) postClubs(c *gin.Context) {
	currentUserClaim := h.currentUserClaim(c)
	if currentUserClaim == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if !currentUserClaim.Admin {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	club := models.NewClub()
	clubAttributes := &models.ClubAttributes{}
	err := c.ShouldBindJSON(clubAttributes)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	club.ClubAttributes = *clubAttributes
	props, err := db.CreateBy(h.dbDriver, club, currentUserClaim.UID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	createdclub := models.ClubFromProps(props)
	c.JSON(http.StatusCreated, createdclub)
}

type clubAttributesUpdate struct {
	Name *string `json:"name"`
}

func (h *ActionHandler) updateClub(c *gin.Context) {
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
	club, err := models.FindClub(h.dbDriver, uid)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	updateAttributes := &clubAttributesUpdate{}
	err = c.ShouldBindJSON(updateAttributes)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	models.UpdateFrom(&club.ClubAttributes, updateAttributes)

	clubProps, err := db.Save(h.dbDriver, club)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(200, models.ClubFromProps(clubProps))
}

func (h *ActionHandler) deleteClub(c *gin.Context) {
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
	club, err := models.FindClub(h.dbDriver, uid)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	err = db.DeleteNode(h.dbDriver, club.UID.String())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *ActionHandler) getAdmins(c *gin.Context) {
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

	clubAdminsAttributes := []models.PublicUserAttributes{}

	dbSession, err := h.dbDriver.Session(neo4j.AccessModeRead)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer dbSession.Close()

	records, err := neo4j.Collect(
		dbSession.Run(
			"match (u:User)-[:ADMINISTERS]->(c:Club {uid: $uid}) return properties(u)",
			map[string]interface{}{"uid": uid},
		),
	)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	for _, record := range records {
		propInterface, ok := record.Get("properties(u)")
		if ok {
			props, ok := propInterface.(map[string]interface{})
			if ok {
				clubAdminsAttributes = append(clubAdminsAttributes, models.UserFromProps(props).PublicAttributes())
			}
		}
	}
	c.JSON(http.StatusOK, clubAdminsAttributes)
}

func (h *ActionHandler) postAdmins(c *gin.Context) {
	currentUserClaim := h.currentUserClaim(c)
	if currentUserClaim == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if !currentUserClaim.Admin {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	uid, err := uuid.Parse(c.Param("uid"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	userPromotion := &userPromotionAttributes{}
	err = c.ShouldBindJSON(userPromotion)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = models.AddAdminToClub(h.dbDriver, uid, userPromotion.UID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, nil)
}
