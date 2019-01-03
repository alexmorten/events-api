package models

import (
	"github.com/alexmorten/events-api/db"
)

//Event ...
type Event struct {
	Model

	Name string `json:"name" neo:"name"`
}

//NewEvent ...
func NewEvent() *Event {
	return &Event{
		Model: newModel(),
	}
}

//NodeName is the label of event-nodes in the database
func (e *Event) NodeName() string {
	return "Event"
}

//EventFromProps tries to get struct fields from the neo4j record
func EventFromProps(props map[string]interface{}) *Event {
	event := &Event{}

	db.UnmarshalNeoFields(event, props)
	return event
}
