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
