# Honoroit [![Matrix](https://img.shields.io/matrix/honoroit:etke.cc?logo=matrix&style=for-the-badge)](https://matrix.to/#/#honoroit:etke.cc) [![Buy me a Coffee](https://shields.io/badge/donate-buy%20me%20a%20coffee-green?logo=buy-me-a-coffee&style=for-the-badge)](https://buymeacoffee.com/etkecc) [![coverage report](https://gitlab.com/etke.cc/honoroit/badges/main/coverage.svg)](https://gitlab.com/etke.cc/honoroit/-/commits/main) [![Go Report Card](https://goreportcard.com/badge/gitlab.com/etke.cc/honoroit)](https://goreportcard.com/report/gitlab.com/etke.cc/honoroit) [![Go Reference](https://pkg.go.dev/badge/gitlab.com/etke.cc/honoroit.svg)](https://pkg.go.dev/gitlab.com/etke.cc/honoroit)

> [more about that name](https://finalfantasy.fandom.com/wiki/Honoroit_Banlardois)

A helpdesk bot, used as part of [etke.cc](https://etke.cc) service.

## Features

* TBD

## TODO

* process files
* process images
* Email<->Matrix helpdesk
* End-to-End Encryption
* autoleave empty rooms and notify about that (requires persistent store)

## How it works

1. You configure Honoroit and start it
2. Any matrix user starts a chat with honoroit bot and and sends it a message
3. Honoroit with copy that message (any any following messages) to the predefined room, where will start a thread with name `Request by <MXID>`, posting all original messages in that thread.
4. You can answer in the same thread and honoroit will copy your messages from that thread to the user's 1:1 with honoroit
5. When request fulfilled, send `@honoroit:your.server done` - bot will send a message to the user's 1:1 chat that ticket has been closed and will leave that chat. You can safely remove the thread after that.


## Configuration

env vars

### mandatory

* **HONOROIT_HOMESERVER** - homeserver url, eg: `https://matrix.example.com`
* **HONOROIT_LOGIN** - user login/localpart, eg: `honoroit`
* **HONOROIT_PASSWORD** - user password
* **HONOROIT_ROOMID** - room ID where threads will be created, eg: `!test:example.com`

### optional

* **HONOROIT_LOGLEVEL** - log level
* **HONOROIT_TEXT_GREETINGS** - a message sent to customer on first contact
* **HONOROIT_TEXT_ERROR** - a message sent to customer if something goes wrong
* **HONOROIT_TEXT_EMPTYROOM** - a message sent to backoffice/threads room when customer left his room
* **HONOROIT_TEXT_DONE** - a message sent to customer when request marked as done in the threads room

You can find default values in [config/defaults.go](config/defaults.go)

## Where to get

[Binary releases](https://gitlab.com/etke.cc/honoroit/-/releases), [docker registry](https://gitlab.com/etke.cc/honoroit/container_registry), [etke.cc](https://etke.cc)
