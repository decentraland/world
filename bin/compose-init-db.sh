#!/bin/bash

set -e
set -u

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
	    CREATE DATABASE statsdb;
	    GRANT ALL PRIVILEGES ON DATABASE statsdb TO $POSTGRES_USER
EOSQL

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" statsdb < /docker-entrypoint-initdb.d/sql/init_stats.sql
