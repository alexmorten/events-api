package db

import (
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
