package models

import (
	"github.com/alexmorten/events-api/db"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

//ClubAttributes ...
type ClubAttributes struct {
	Name string `json:"name" neo:"name"`
}

//Club ...
type Club struct {
	Model
	ClubAttributes
}

//NewClub ...
func NewClub() *Club {
	return &Club{
		Model: newModel(),
	}
}

//FindClub with its uid
func FindClub(dbDriver neo4j.Driver, ClubUID string) (*Club, error) {
	props, err := db.FindNode(dbDriver, ClubUID)
	if err != nil {
		return nil, err
	}

	return ClubFromProps(props), nil
}

//NodeName is the label of event-nodes in the database
func (c *Club) NodeName() string {
	return "Club"
}

//ClubFromProps tries to get struct fields from the neo4j record
func ClubFromProps(props map[string]interface{}) *Club {
	if props == nil {
		return nil
	}

	club := &Club{}

	db.UnmarshalNeoFields(club, props)
	return club
}

//AddAdminToClub ...
func AddAdminToClub(dbDriver neo4j.Driver, clubUID, userUID uuid.UUID) error {
	_, err := db.CreateRelation(dbDriver, userUID, clubUID, UserAdministersClub)
	return err
}
