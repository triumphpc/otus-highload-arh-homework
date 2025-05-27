#!/bin/bash

set -e

echo "Waiting for PostgreSQL to become ready..."

export PGPASSWORD=$POSTGRES_PASSWORD

until pg_isready -U $POSTGRES_USER -h postgres; do
  sleep 1
done

echo "PostgreSQL is ready, starting data seeding..."

psql -v ON_ERROR_STOP=1 \
     -U $POSTGRES_USER \
     -h postgres \
     -d $POSTGRES_DB \
     -f /docker-entrypoint-initdb.d/seed.sql

echo "Data seeding completed successfully!"