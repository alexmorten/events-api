package db

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/neo4j/neo4j-go-driver/neo4j"
)

//Driver creates a connection to neo4j
func Driver(neo4jAddress string) neo4j.Driver {
	driver, err := neo4j.NewDriver(neo4jAddress, neo4j.NoAuth())
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

	var record neo4j.Record
	if model.Created() {
		record, err = neo4j.Single(dbSession.Run(fmt.Sprintf("create (n:%v {%v}) return properties(n)", model.NodeName(), NeoPropString(model)), MarshalNeoFields(model)))
	} else {
		record, err = neo4j.Single(dbSession.Run(fmt.Sprintf("match (n:%v {uid: $uid}) set n += {%v} return properties(n)", model.NodeName(), NeoPropString(model)), MarshalNeoFields(model)))
	}
	if err != nil {
		return nil, err
	}

	propInterface, ok := record.Get("properties(n)")
	if ok {
		props, ok := propInterface.(map[string]interface{})
		if ok {
			return props, nil
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
	record, err := neo4j.Single(dbSession.Run(fmt.Sprintf("match (u:User {uid: $user_uid}) create (n:%v {%v})-[r:CREATED_BY]->(u) return properties(n)", model.NodeName(), NeoPropString(model)), neoFields))
	if err != nil {
		return nil, err
	}

	propInterface, ok := record.Get("properties(n)")
	if ok {
		props, ok := propInterface.(map[string]interface{})
		if ok {
			return props, nil
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
	record, err := neo4j.Single(dbSession.Run("match (n {uid: $uid}) return properties(n)", map[string]interface{}{"uid": uid}))
	if err != nil {
		return nil, err
	}
	propInterface, ok := record.Get("properties(n)")
	if ok {
		props, ok := propInterface.(map[string]interface{})
		if ok {
			return props, nil
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

	record, err := neo4j.Single(dbSession.Run(
		fmt.Sprintf("match (fromNode {uid: $from_uid})-[r:%v]->(toNode {uid: $to_uid}) return properties(r)", relationName),
		map[string]interface{}{"from_uid": fromNodeUID, "to_uid": toNodeUID},
	))
	if err != nil {
		return nil, err
	}

	propInterface, ok := record.Get("properties(r)")
	if ok {
		props, ok := propInterface.(map[string]interface{})
		if ok {
			return props, nil
		}
	}
	return nil, nil
}

//DeleteNode with given uid, detaching all relationships attached to it
func DeleteNode(dbDriver neo4j.Driver, uid string) (err error) {
	dbSession, err := dbDriver.Session(neo4j.AccessModeRead)
	if err != nil {
		return err
	}

	_, err = dbSession.Run("match (n {uid: $uid}) detach delete n", map[string]interface{}{"uid": uid})
	return err
}

//CreateRelation creates the model node together with a relationship to a user with the given id
func CreateRelation(dbDriver neo4j.Driver, fromUID, toUID uuid.UUID, relationName string) (props map[string]interface{}, err error) {
	dbSession, err := dbDriver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return nil, err
	}

	record, err := neo4j.Single(dbSession.Run(
		fmt.Sprintf(
			`
			match (from_n {uid: $from_uid}), (to_n {uid: $to_uid})
			create (from_n)-[r:%v]->(to_n)
			return properties(r)
			`,
			relationName,
		),
		map[string]interface{}{
			"from_uid": fromUID.String(),
			"to_uid":   toUID.String(),
		},
	))
	if err != nil {
		return nil, err
	}

	propInterface, ok := record.Get("properties(r)")
	if ok {
		props, ok := propInterface.(map[string]interface{})
		if ok {
			return props, nil
		}
	}
	return nil, errors.New("creating relation went wrong")
}
