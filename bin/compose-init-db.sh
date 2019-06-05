#!/bin/bash

set -e
set -u

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
	    CREATE USER profile;
	    CREATE DATABASE profile;
	    GRANT ALL PRIVILEGES ON DATABASE profile TO profile;
EOSQL

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" profile < /docker-entrypoint-initdb.d/sql/init_profile.sql

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
	    CREATE USER profiletest;
	    CREATE DATABASE profiletest;
	    GRANT ALL PRIVILEGES ON DATABASE profiletest TO profiletest;
EOSQL
