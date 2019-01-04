package actions

import (
	"github.com/alexmorten/events-api/db"
	"github.com/alexmorten/events-api/models"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

//ActionHandler holds things that are shared between actions
type ActionHandler struct {
	dbDriver neo4j.Driver
}

//NewActionHandler ...
func NewActionHandler() *ActionHandler {
	return &ActionHandler{
		dbDriver: db.NewDB(),
	}
}

func (*ActionHandler) currentUserClaim(c *gin.Context) *models.UserClaim {
	userClaimInterface, exists := c.Get("currentUserClaim")
	if !exists {
		return nil
	}
	userClaim, conversionOk := userClaimInterface.(*models.UserClaim)
	if !conversionOk {
		return nil
	}
	return userClaim
}
