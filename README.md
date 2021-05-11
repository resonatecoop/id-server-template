# Resonate ID

Resonate's ID server is a Go OAuth2 Server based on [RichardKnop/go-oauth2-server](https://github.com/RichardKnop/go-oauth2-server).

* [Setup](#setup)
* [Run](#run)
* [Deploy](#deploy)

See also

* [OAuth 2.0](docs/oauth2.md)
* [Environment](docs/environment.md)
* [Docker](docs/docker.md)
* [Plugins](docs/plugins.md)
* [Tests](docs/tests.md)

## Setup

There are three setup tasks when working with the ID server

1. Setup the config store
2. Setup the data stores
3. Compile the server

### Config Store

The ID server uses [etcd](https://github.com/etcd-io/etcd) as a config store.

Install etcd (if needed, and platform specific) and run it

```
brew install etcd
etcd
```

Load the sample config and verify it

```
etcdctl put /config/go_oauth2_server.json "$(cat ./config.sample.json)"
etcdctl get /config/go_oauth2_server.json
```

### Data Store

There are currently 3 data stores

1. Wordpress (legacy)
2. OAuth2 Default Postgres (legacy)
3. User API

We'll be replacing both the Wordpress and Go OAuth2 Default Postgres with a direct database connection to the User API, which will own all of the ID data.

#### Wordpress (legacy)

To work with the Wordpress store you need your own local Resonate-configured Wordpress database.

First, create a mysql database user, using the values in ``./wp-config.php``.

```
mysql
CREATE USER 'go_oauth2_server'@'localhost' IDENTIFIED BY '';
```

Next, using the [WP-CLI](https://wp-cli.org/), install Wordpress in a tmp directory

```
mkdir $TMPDIR/wordpress
wp core download --path=$TMPDIR/wordpress
```

Copy across the wp-config.php and create a new database

```
cp ./wp-config.php $TMPDIR/wordpress/wp-config.php
wp db create --path=$TMPDIR/wordpress
```

Finally, install the Wordpress with an admin user

```
wp core install --url=resonate.is --title=Resonate --admin_user=angus --admin_password=mypass --admin_email=angus@example.com --path=$TMPDIR/wordpress
```

You now have a local Wordpress database to use when testing the legacy user data source. You can create new test users as needed like so

```
wp user create bruce bruce@wayne-enterprises.com --role=subscriber --user_pass=batman
```

#### OAuth2 Default Postgres (legacy)

The default data store for the OAuth2 Server is a Postgres database. To set it up, first install Postgres (platform specific)

```
brew install postgres
```

Then create a local database for the server

```
createuser --createdb go_oauth2_server
createdb -U go_oauth2_server go_oauth2_server
```

Finally, load test data from the ``oauth/fixtures`` to work with locally

```
go-oauth2-server loaddata oauth/fixtures/scopes.yml oauth/fixtures/roles.yml oauth/fixtures/test_clients.yml
```

In your local go_oauth2_server postgres database you should now see scopes, roles and clients in your database. Rather than loading the users from the fixtures, we'll load the users from our other legacy database, Wordpress. 

First, set the constants in ``migrate-wp-users/.env-example`` in your environment. Then run the perl script in the
same folder to import the wordpress user(s) created in the previous section into the ``oauth_users`` table in the 
``go_oauth2_server`` database.

```
perl migrate_wp_user_to_oauth.pl
```

#### User API

(How to setup the User API locally)

### Compile

To compile the server run

```
go install .
```

## Run

First, run migrations

```
go-oauth2-server migrate
```

Then run the server

```
go-oauth2-server runserver
```

## Deploy

(How to deploy to staging and production using [docker](docs/docker.md))

## Develop

Add a git hook for proper formatting
```
./add_gofmt_hook.sh
```
