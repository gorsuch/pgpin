#!/bin/bash

set -exo pipefail

# Env
touch /home/vagrant/.profile
chown vagrant:vagrant /home/vagrant/.profile

# Synced folder ownership
chown -R vagrant:vagrant /home/vagrant/src

# Go
apt-get update
apt-get install -y --no-install-recommends build-essential curl git-core mercurial bzr

mkdir -p /opt/go
curl -s https://storage.googleapis.com/golang/go1.2.2.linux-amd64.tar.gz | tar xzf - -C /opt/go --strip-components=1

cat >> /home/vagrant/.profile <<EOF
export GOROOT=/opt/go
export GOPATH=\$HOME
export PATH=\$HOME/bin:/opt/go/bin:\$PATH
EOF

# Postgres
cat > /etc/apt/sources.list.d/pgdg.list <<EOF
deb http://apt.postgresql.org/pub/repos/apt/ trusty-pgdg main
EOF

curl -s https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
apt-get update

apt-get install -y --no-install-recommends postgresql-9.3

sudo -u postgres psql -U postgres -d postgres -c "alter user postgres with password 'secret';"
sudo -u postgres createdb pgpin-development
sudo -u postgres createdb pgpin-test

cat >> /home/vagrant/.profile <<EOF
export DEVELOPMENT_DATABASE_URL=postgres://postgres:secret@127.0.0.1:5432/pgpin-development
export TEST_DATABASE_URL=postgres://postgres:secret@127.0.0.1:5432/pgpin-test
export DATABASE_URL=\$DEVELOPMENT_DATABASE_URL
EOF

# Goreman and Godep
sudo -u vagrant -i go get github.com/mattn/goreman
sudo -u vagrant -i go get github.com/tools/godep

# Config
cat >> /home/vagrant/.profile <<EOF
export API_AUTH="client:"$(openssl rand -hex 12)
export FERNET_KEY=$(openssl rand -base64 32)
export PGPIN_API_URL=http://\$API_AUTH@127.0.0.1:5000
cd ~/src/github.com/mmcgrana/pgpin
EOF
