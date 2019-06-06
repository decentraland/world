# World

## Components

- Identity service
- Profile service
- Communications coordinator
- Communications server

## Running

The easiest way to run a world node is to use the provided docker-compose file. Remeber to set your gopath (export GOPATH=$HOME/go)

## Cli Usage

Start a bot that will walk around the world and send messages:
```
build/cli_bot --email= --password= --auth0Domain= --auth0ClientID= --auth0ClientSecret= --auth0Audience= --keyPath=  authURL=
```

Store a given json as user's profile:
```
cat profile | build/cli_profile --store --email= --password= --auth0Domain= --auth0ClientID= --auth0ClientSecret= --auth0Audience= --keyPath= --profileURL= authURL=
```

Retrieve user's profile:
```
build/cli_profile --retrieve --email= --password= --auth0Domain= --auth0ClientID= --auth0ClientSecret= --auth0Audience= --keyPath= --profileURL= authURL=
```
