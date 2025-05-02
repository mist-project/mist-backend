# Mist Backend

## Installation

### Install homebrew

```
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

echo >> ~/.bashrc
echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> ~/.bashrc
eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
```

### Install GO

As of right now we're going to be using GO version `1.23.4`
https://go.dev/doc/install

```
# This example is for LINUX

# Remove any existing GO installation and install the one downloaded
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz

# Add /usr/local/go/bin to the PATH environment variable.
export PATH=$PATH:/usr/local/go/bin

# Verify successfull install
go version
```

### Install PostgreSQL

```
# This is for LINUX ARM64 architecture, using version 17.2

sudo apt install curl ca-certificates
sudo install -d /usr/share/postgresql-common/pgdg
sudo curl -o /usr/share/postgresql-common/pgdg/apt.postgresql.org.asc --fail https://www.postgresql.org/media/keys/ACCC4CF8.asc
sudo sh -c 'echo "deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] https://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
sudo apt update
sudo apt -y install postgresql

TODO: add/figure out db roles
# After installation create new database called mist
sudo -u postgres psql

CREATE DATABASE mist;
```

#### Install DB migration plugin

go install -tags='no_clickhouse no_libsql no_mssql no_mysql no_vertica no_ydb' github.com/pressly/goose/v3/cmd/goose@latest

### Install SQL generator tool

sudo snap install sqlc

### Install protobuf compiler

```

brew install bufbuild/buf/buf

# On linux install, install version 3.12.4
apt install -y protobuf-compiler # Idk if you need this anymore

# Install go plugin for the protocol compiler, version 1.35.2
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Install plugin for the protocol compiler, version 1.5.1
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# update your PATH so that the protoc compiler can find the plugin
export PATH="$PATH:$(go env GOPATH)/bin"

# install protoc validate
buf dep update
```

### Install live reloader

go install github.com/air-verse/air@1.61.1

### Install linter

```

# binary will be in $(go env GOPATH)/bin/golangci-lint

curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.62.2

golangci-lint --version
```


ENV SETUP 

```

export APP_PORT=4000
export TEST_APP_PORT=5000

# ----- KAFKA CONFIGURATION -----
export KAFKA_HOST=localhost
export KAFKA_PORT=4003
export KAFKA_MAIN_BROKER=${KAFKA_HOST}:${KAFKA_PORT}
export KAFKA_EVENT_TOPIC="app-events"

# ----- DATABASE CONFIGURATION -----
export DATABASE_NAME=m
export DATABASE_URL=

export TEST_DATABASE_NAME=
export TEST_DATABASE_URL=

# TODO: figure out why this is needed
export DATABASE_ROLE=omarcruz

export DB_SOURCE_DIR=src/psql_db
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING=${DATABASE_URL}
export GOOSE_MIGRATION_DIR=${DB_SOURCE_DIR}/migrations

export PROJECT_ROOT_PATH=$(pwd)

# ----- JWT AUTHORIZATION CONFIG -----
export MIST_API_JWT_SECRET_KEY=""
export MIST_API_JWT_AUDIENCE=""
export MIST_API_JWT_ISSUER=""

```