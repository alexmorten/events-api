package models

import (
	"github.com/alexmorten/events-api/db"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

//EventAttributes ...
type EventAttributes struct {
	Name string `json:"name" neo:"name"`
}

//Event ...
type Event struct {
	Model
	EventAttributes
}

//NewEvent ...
func NewEvent() *Event {
	return &Event{
		Model: newModel(),
	}
}

//FindEvent with its uid
func FindEvent(dbDriver neo4j.Driver, eventUID string) (*Event, error) {
	props, err := db.FindNode(dbDriver, eventUID)
	if err != nil {
		return nil, err
	}

	return EventFromProps(props), nil
}

//NodeName is the label of event-nodes in the database
func (e *Event) NodeName() string {
	return "Event"
}

//EventFromProps tries to get struct fields from the neo4j record
func EventFromProps(props map[string]interface{}) *Event {
	if props == nil {
		return nil
	}

	event := &Event{}

	db.UnmarshalNeoFields(event, props)
	return event
}

//CanBeEditedBy user with given uid
func (e *Event) CanBeEditedBy(dbDriver neo4j.Driver, userUID uuid.UUID) bool {
	relationProps, err := db.FindRelation(dbDriver, e.UID.String(), userUID.String(), "CREATED_BY")
	return err == nil && relationProps != nil
}
