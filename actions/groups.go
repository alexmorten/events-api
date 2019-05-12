package actions

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/alexmorten/events-api/db"

	"github.com/alexmorten/events-api/models"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

//RegisterGroupRoutes within the given router group
func (h *ActionHandler) RegisterGroupRoutes(group *gin.RouterGroup) {
	group.GET("/:uid", h.getGroup)
	group.GET("/:uid/groups", h.getGroups)
	group.PATCH("/:uid", h.updateGroup)
	group.POST("/:uid/groups", h.postGroup)
	group.DELETE("/:uid", h.deleteGroup)

	group.GET("/:uid/admins", h.getGroupAdmins)
	group.POST("/:uid/admins", h.postGroupAdmins)
}

func (h *ActionHandler) getGroup(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("uid can't be empty"))
		return
	}

	group, err := models.FindGroup(h.dbDriver, uid)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	c.JSON(200, group)
}

func (h *ActionHandler) getGroups(c *gin.Context) {
	groups := []*models.Group{}

	uid := c.Param("uid")
	if uid == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("uid can't be empty"))
		return
	}

	parentGroup, _ := models.FindGroup(h.dbDriver, uid)
	parentClub, _ := models.FindClub(h.dbDriver, uid)

	if parentClub == nil && parentGroup == nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("no parent group or club found"))
		return
	}

	dbSession, err := h.dbDriver.Session(neo4j.AccessModeRead)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer dbSession.Close()

	records, err := neo4j.Collect(dbSession.Run(
		fmt.Sprintf("match (n:Group)-[:%v]->(parent {uid: $uid}) return properties(n)", models.GroupBelongsToGroupOrClub),
		map[string]interface{}{"uid": uid}))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		fmt.Println(err)
		return
	}

	for _, record := range records {
		propInterface, ok := record.Get("properties(n)")
		if ok {
			props, ok := propInterface.(map[string]interface{})
			if ok {
				groups = append(groups, models.GroupFromProps(props))
			}
		}
	}
	c.JSON(http.StatusOK, groups)
}

func (h *ActionHandler) postGroup(c *gin.Context) {
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
		return
	}

	parentGroup, _ := models.FindGroup(h.dbDriver, uid)
	parentClub, _ := models.FindClub(h.dbDriver, uid)

	if parentClub == nil && parentGroup == nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("no parent group or club found"))
		return
	}
	var parentUID uuid.UUID
	if parentClub != nil {
		parentUID = parentClub.UID
	}
	if parentGroup != nil {
		parentUID = parentGroup.UID
	}

	group := models.NewGroup()
	groupAttributes := &models.GroupAttributes{}
	err := c.ShouldBindJSON(groupAttributes)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	group.GroupAttributes = *groupAttributes
	props, err := db.CreateBy(h.dbDriver, group, currentUserClaim.UID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	createdgroup := models.GroupFromProps(props)

	_, err = db.CreateRelation(h.dbDriver, createdgroup.UID, parentUID, models.GroupBelongsToGroupOrClub)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, createdgroup)
}

type groupAttributesUpdate struct {
	Name *string `json:"name"`
}

func (h *ActionHandler) updateGroup(c *gin.Context) {
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
	group, err := models.FindGroup(h.dbDriver, uid)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	updateAttributes := &groupAttributesUpdate{}
	err = c.ShouldBindJSON(updateAttributes)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	models.UpdateFrom(&group.GroupAttributes, updateAttributes)

	groupProps, err := db.Save(h.dbDriver, group)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(200, models.GroupFromProps(groupProps))
}

func (h *ActionHandler) deleteGroup(c *gin.Context) {
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
	group, err := models.FindGroup(h.dbDriver, uid)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	err = db.DeleteNode(h.dbDriver, group.UID.String())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *ActionHandler) getGroupAdmins(c *gin.Context) {
	currentUserClaim := h.currentUserClaim(c)
	if currentUserClaim == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	uid := c.Param("uid")
	if uid == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("uid can't be empty"))
	}

	groupAdminsAttributes := []models.PublicUserAttributes{}

	dbSession, err := h.dbDriver.Session(neo4j.AccessModeRead)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer dbSession.Close()

	records, err := neo4j.Collect(
		dbSession.Run(
			"match (u:User)-[:ADMINISTERS]->(c:Group {uid: $uid}) return properties(u)",
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
				groupAdminsAttributes = append(groupAdminsAttributes, models.UserFromProps(props).PublicAttributes())
			}
		}
	}
	c.JSON(http.StatusOK, groupAdminsAttributes)
}

func (h *ActionHandler) postGroupAdmins(c *gin.Context) {
	currentUserClaim := h.currentUserClaim(c)
	if currentUserClaim == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	uid, err := uuid.Parse(c.Param("uid"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	group, err := models.FindGroup(h.dbDriver, uid.String())
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	if !group.AdministeredByUser(h.dbDriver, currentUserClaim.UID) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	userPromotion := &userPromotionAttributes{}
	err = c.ShouldBindJSON(userPromotion)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = models.AddAdminToGroup(h.dbDriver, uid, userPromotion.UID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, nil)
}
