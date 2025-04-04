# Honoroit [![Matrix](https://img.shields.io/matrix/honoroit:etke.cc?logo=matrix&style=for-the-badge)](https://matrix.to/#/#honoroit:etke.cc)

> [more about that name](https://finalfantasy.fandom.com/wiki/Honoroit_Banlardois)


<!-- vim-markdown-toc GFM -->

* [Features](#features)
* [How it looks like](#how-it-looks-like)
    * [Step 1: a matrix user (customer) sends a message to Honoroit bot in direct 1:1 chat](#step-1-a-matrix-user-customer-sends-a-message-to-honoroit-bot-in-direct-11-chat)
    * [Step 2: a new thread created in the backoffice room](#step-2-a-new-thread-created-in-the-backoffice-room)
    * [Step 3: operator(-s) chat with customer in that thread](#step-3-operator-s-chat-with-customer-in-that-thread)
    * [Step 4: customer sees that like a direct 1:1 chat with honoroit user](#step-4-customer-sees-that-like-a-direct-11-chat-with-honoroit-user)
    * [Step 5: operator closes the request](#step-5-operator-closes-the-request)
    * [Step 6: customer receives special message and bot leaves the room](#step-6-customer-receives-special-message-and-bot-leaves-the-room)
* [TODO](#todo)
* [Commands](#commands)
* [Configuration](#configuration)
    * [mandatory](#mandatory)
    * [optional](#optional)
        * [honoroit internals](#honoroit-internals)
        * [redmine](#redmine)
* [Where to get](#where-to-get)
    * [etke.cc](#etkecc)
    * [Matrix Docker Ansible Deploy](#matrix-docker-ansible-deploy)
    * [Docker](#docker)
    * [Binary](#binary)
    * [Build from source](#build-from-source)

<!-- vim-markdown-toc -->

A helpdesk bot, used as part of [etke.cc](https://etke.cc) service.

The main idea of that bot is to give you the same abilities as with website chats (like Intercom, jivosite, etc) inside the matrix.

## Features

* chat-based configuration
* optional silent mode (bot won't send any automatic messages to the customer)
* prometheus metrics on `/metrics` endpoint
* [Redmine integration](./docs/redmine.md)
* [MSC4144 integration](./docs/msc4144.md)
* End-to-End encryption
* Get a message from any matrix user proxied to a specific room. Any message that user will send in his 1:1 room with Honoroit will be proxied as thread messages
* Chat with that user through the honoroit bot in a thread inside your special room. Any member of that special room can participate in discussion
* When request fulfilled - send a `!ho done` in that thread - thread topic will be renamed and "proxied user" will know that request was closed (bot will leave user's room with special notice)

## How it looks like

### Step 1: a matrix user (customer) sends a message to Honoroit bot in direct 1:1 chat

![step 1](contrib/screenshots/1-customer-sends-a-message.png)

### Step 2: a new thread created in the backoffice room

![step 2](contrib/screenshots/2-new-thread-created-in-the-backoffice-room.png)

### Step 3: operator(-s) chat with customer in that thread

![step 3](contrib/screenshots/3-operators-chat-with-customer-in-that-thread.png)

### Step 4: customer sees that like a direct 1:1 chat with honoroit user

![step 4](contrib/screenshots/4-customer-sees-that-like-a-direct-1-1-chat-with-honoroit-user.png)

### Step 5: operator closes the request

![step 5](contrib/screenshots/5-operator-closes-the-request.png)

### Step 6: customer receives special message and bot leaves the room

![step 6](contrib/screenshots/6-customer-receives-special-message-and-bot-leaves-the-room.png)

## TODO

* Unit tests

## Commands

Available commands in the threads. Note that all commands should be called with prefix, so `!ho done` will work, but simple `done` will not.

* `done` - close the current request and mark is as done. Customer will receive special message and honoroit bot will leave 1:1 chat with customer. Any new message to the thread will not work and return error.
* `rename TXT` - rename the thread topic title, when you want to change the standard message to something different
* `note NOTE` - a message prefixed with `!ho note` will **not** be sent anywhere, it's a safe place to keep notes for other operations in a thread with a customer, example: `!ho note @room need help with this one`
* `invite` - invite yourself into the customer 1:1 room
* `start MXID` - start a conversation with a MXID from the honoroit (like a new thread, but initialized by operator), eg: `!ho start @user:example.com`
* `count MXID` - count a request from MXID and their homeserver, but don't actually create a room or invite them
* `config` - show all config options
* `config KEY` - show specific config option and its current value
* `config KEY VALUE` - update value of the specific config option


## Configuration

env vars

### mandatory

* **HONOROIT_HOMESERVER** - homeserver url, eg: `https://matrix.example.com`
* **HONOROIT_LOGIN** - user login, localpart when using password (e.g., `honoroit`), OR full MXID when using shared secret (e.g., `@honoroit:example.com`)
* **HONOROIT_PASSWORD** - user password, alternatively you may use shared secret
* **HONOROIT_SHAREDSECRET** - alternative to password, shared secret ([details](https://github.com/devture/matrix-synapse-shared-secret-auth))
* **HONOROIT_ROOMID** - room ID where threads will be created, eg: `!test:example.com`

### optional

#### honoroit internals

* **HONOROIT_PREFIX** - command prefix
* **HONOROIT_PORT** - http port
* **HONOROIT_DATA_SECRET** - secure key (password) to encrypt account data, must be 16, 24, or 32 bytes long
* **HONOROIT_MONITORING_SENTRY_DSN** - sentry DSN
* **HONOROIT_MONITORING_SENTRY_RATE** - sentry sample rate, from 0 to 100 (default: 20)
* **HONOROIT_MONITORING_HEALTHCHECKS_URL** - healthchecks.io url, default: `https://hc-ping.com`
* **HONOROIT_MONITORING_HEALTHCHECKS_UUID** - healthchecks.io UUID
* **HONOROIT_MONITORING_HEALTHCHECKS_DURATION** - heathchecks.io duration between pings in seconds (default: 5)
* **HONOROIT_LOGLEVEL** - log level
* **HONOROIT_CACHESIZE** - max allowed mappings in cache
* **HONOROIT_NOENCRYPTIONWARNING** - disable e2e encryption warning message
* **HONOROIT_DB_DSN** - database connection string
* **HONOROIT_DB_DIALECT** - database dialect (postgres, sqlite3)
* **HONOROIT_METRICS_LOGIN** - /metrics login
* **HONOROIT_METRICS_PASSWORD** - /metrics password
* **HONOROIT_METRICS_IPS** - /metrics allowed ips

#### redmine

Optional 2-way sync with redmine

* **HONOROIT_REDMINE_HOST** - redmine host, e.g. `https://redmine.example.com`
* **HONOROIT_REDMINE_APIKEY** - redmine API key
* **HONOROIT_REDMINE_PROJECT** - redmine project identifier, e.g. `internal-project`
* **HONOROIT_REDMINE_TRACKERID** - redmine tracker ID, e.g. `1`
* **HONOROIT_REDMINE_NEWSTATUSID** - redmine "new" status ID, e.g. `1`
* **HONOROIT_REDMINE_INPROGRESSSTATUSID** - redmine "in progress" status ID, e.g. `2`
* **HONOROIT_REDMINE_DONESTATUSID** - redmine "done" status ID, e.g. `3`

You can find default values in [config/defaults.go](config/defaults.go)

## Where to get

### [etke.cc](https://etke.cc)

You can get a hosted version of the bot with a support and maintenance plan by the developers on [etke.cc](https://etke.cc)

### [Matrix Docker Ansible Deploy](https://github.com/spantaleev/matrix-docker-ansible-deploy)

You can get the bot using MDAD playbook, just [follow the docs](https://github.com/spantaleev/matrix-docker-ansible-deploy/blob/master/docs/configuring-playbook-bot-honoroit.md)

### Docker

You can get the bot using docker image from [the registry](https://github.com/etkecc/honoroit/pkgs/container/honoroit)

```bash
# 1. Get UID and GID of the system user you want to use to run the bot's container:
# id
# 2. Prepare the configuration in the env file (see Configuration section), alternatively you can use `docker --env` flags
# 2. Run the bot using UID and GID from step 1:
docker run --user YOUR_UID:YOUR_GID --env-file /YOUR_ENV_FILE ghcr.io/etkecc/honoroit:latest
```

### Binary

You can get binary from [the releases page](https://github.com/etkecc/honoroit/releases):

```bash
# 1. Prepare the configuration in the .env file (see Configuration section), alternatively you can use env vars
# 2. Run the bot (it will load .env file in the current directory automatically)
./honoroit
```

### Build from source

```bash
# 1. Clone the repo
# 2. Prepare the configuration in the .env file (see Configuration section), alternatively you can use env vars
# 3. (if `just` is installed)
just build
# or
go build -ldflags '-extldflags "-static"' -tags timetzdata,goolm -v ./cmd/honoroit
```
