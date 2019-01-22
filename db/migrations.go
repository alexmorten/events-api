package db

import (
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

//MustCreateConstraints on neo4j database or panic
func MustCreateConstraints(dbDriver neo4j.Driver) {
	dbSession, err := dbDriver.Session(neo4j.AccessModeWrite)
	if err != nil {
		panic(err)
	}

	panicOnErrSummary(dbSession.Run("CREATE CONSTRAINT ON (u:User) ASSERT u.uid IS UNIQUE", nil))
	panicOnErrSummary(dbSession.Run("CREATE CONSTRAINT ON (u:User) ASSERT u.email IS UNIQUE", nil))
	panicOnErrSummary(dbSession.Run("CREATE CONSTRAINT ON (e:Event) ASSERT e.uid IS UNIQUE", nil))
	panicOnErrSummary(dbSession.Run("CREATE CONSTRAINT ON (c:Club) ASSERT c.uid IS UNIQUE", nil))
	panicOnErrSummary(dbSession.Run("CREATE CONSTRAINT ON (s:Sport) ASSERT s.uid IS UNIQUE", nil))
}

func panicOnErrSummary(result neo4j.Result, err error) {
	if err != nil {
		panic(err)
	}
	if _, err = result.Summary(); err != nil {
		panic(err)
	}
}
