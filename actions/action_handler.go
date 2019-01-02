package actions

import (
	"github.com/alexmorten/events-api/db"
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
