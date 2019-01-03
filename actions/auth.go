package actions

import (
	"net/http"
	"os"

	"github.com/alexmorten/events-api/db"

	"github.com/alexmorten/events-api/models"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

//RegisterAuthRoutes responsible for authentication handling
func (h *ActionHandler) RegisterAuthRoutes(group *gin.RouterGroup) {
	googleProvider := google.New(os.Getenv("GOOGLE_CLIENT"), os.Getenv("GOOGLE_SECRET"), "http://localhost:3000/auth/google/callback")
	goth.UseProviders(googleProvider)

	group.GET("/:provider", func(c *gin.Context) {
		q := c.Request.URL.Query()
		q.Add("provider", c.Param("provider"))
		c.Request.URL.RawQuery = q.Encode()

		gothic.BeginAuthHandler(c.Writer, c.Request)
	})

	group.GET("/:provider/callback", func(c *gin.Context) {
		q := c.Request.URL.Query()
		q.Add("provider", c.Param("provider"))
		c.Request.URL.RawQuery = q.Encode()

		gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		user, err := models.FindOrCreateUserByEmail(h.dbDriver, gothUser.Email)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		user.UpdateFromGothUser(gothUser)

		props, err := db.Save(h.dbDriver, user)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		savedUser := models.UserFromProps(props)
		c.JSON(http.StatusCreated, savedUser)
		return
	})
}
