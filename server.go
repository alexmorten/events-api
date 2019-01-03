package api

import (
	"log"

	"github.com/alexmorten/events-api/actions"
	"github.com/gin-gonic/gin"
)

//Server is the outer most shell of the application
//responsible for serving http
type Server struct {
	address string
	engine  *gin.Engine
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
	s.engine = gin.Default()
	actionHandler.RegisterEventRoutes(s.engine.Group("/events"))
	actionHandler.RegisterAuthRoutes(s.engine.Group("/auth"))
}

//Run the Server
func (s *Server) Run() {
	log.Fatal(s.engine.Run(s.address))
}
