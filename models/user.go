package models

import (
	"errors"
	"time"

	"github.com/alexmorten/events-api/db"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

//User ...
type User struct {
	Model

	Admin bool `json:"admin" neo:"admin"`

	//goth-user attributes
	Provider    string `json:"provider" neo:"provider"`
	Email       string `json:"email" neo:"email"`
	Name        string `json:"name" neo:"name"`
	FirstName   string `json:"first_name" neo:"first_name"`
	LastName    string `json:"last_name" neo:"last_name"`
	NickName    string `json:"nick_name" neo:"nick_name"`
	Description string `json:"description" neo:"description"`
	UserID      string `json:"user_id" neo:"user_id"`
	AvatarURL   string `json:"avatar_url" neo:"avatar_url"`
	Location    string `json:"location" neo:"location"`
}

//PublicUserAttributes that can be shared with the users
type PublicUserAttributes struct {
	UID       uuid.UUID `json:"uid"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
}

//NewUser ...
func NewUser() *User {
	return &User{
		Model: newModel(),
	}
}

//FindOrCreateUserByEmail returns a user if it finds a user in the DB or creates a new user (that is not yet saved to the DB!).
//returns error db related tasks fail
func FindOrCreateUserByEmail(dbDriver neo4j.Driver, email string) (*User, error) {
	userOrNil, err := FindUserByEmail(dbDriver, email)
	if err != nil {
		return nil, err
	}

	if userOrNil != nil {
		return userOrNil, nil
	}

	createdUser := NewUser()
	createdUser.Email = email
	return createdUser, nil
}

//FindUserByEmail returns a pointer to a user or nil if no user was found
func FindUserByEmail(dbDriver neo4j.Driver, email string) (*User, error) {
	session, err := dbDriver.Session(neo4j.AccessModeRead)
	if err != nil {
		return nil, err
	}
	result, err := session.Run("match (n:User {email: $email}) return properties(n)", map[string]interface{}{"email": email})
	if err != nil {
		return nil, err
	}
	if result.Next() {
		record := result.Record()
		propInterface, ok := record.Get("properties(n)")
		if ok {
			props, ok := propInterface.(map[string]interface{})
			if ok {
				user := UserFromProps(props)
				return user, nil
			}
		}

	}
	return nil, nil
}

//UserFromProps tries to get struct fields from the neo4j record
func UserFromProps(props map[string]interface{}) *User {
	if props == nil {
		return nil
	}

	user := &User{}
	db.UnmarshalNeoFields(user, props)
	return user
}

//NodeName is the label of user-nodes in the database
func (u *User) NodeName() string {
	return "User"
}

//UpdateFromGothUser updates the user from the provided goth.User
func (u *User) UpdateFromGothUser(gothUser goth.User) {
	u.Provider = gothUser.Provider
	u.Email = gothUser.Email
	u.Name = gothUser.Name
	u.FirstName = gothUser.FirstName
	u.LastName = gothUser.LastName
	u.NickName = gothUser.NickName
	u.Description = gothUser.Description
	u.UserID = gothUser.UserID
	u.AvatarURL = gothUser.AvatarURL
	u.Location = gothUser.Location
}

//PublicAttributes of user
func (u *User) PublicAttributes() PublicUserAttributes {
	return PublicUserAttributes{
		UID:       u.UID,
		Name:      u.Name,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
	}
}

//Claim returns a claim for jwt
func (u *User) Claim() *UserClaim {
	return &UserClaim{
		UID:      u.UID,
		IssuedAt: time.Now(),
		Admin:    u.Admin,
	}
}

//UserClaim is a struct representing the claim issued in the jwt on authentication
type UserClaim struct {
	UID      uuid.UUID
	Admin    bool
	IssuedAt time.Time
}

//UserClaimFromMap returns a UserClaim if the provided map contains the correct fields
func UserClaimFromMap(m map[string]interface{}) (*UserClaim, error) {
	claim := &UserClaim{}
	uidInterface, ok := m["uid"]
	if !ok {
		return nil, errors.New("uid not present in claim")
	}
	uidString, ok := uidInterface.(string)
	if !ok {
		return nil, errors.New("uid not string in claim")
	}
	uid, err := uuid.Parse(uidString)
	if err != nil {
		return nil, err
	}
	claim.UID = uid

	issuedAtInterface, ok := m["issued_at"]
	if !ok {
		return nil, errors.New("issued_at not present in claim")
	}
	issuedAtString, ok := issuedAtInterface.(string)
	if !ok {
		return nil, errors.New("issued_at not string in claim")
	}
	issuedAt, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", issuedAtString)
	if err != nil {
		return nil, err
	}
	claim.IssuedAt = issuedAt

	adminInterface, ok := m["admin"]
	if ok {
		admin, ok := adminInterface.(bool)
		if ok {
			claim.Admin = admin
		}
	}

	return claim, nil
}

//Map of claims in the jwt
func (c *UserClaim) Map() jwt.MapClaims {
	return jwt.MapClaims{
		"uid":       c.UID.String(),
		"issued_at": c.IssuedAt.Format("2006-01-02 15:04:05.999999999 -0700 MST"),
		"admin":     c.Admin,
	}
}
