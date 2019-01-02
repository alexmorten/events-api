package models

import "fmt"

//Event ...
type Event struct {
	Model

	Name string `json:"name"`
}

//EventNeoPropString can be used in a neo4j query to define all properties of an Event node
func EventNeoPropString() string {
	eventProps := "name: $name"
	return fmt.Sprintf("%v, %v", modelPropsString(), eventProps)
}

//NewEvent ...
func NewEvent() *Event {
	return &Event{
		Model: newModel(),
	}
}

//EventFromProps tries to get struct fields from the neo4j record
func EventFromProps(props map[string]interface{}) *Event {
	event := &Event{
		Model: modelFromProps(props),
	}

	if value, ok := props["name"]; ok {
		event.Name = value.(string)
	}
	return event
}

//NeoPropMap returns a map that can be passed to a neo4j session.Run call
func (e *Event) NeoPropMap() map[string]interface{} {
	return map[string]interface{}{
		"uid":        e.UID.String(),
		"name":       e.Name,
		"created_at": e.CreatedAt.Format("2006-01-02 15:04:05.999999999 -0700 MST"),
	}
}
