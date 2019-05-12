package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/alexmorten/events-api/search"

	"github.com/alexmorten/events-api/db"

	cors "github.com/rs/cors/wrapper/gin"

	"github.com/alexmorten/events-api/models"

	"github.com/alexmorten/events-api/actions"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

//Server is the outer most shell of the application
//responsible for serving http
type Server struct {
	config ServerConfig
	Engine *gin.Engine
}

//ServerConfig contains all configuration for the Server
type ServerConfig struct {
	Port                  int
	Neo4jAddress          string
	ElasticsearchAddress  string
	LazyInitializeElastic bool
}

//DefaultServerConfig ...
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Port:                  3000,
		Neo4jAddress:          "bolt://localhost:7687",
		ElasticsearchAddress:  "http://0.0.0.0:9200",
		LazyInitializeElastic: true,
	}
}

//NewServer returns a server listening on the specified address
func NewServer(config ServerConfig) *Server {
	return &Server{
		config: config,
	}
}

//Init the Server
func (s *Server) Init() {
	dbDriver := db.Driver(s.config.Neo4jAddress)
	db.MustCreateConstraints(dbDriver)

	searchClient := s.mustCreateSearchClient()

	actionHandler := actions.NewActionHandler(dbDriver, searchClient)

	s.Engine = gin.Default()
	s.Engine.Use(cors.AllowAll())
	rootGroup := s.Engine.Group("/", jwtHandler)
	actionHandler.RegisterAuthRoutes(rootGroup.Group("auth"))
	actionHandler.RegisterClubRoutes(rootGroup.Group("clubs"))
	actionHandler.RegisterGroupRoutes(rootGroup.Group("groups"))
	actionHandler.RegisterEventRoutes(rootGroup.Group("events"))
	actionHandler.RegisterSportRoutes(rootGroup.Group("sports"))
}

//Run the Server
func (s *Server) Run() {
	log.Fatal(s.Engine.Run(fmt.Sprintf(":%d", s.config.Port)))
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

func (s *Server) mustCreateSearchClient() *search.Client {
	client, err := search.NewClient(s.config.ElasticsearchAddress, s.config.LazyInitializeElastic)
	if err != nil {
		panic(err)
	}
	return client
}

func tokenFromBearer(bearer string) string {
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}
	return ""
}
