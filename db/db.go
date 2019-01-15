package db

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/neo4j/neo4j-go-driver/neo4j"
)

//NewDB creates a connection to neo4j
func NewDB() neo4j.Driver {
	driver, err := neo4j.NewDriver("bolt://localhost:7687", neo4j.BasicAuth("username", "password", ""))
	if err != nil {
		panic(err) // handle error
	}
	return driver
}

//Model that tells the database if it has to create or update a node
type Model interface {
	Created() bool
	NodeName() string
}

//Save the model to the database
func Save(dbDriver neo4j.Driver, model Model) (props map[string]interface{}, err error) {
	dbSession, err := dbDriver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return nil, err
	}

	var result neo4j.Result
	if model.Created() {
		result, err = dbSession.Run(fmt.Sprintf("create (n:%v {%v}) return properties(n)", model.NodeName(), NeoPropString(model)), MarshalNeoFields(model))
	} else {
		result, err = dbSession.Run(fmt.Sprintf("match (n:%v {uid: $uid}) set n = {%v} return properties(n)", model.NodeName(), NeoPropString(model)), MarshalNeoFields(model))
	}
	if err != nil {
		return nil, err
	}
	if result.Next() {
		record := result.Record()
		propInterface, ok := record.Get("properties(n)")
		if ok {
			props, ok := propInterface.(map[string]interface{})
			if ok {
				return props, nil
			}
		}
	}
	return nil, errors.New("saving node went wrong")
}

//CreateBy creates the model node together with a relationship to a user with the given id
func CreateBy(dbDriver neo4j.Driver, model Model, userUID uuid.UUID) (props map[string]interface{}, err error) {
	dbSession, err := dbDriver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return nil, err
	}
	neoFields := MarshalNeoFields(model)
	neoFields["user_uid"] = userUID.String()
	result, err := dbSession.Run(fmt.Sprintf("match (u:User {uid: $user_uid}) create (n:%v {%v})-[r:CREATED_BY]->(u) return properties(n)", model.NodeName(), NeoPropString(model)), neoFields)
	if err != nil {
		return nil, err
	}
	if result.Next() {
		record := result.Record()
		propInterface, ok := record.Get("properties(n)")
		if ok {
			props, ok := propInterface.(map[string]interface{})
			if ok {
				return props, nil
			}
		}
	}
	return nil, errors.New("creating node went wrong")
}

//FindNode props for uid
func FindNode(dbDriver neo4j.Driver, uid string) (props map[string]interface{}, err error) {
	dbSession, err := dbDriver.Session(neo4j.AccessModeRead)
	if err != nil {
		return nil, err
	}
	result, err := dbSession.Run("match (n {uid: $uid}) return properties(n)", map[string]interface{}{"uid": uid})
	if err != nil {
		return nil, err
	}
	if result.Next() {
		record := result.Record()
		propInterface, ok := record.Get("properties(n)")
		if ok {
			props, ok := propInterface.(map[string]interface{})
			if ok {
				return props, nil
			}
		}
	}
	return nil, nil
}

//FindRelation between two nodes
func FindRelation(dbDriver neo4j.Driver, fromNodeUID, toNodeUID, relationName string) (props map[string]interface{}, err error) {
	dbSession, err := dbDriver.Session(neo4j.AccessModeRead)
	if err != nil {
		return nil, err
	}

	result, err := dbSession.Run(
		fmt.Sprintf("match (fromNode {uid: $from_uid})-[r:%v]->(toNode {uid: $to_uid}) return properties(r)", relationName),
		map[string]interface{}{"from_uid": fromNodeUID, "to_uid": toNodeUID},
	)
	if err != nil {
		return nil, err
	}
	if result.Next() {
		record := result.Record()
		propInterface, ok := record.Get("properties(r)")
		if ok {
			props, ok := propInterface.(map[string]interface{})
			if ok {
				return props, nil
			}
		}
	}
	return nil, nil
}
