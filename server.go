package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	cors "github.com/rs/cors/wrapper/gin"

	"github.com/alexmorten/events-api/models"

	"github.com/alexmorten/events-api/actions"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

//Server is the outer most shell of the application
//responsible for serving http
type Server struct {
	address string
	Engine  *gin.Engine
}

//NewServer returns a server listening on the specified address
func NewServer(address string) *Server {
	return &Server{
		address: address,
	}
}

//Init the Server
func (s *Server) Init() {
	actionHandler := actions.NewActionHandler()
	s.Engine = gin.Default()
	s.Engine.Use(cors.AllowAll())
	rootGroup := s.Engine.Group("/", jwtHandler)
	actionHandler.RegisterEventRoutes(rootGroup.Group("events"))
	actionHandler.RegisterAuthRoutes(rootGroup.Group("auth"))
}

//Run the Server
func (s *Server) Run() {
	log.Fatal(s.Engine.Run(s.address))
}

func jwtHandler(c *gin.Context) {
	bearer := c.GetHeader("Authorization")
	tokenString := tokenFromBearer(bearer)
	if tokenString != "" {
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// validate the alg is what we expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.AbortWithError(http.StatusUnauthorized, errors.New("jwt token invalid"))
			return
		}
		userClaim, err := models.UserClaimFromMap(claims)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}
		c.Set("currentUserClaim", userClaim)
	}

	c.Next()
}

func tokenFromBearer(bearer string) string {
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}
	return ""
}
