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
`make test`

### run server
`make run`

### build docker-image
`make image`


If you clone this repo inside your `$GOPATH` you will likely need to set the environment variable `GO111MODULE=on`

## Endpoints (preliminary draft)

### Auth (with oauth2) 
visiting `/auth/:provider` in the browser will redirect the user to the specified provider (so far only `google` is implemented)

if the authentication flow was successful the user will be redirected to `<auth_origin_url (for now http://localhost:9876)>/login_redirect?jwt=<jwt token>` 

For requests to routes that need authentication, the jwt-token has to be included in the `Authorization` header as a bearer token:
(`Authorization: Bearer <jwt-token>`)

TODOS:

- [ ] add query param `auth_origin_url` to `/auth/:provider` to dynamically set the redirect on successful login

### Events 
```
 GET    /events/:uid              --> github.com/alexmorten/events-api/actions.(*ActionHandler).getEvent-fm (4 handlers)
 GET    /events                   --> github.com/alexmorten/events-api/actions.(*ActionHandler).getEvents-fm (4 handlers)
 POST   /events                   --> github.com/alexmorten/events-api/actions.(*ActionHandler).postEvents-fm (4 handlers)
```
TODOS:
- [ ] finish event routes, add needed properties
 add CRUD routes for: 
 - [ ] `clubs`
 - [ ] `groups` 
 - [ ] `sports` 
 - [ ] `tags`
 
 add additional routes for:
 - [ ] user verification
 
 

