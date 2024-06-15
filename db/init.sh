#!/bin/bash

psql -v ON_ERROR_STOP=1 --username postgres <<-EOSQL
    CREATE DATABASE engineer_test;
EOSQL

psql -v ON_ERROR_STOP=1 --username postgres -d engineer_test -a -f /docker-entrypoint-initdb.d/schema.sql <<-EOSQL
EOSQL
