# Mist Backend

## Installation

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

# After installation create new database called mist
sudo -u postgres psql

CREATE DATABASE mist;
```
