# World

## Components

- World definition service
- Identity service
- Profile service
- Communications coordinator
- Communications server

## Running

The easiest way to run a world node is to use the provided docker-compose file. Remeber to set your GOPATH first (export GOPATH=$HOME/go)

## Cli Usage

You will need to generate a key first (it will represent the browser's local storage)
```
build/cli_keygen --curve s256 --outputDir ./keys
```

Start a bot that will walk around the world and send messages:
```
build/cli_bot --email= --password= --auth0ClientSecret= --keyPath=./keys/client.key
```

Store a given json as user's profile:
```
cat profile | build/cli_profile --store --email= --password= --auth0ClientSecret= --keyPath./keys/client.key=
```

Retrieve user's profile:
```
build/cli_profile --retrieve --email= --password= --auth0ClientSecret= --keyPath=./keys/client.key
```

Note:

To be able to use this tool locally if you are using docker-compose you may want to add this to your /etc/hosts:

```
127.0.0.1 profile identity coordinator
```
