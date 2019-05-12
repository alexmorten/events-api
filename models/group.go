package models

import (
	"fmt"

	"github.com/alexmorten/events-api/db"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

//GroupAttributes ...
type GroupAttributes struct {
	Name string `json:"name" neo:"name"`
}

//Group ...
type Group struct {
	Model
	GroupAttributes
}

//NewGroup ...
func NewGroup() *Group {
	return &Group{
		Model: newModel(),
	}
}

//FindGroup with its uid
func FindGroup(dbDriver neo4j.Driver, GroupUID string) (*Group, error) {
	props, err := db.FindNode(dbDriver, GroupUID)
	if err != nil {
		return nil, err
	}

	return GroupFromProps(props), nil
}

//NodeName is the label of event-nodes in the database
func (g *Group) NodeName() string {
	return "Group"
}

// AdministeredByUser , is the group Administered be the user?
func (g *Group) AdministeredByUser(dbDriver neo4j.Driver, userUID uuid.UUID) bool {
	user, err := FindUser(dbDriver, userUID.String())

	if err != nil {
		fmt.Println(err)
		return false
	}

	if user.Admin {
		return true
	}

	dbSession, err := dbDriver.Session(neo4j.AccessModeRead)
	if err != nil {
		return false
	}
	defer dbSession.Close()

	records, err := neo4j.Collect(dbSession.Run(
		fmt.Sprintf(
			"match (n:User {uid: $user_uid})-[:%v*1..10]->()<-[:%s*0..10]-(group:Group {uid: $group_uid}) return properties(n)",
			UserAdministersGroupOrClub,
			GroupBelongsToGroupOrClub),
		map[string]interface{}{"user_uid": user.UID.String(), "group_uid": g.UID.String()}))
	fmt.Println("records: ", records, " err: ", err)
	if err != nil {
		return false
	}

	return len(records) > 0
}

//GroupFromProps tries to get struct fields from the neo4j record
func GroupFromProps(props map[string]interface{}) *Group {
	if props == nil {
		return nil
	}

	Group := &Group{}

	db.UnmarshalNeoFields(Group, props)
	return Group
}

//AddAdminToGroup ...
func AddAdminToGroup(dbDriver neo4j.Driver, GroupUID, userUID uuid.UUID) error {
	_, err := db.CreateRelation(dbDriver, userUID, GroupUID, UserAdministersGroupOrClub)
	return err
}
