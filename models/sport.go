package models

import (
	"github.com/alexmorten/events-api/db"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

//SportAttributes ...
type SportAttributes struct {
	Name string `json:"name" neo:"name"`
}

//Sport ...
type Sport struct {
	Model
	SportAttributes
}

//NewSport ...
func NewSport() *Sport {
	return &Sport{
		Model: newModel(),
	}
}

//FindSport with its uid
func FindSport(dbDriver neo4j.Driver, SportUID string) (*Sport, error) {
	props, err := db.FindNode(dbDriver, SportUID)
	if err != nil {
		return nil, err
	}

	return SportFromProps(props), nil
}

//NodeName is the label of event-nodes in the database
func (s *Sport) NodeName() string {
	return "Sport"
}

//SportFromProps tries to get struct fields from the neo4j record
func SportFromProps(props map[string]interface{}) *Sport {
	if props == nil {
		return nil
	}

	sport := &Sport{}

	db.UnmarshalNeoFields(sport, props)
	return sport
}
