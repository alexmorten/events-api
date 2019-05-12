# events-api

## setup
you need a working go installation (>=1.11) -> https://golang.org/dl/

for development you need to add a `.env` file to the root of this repo with 
```
GOOGLE_CLIENT=...
GOOGLE_SECRET=...
SESSION_SECRET="1234567890"

```
### run tests
`neo4j-test` to start the test database container

wait a few seconds, then:

`make test`

### run server
`neo4j-dev` to start the devlopment database containers 

after a minute:

`make run`

### build docker-image
`make image`


If you clone this repo inside your `$GOPATH` you will likely need to set the environment variable `GO111MODULE=on`

## Endpoints (preliminary draft)

```
 GET    /clubs/:uid               --> github.com/alexmorten/events-api/actions.(*ActionHandler).getClub-fm (5 handlers)
 GET    /clubs                    --> github.com/alexmorten/events-api/actions.(*ActionHandler).getClubs-fm (5 handlers)
 PATCH  /clubs/:uid               --> github.com/alexmorten/events-api/actions.(*ActionHandler).updateClub-fm (5 handlers)
 POST   /clubs                    --> github.com/alexmorten/events-api/actions.(*ActionHandler).postClubs-fm (5 handlers)
 DELETE /clubs/:uid               --> github.com/alexmorten/events-api/actions.(*ActionHandler).deleteClub-fm (5 handlers)
 POST   /clubs/:uid/groups        --> github.com/alexmorten/events-api/actions.(*ActionHandler).postGroup-fm (5 handlers)
 GET    /clubs/:uid/groups        --> github.com/alexmorten/events-api/actions.(*ActionHandler).getGroups-fm (5 handlers)
 GET    /clubs/:uid/admins        --> github.com/alexmorten/events-api/actions.(*ActionHandler).getAdmins-fm (5 handlers)
 POST   /clubs/:uid/admins        --> github.com/alexmorten/events-api/actions.(*ActionHandler).postAdmins-fm (5 handlers)
 GET    /groups/:uid              --> github.com/alexmorten/events-api/actions.(*ActionHandler).getGroup-fm (5 handlers)
 GET    /groups/:uid/groups       --> github.com/alexmorten/events-api/actions.(*ActionHandler).getGroups-fm (5 handlers)
 PATCH  /groups/:uid              --> github.com/alexmorten/events-api/actions.(*ActionHandler).updateGroup-fm (5 handlers)
 POST   /groups/:uid/groups       --> github.com/alexmorten/events-api/actions.(*ActionHandler).postGroup-fm (5 handlers)
 DELETE /groups/:uid              --> github.com/alexmorten/events-api/actions.(*ActionHandler).deleteGroup-fm (5 handlers)
 GET    /groups/:uid/admins       --> github.com/alexmorten/events-api/actions.(*ActionHandler).getGroupAdmins-fm (5 handlers)
 POST   /groups/:uid/admins       --> github.com/alexmorten/events-api/actions.(*ActionHandler).postGroupAdmins-fm (5 handlers)
 GET    /events/:uid              --> github.com/alexmorten/events-api/actions.(*ActionHandler).getEvent-fm (5 handlers)
 GET    /events                   --> github.com/alexmorten/events-api/actions.(*ActionHandler).getEvents-fm (5 handlers)
 PATCH  /events/:uid              --> github.com/alexmorten/events-api/actions.(*ActionHandler).updateEvent-fm (5 handlers)
 POST   /events                   --> github.com/alexmorten/events-api/actions.(*ActionHandler).postEvents-fm (5 handlers)
 DELETE /events/:uid              --> github.com/alexmorten/events-api/actions.(*ActionHandler).deleteEvent-fm (5 handlers)
 GET    /sports/:uid              --> github.com/alexmorten/events-api/actions.(*ActionHandler).getSport-fm (5 handlers)
 GET    /sports                   --> github.com/alexmorten/events-api/actions.(*ActionHandler).getSports-fm (5 handlers)
 PATCH  /sports/:uid              --> github.com/alexmorten/events-api/actions.(*ActionHandler).updateSport-fm (5 handlers)
 POST   /sports                   --> github.com/alexmorten/events-api/actions.(*ActionHandler).postSports-fm (5 handlers)
 DELETE /sports/:uid              --> github.com/alexmorten/events-api/actions.(*ActionHandler).deleteSport-fm (5 handlers)
```

### Auth (with oauth2) 
visiting `/auth/:provider` in the browser will redirect the user to the specified provider (so far only `google` is implemented)

if the authentication flow was successful the user will be redirected to `<auth_origin_url (for now http://localhost:9876)>/login_redirect?jwt=<jwt token>` 

For requests to routes that need authentication, the jwt-token has to be included in the `Authorization` header as a bearer token:
(`Authorization: Bearer <jwt-token>`)



TODOS:

- [ ] add query param `auth_origin_url` to `/auth/:provider` to dynamically set the redirect on successful login
- [ ] add `POST /clubs/:uid/events`
- [ ] add `POST /groups/:uid/events`
- [ ] add CRUD endpoints for tags

add additional routes for:
- [ ] user verification
 
 

# Swagger

## Swagger UI

- pull docker ``make swagger-ui-pull``
- start docker ``make swagger-ui-run``



## Swagger Editor

- pull docker ``make swagger-editor-pull``
- start docker ``make swagger-editor-run``


