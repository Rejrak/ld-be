#!/bin/bash
set -e

echo "Creating multiple databases: $POSTGRES_MULTIPLE_DATABASES"
for db in $(echo $POSTGRES_MULTIPLE_DATABASES | tr ',' ' '); do
  echo "Creating database: $db"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE "$db";
EOSQL
done