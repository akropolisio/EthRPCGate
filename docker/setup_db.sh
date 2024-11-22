#!/bin/bash

# Database connection parameters
HOST="$SQL_HOST"
PORT="$SQL_PORT"
USER="$SQL_USER"
PASSWORD="$SQL_PASSWORD"
DBNAME="$SQL_DBNAME"
SSLMODE="$SQL_SSL"

# Export password to avoid prompting
export PGPASSWORD=$PASSWORD

# Execute DROP DATABASE commands
psql -h $HOST -p $PORT -U $USER -d $DBNAME -c "DROP DATABASE IF EXISTS ${DBNAME};"

# Execute CREATE DATABASE commands
psql -h $HOST -p $PORT -U $USER -d $DBNAME -c "CREATE DATABASE ${DBNAME};"

# Grant privileges
GRANT_COMMANDS="
GRANT ALL PRIVILEGES ON DATABASE ${DBNAME} TO ${USER};
"

# Execute grant commands
psql -h $HOST -p $PORT -U $USER -d $DBNAME -c "$GRANT_COMMANDS"

# Unset the password for security reasons
unset PGPASSWORD

echo "PostgreSQL setup completed."
echo "Database: ${DBNAME}"
echo "User: ${USER}"