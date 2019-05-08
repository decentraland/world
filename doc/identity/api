
An user is expected to refresh his access tokens using a user token provided
by auth0. The flow is simple, it must login if there's no user_token or it is
expired against auth0. Then he queries new tokens as needed using his
user_token.
It also provides methods for checking current status

GET  /login/:client_id

POST /:user/token
     Requires an user token given by auth0
     Requires an ephemeral public key
     Returns an access token signed by auth-service

GET  /status
     Returns current server status

GET  /pub_key
     Returns current used public key
