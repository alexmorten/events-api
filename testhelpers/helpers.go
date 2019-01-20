package testhelpers

import (
	"net/http"
	"os"

	"github.com/alexmorten/events-api/db"
	"github.com/alexmorten/events-api/models"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

//Clear is useful to keep tests side-effect free
func Clear(dbDriver neo4j.Driver) {
	sess, err := dbDriver.Session(neo4j.AccessModeWrite)
	panicOnErr(err)
	_, err = sess.Run("match(n) detach delete n", nil)
	panicOnErr(err)
}

//CreateSomeUser and return it
func CreateSomeUser(dbDriver neo4j.Driver) *models.User {
	user := models.NewUser()

	_, err := db.Save(dbDriver, user)
	panicOnErr(err)
	return user
}

//CreateAdminUser and return it
func CreateAdminUser(dbDriver neo4j.Driver) *models.User {
	user := models.NewUser()
	user.Admin = true
	_, err := db.Save(dbDriver, user)
	panicOnErr(err)
	return user
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

//AddSomeAuthorization creates a user and adds the Authorization header to the request
func AddSomeAuthorization(dbDriver neo4j.Driver, req *http.Request) {
	user := CreateSomeUser(dbDriver)
	AddAuthorizationHeader(req, user)
}

//AddAuthorizationHeader with a jwt token from the user
func AddAuthorizationHeader(req *http.Request, user *models.User) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, user.Claim().Map())
	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	panicOnErr(err)
	req.Header.Add("Authorization", "Bearer "+tokenString)
}
