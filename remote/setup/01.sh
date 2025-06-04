#!/bin/bash
set -eu

# ============================================================================== #
# VARIABLES
# ============================================================================== #

# set the timezone for the server
# timedatectl list-timezones
TIMEZONE=America/New_York

# set the name of the new user to create
USERNAME=greenlight

# prompt to enter a password for the db greenlight user (no hardcoding password)
read -p -r "Enter password for greenlight DB user: " DB_PASSWORD

# force all output to be presented in en_US for the duration of the script
#  this avoids any "settings locale failed" errors while the script is running
export LC_ALL=en_US.UTF8

# ============================================================================== #
# SCRIPT LOGIC
# ============================================================================== #

# enable the universe repo
add-apt-repository --yes universe

# update all
# confnew means cocnfig. files will be replaced if newer ones are avail.
apt-update
apt --yes -o Dpkg::Options::="--force-confnew" upgrade

# set system timezone and install all locales
timedatectl set-timezone ${TIMEZONE}
apt --yes install locales-all

# add new user and give sudo privvies
useradd --create-home --shell /bin/bash --groups sudo ${USERNAME}

# force a pwd to be set for the new user the first time they login
passwd --delete "${USERNAME}"
chage --lastday 0 "${USERNAME}"

# change ssh keys from root to new user
rsync --archive --chown=${USERNAME}:${USERNAME} /root/.ssh /home/${USERNAME}

# configure firewall to allow ssh, http and https traffic
ufw allow 22
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# Install fail2ban.
apt --yes install fail2ban

# Install the migrate CLI tool.
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
mv migrate.linux-amd64 /usr/local/bin/migrate

# Install PostgreSQL.
apt --yes install postgresql

# Set up the greenlight DB and create a user account with the password entered earlier.
sudo -i -u postgres psql -c "CREATE DATABASE greenlight"
sudo -i -u postgres psql -d greenlight -c "CREATE EXTENSION IF NOT EXISTS citext"
sudo -i -u postgres psql -d greenlight -c "CREATE ROLE greenlight WITH LOGIN PASSWORD '${DB_PASSWORD}'"

# Add a DSN for connecting to the greenlight database to the system-wide environment
# variables in the /etc/environment file.
echo "GREENLIGHT_DB_DSN='postgres://greenlight:${DB_PASSWORD}@localhost/greenlight'" >> /etc/environment

# Install Caddy (see https://caddyserver.com/docs/install#debian-ubuntu-raspbian).
apt --yes install -y debian-keyring debian-archive-keyring apt-transport-https
curl -L https://dl.cloudsmith.io/public/caddy/stable/gpg.key | sudo apt-key add -
curl -L https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt | sudo tee -a /etc/apt/sources.list.d/caddy-stable.list
apt update
apt --yes install caddy
echo "Script complete! Rebooting..."
reboot
