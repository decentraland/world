Identity specifications summary

In this document I'll present a summary of the technical
specifications for the identity service, with good techincal details

These are some requirements or assumptions:

      - Users must login to worlds in decentraland. In each world, all
        users share the same view of the state of everything. In such
        a way that all users that will be seen on screen will be
        logged to the same world and viceversa.

      - There are going to be multiple services consuming the user's
        identity. We may call these services communication server,
        game server or just service

Our solution involves 4 actors:

      - user: person using a browser or mobile app
    
      - services: backend service that wants to identify the user

      - auth server: a server to be described here that will sign
        tokens for the user

      - auth0 server: it will save and check login credentials (like
        matching user and password)

The basic idea is that a user will authenticate against the auth
server using the auth0 protocol. The auth server will grant a json web
token (jwt) to the user which will be used against the services.

		 -------        -------------
		 |auth0| <----> |auth server|
		 -------        -------------
		    ^                 ^
		    |                 |
		    |     ------      |
		    ----> |user| <-----
			  ------
			    |
			    v
			---------
			|service|
			---------

The key point here is that while the service only trust the auth
server, there's no communication between these two.

----

For login against the auth server we will use the auth0 protocol. We
don't care really care here what auth0 protocol we use. The important
aspect of this is that the actual validation of username / passwords
happens is done by auth0. So when the user wants to login against the
our auth server protocol, it must accept to share his profile and from
there on we assume that the user is logged in against our auth
server. From here on we can ignore everything about the auth0 protocol
and server and define how the user, the auth server and the services
identify the user

So we need to define the flow of information in this chart

   -------------        ------        ---------
   |auth server| <----> |user| <----> |service|
   -------------        ------        ---------

When the user authenticates against the auth server, tha later
responds with a refresh token, a random long live string that will be
used by the user to identify against the auth server later.

The user creates a ecdsa key par.

The user uses the public key and the refresh token to request the auth
server a short lived json web token

The auth server creates a jwt with the user id, user mail, public key,
signs it with its own key and returns the jwt to the user

The user validates the jwt

The user creates a request with the jwt signed with its own key pair

The user request the service with the previous request

The service validates the jwt (expiration, signed by auth server)

The service validates the request (signed with the key that
corresponds to the public key in the jwt)

The user is authenticated in the service

---

There are a few points to consdired about this

The jwt is short lived (~60 seconds proposed here, but may
change). This implies that for logging out from a session a user must
invalidate a refresh token and that's it.

Replay attacks are prevented by the use of the ecdsa key, also called
ephemeral key
