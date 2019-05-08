# Communications

To run a local simulation you will need to:

```
$ make
$ build/coordinator --noopAuthEnabled
$ build/server --noopAuthEnabled --authMethod=noop
$ build/bots -n 10 --subscribe
```

You can optionally start an extra bot to print stats

```
$ build/bots -n 1 --subscribe --trackStats
```

# Profile Service

The service takes a postgresql connecting string as a parameter: `build/profile --connStr="postgres://<username>:<password>@<host>/<db>"

## Using postgresql with docker for dev

```
$ docker pull postgres
$ mkdir -p ~/docker/volumes/postgres
$ docker run --rm --name pg-docker -e POSTGRES_PASSWORD=docker -d -p 5432:5432 -v $HOME/docker/volumes/postgres:/var/lib/postgresql/data  postgres
```

Then create a database:

```
$ psql -h localhost -U postgres -d postgres

create database profiledb;
```

And the required tables:

```
psql -h localhost -U postgres -d  profiledb < internal/profile/db.sql
```

Finally, to run the service: `build/profile --connStr="postgres://postgres:docker@localhost/profiledb?sslmode=disable"`

# Identity

## Building the service

```go
$ make build
```

## Run the service locally

```go
$ ./build/identity 
```

By default this will start the service on the 9091 port.

To sign the tokens by default it will use the keys under `config/identity/defaultKeys`. 

### Key generation

You can generate your own keys

```go
$ ./build/keygen <PATH WHERE TO STORE THE GEN KEY>
```

To make the auth server use the generated set the  `PRIV_KEY_PATH` env variable to point to the generated private key `<PATH WHERE TO STORE THE GEN KEY>/generated.key`

 