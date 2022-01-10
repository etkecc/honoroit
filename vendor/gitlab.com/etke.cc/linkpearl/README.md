# linkpearl [![Buy me a Coffee](https://shields.io/badge/donate-buy%20me%20a%20coffee-green?logo=buy-me-a-coffee&style=for-the-badge)](https://buymeacoffee.com/etkecc)[![Go Report Card](https://goreportcard.com/badge/gitlab.com/etke.cc/linkpearl)](https://goreportcard.com/report/gitlab.com/etke.cc/linkpearl) [![Go Reference](https://pkg.go.dev/badge/gitlab.com/etke.cc/linkpearl.svg)](https://pkg.go.dev/gitlab.com/etke.cc/linkpearl)

> [more about that name](https://ffxiv.gamerescape.com/wiki/Linkpearl)

A wrapper around [mautrix-go](https://github.com/mautrix/go) with infrastructure/glue code included

## How to get

```
go get gitlab.com/etke.cc/linkpearl
```

```
lp, err := linkpearl.New(&linkpearl.Config{
	// your options here
})
if err != nil {
	panic(err)
}

go lp.Start()
```

## TODO

* Integrate with [Honoroit](https://gitlab.com/etke.cc/honoroit)
* Unit tests

## Features

* Zero configuration End-to-End encryption
* Zero configuration persistent storage
* Zero configuration session restores
* All wrapped components exported
