#!/bin/sh

set -eu

ADMIN_DB="${POSTGRES_DB:-postgres}"
APP_DB="${DB_NAME:-cloud_storage}"
APP_USER="${DB_USERNAME:-soggy}"
APP_PASSWORD="${DB_PASSWORD:-password}"

role_exists="$(psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$ADMIN_DB" -tAc "SELECT 1 FROM pg_roles WHERE rolname = '$APP_USER'")"
if [ "$role_exists" != "1" ]; then
  psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$ADMIN_DB" \
    -v app_user="$APP_USER" \
    -v app_password="$APP_PASSWORD" <<'SQL'
CREATE ROLE :"app_user" WITH LOGIN PASSWORD :'app_password';
SQL
fi

db_exists="$(psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$ADMIN_DB" -tAc "SELECT 1 FROM pg_database WHERE datname = '$APP_DB'")"
if [ "$db_exists" != "1" ]; then
  psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$ADMIN_DB" \
    -v app_db="$APP_DB" \
    -v app_user="$APP_USER" <<'SQL'
CREATE DATABASE :"app_db" OWNER :"app_user";
SQL
fi

psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$APP_DB" \
  -v app_db="$APP_DB" \
  -v app_user="$APP_USER" <<'SQL'
ALTER DATABASE :"app_db" OWNER TO :"app_user";
GRANT ALL PRIVILEGES ON DATABASE :"app_db" TO :"app_user";
ALTER SCHEMA public OWNER TO :"app_user";
GRANT ALL ON SCHEMA public TO :"app_user";
SQL
