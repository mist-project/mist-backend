#!/bin/bash

# ---- TENANT ACCOUNT ----
TENANT_USERNAME=$(op item get $TENANT_ACCOUNT_ID --vault "$OP_VAULT" --fields TENANT_USERNAME --reveal)
TENANT_PRIVATE_SSH_KEY=$(op item get $TENANT_ACCOUNT_ID --vault "$OP_VAULT" --fields TENANT_PRIVATE_SSH_KEY | tr -d '"')
TENANT_ADDRESS=$(op item get $TENANT_ACCOUNT_ID --vault "$OP_VAULT" --fields TENANT_ADDRESS)

# ---- BACKEND ACCOUNT ----
LOG_LEVEL=$(op item get $BACKEND_ACCOUNT_ID --vault "$OP_VAULT" --fields LOG_LEVEL --reveal)

REDIS_HOSTNAME=$(op item get $BACKEND_ACCOUNT_ID --vault "$OP_VAULT" --fields REDIS_HOSTNAME --reveal)
REDIS_PORT=$(op item get $BACKEND_ACCOUNT_ID --vault "$OP_VAULT" --fields REDIS_PORT --reveal)
REDIS_USERNAME=$(op item get $BACKEND_ACCOUNT_ID --vault "$OP_VAULT" --fields REDIS_USERNAME --reveal)
REDIS_PASSWORD=$(op item get $BACKEND_ACCOUNT_ID --vault "$OP_VAULT" --fields REDIS_PASSWORD --reveal)
REDIS_DB=$(op item get $BACKEND_ACCOUNT_ID --vault "$OP_VAULT" --fields REDIS_DB --reveal)
REDIS_NOTIFICATION_CHANNEL=$(op item get $BACKEND_ACCOUNT_ID --vault "$OP_VAULT" --fields REDIS_NOTIFICATION_CHANNEL --reveal)

DATABASE_URL=$(op item get $BACKEND_ACCOUNT_ID --vault "$OP_VAULT" --fields DATABASE_URL --reveal)

MIST_API_JWT_SECRET_KEY=$(op item get $BACKEND_ACCOUNT_ID --vault "$OP_VAULT" --fields MIST_API_JWT_SECRET_KEY --reveal)
MIST_API_JWT_AUDIENCE=$(op item get $BACKEND_ACCOUNT_ID --vault "$OP_VAULT" --fields MIST_API_JWT_AUDIENCE --reveal)
MIST_API_JWT_ISSUER=$(op item get $BACKEND_ACCOUNT_ID --vault "$OP_VAULT" --fields MIST_API_JWT_ISSUER --reveal)

APP_PORT=$(op item get $BACKEND_ACCOUNT_ID --vault "$OP_VAULT" --fields APP_PORT --reveal)

# Define file paths
KEY_FILE="key.pem"
INVENTORY_FILE="ansible/inventory/hosts.ini"

# Create ansible inventory directory and temporary file
mkdir -p ansible/inventory
touch $INVENTORY_FILE
echo "[mist-service]" >> "$INVENTORY_FILE"
echo $TENANT_ADDRESS >> "$INVENTORY_FILE"

# Create PRIVATE_SSH_KEY temporary file
touch $KEY_FILE
chmod 600 "$KEY_FILE"
echo -e "$TENANT_PRIVATE_SSH_KEY" >> "$KEY_FILE"


# Create tmporary environment variables file
touch .tmpenvs
echo "export TENANT_USERNAME=$TENANT_USERNAME" >> ".tmpenvs"

echo "export LOG_LEVEL=$LOG_LEVEL" >> ".tmpenvs"
echo "export REDIS_HOSTNAME=$REDIS_HOSTNAME" >> ".tmpenvs"
echo "export REDIS_PORT=$REDIS_PORT" >> ".tmpenvs"
echo "export REDIS_USERNAME=$REDIS_USERNAME" >> ".tmpenvs"
echo "export REDIS_PASSWORD=\"$REDIS_PASSWORD\"" >> ".tmpenvs"
echo "export REDIS_DB=$REDIS_DB" >> ".tmpenvs"
echo "export REDIS_NOTIFICATION_CHANNEL=$REDIS_NOTIFICATION_CHANNEL" >> ".tmpenvs"
echo "export DATABASE_URL=$DATABASE_URL" >> ".tmpenvs"

echo "export MIST_API_JWT_SECRET_KEY=\"$MIST_API_JWT_SECRET_KEY\"" >> ".tmpenvs"
echo "export MIST_API_JWT_AUDIENCE=\"$MIST_API_JWT_AUDIENCE\"" >> ".tmpenvs"
echo "export MIST_API_JWT_ISSUER=\"$MIST_API_JWT_ISSUER\"" >> ".tmpenvs"
echo "export APP_PORT=$APP_PORT" >> ".tmpenvs"

echo "export GOOSE_DRIVER=postgres" >> ".tmpenvs"
echo "export GOOSE_DBSTRING=${DATABASE_URL}" >> ".tmpenvs"
echo "export GOOSE_MIGRATION_DIR=src/psql_db/migrations" >> ".tmpenvs"