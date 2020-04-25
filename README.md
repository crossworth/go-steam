# Steam for Go

[![BSD](https://img.shields.io/github/license/13k/go-steam)](https://github.com/13k/go-steam/blob/master/LICENSE.txt)
[![GoDoc](https://godoc.org/github.com/13k/go-steam?status.svg)](https://pkg.go.dev/github.com/13k/go-steam)
[![GoReportCard](https://goreportcard.com/badge/github.com/13k/go-steam)](https://goreportcard.com/report/github.com/13k/go-steam)

> Automate actions on the Steam network with Go.

## Introduction

This library implements Steam's protocol to allow automation of different actions on Steam without
running an actual Steam client. It is based on [SteamKit2](https://github.com/SteamRE/SteamKit), a
dotnet library.

In addition, it contains APIs to Steam Community features, like trade offers and inventories.

Some of the currently implemented features:

- Trading and trade offers, including inventories and notifications
- Friend and group management
- Chatting with friends
- Persona states (online, offline, looking to trade, etc.)
- SteamGuard with two-factor authentication
- Team Fortress 2: Crafting, moving, naming and deleting items

If this is useful to you, there's also the [geyser](https://github.com/13k/geyser) package that
implements the official Steam Web API.

This package was originally authored by [Philipp15b](https://github.com/Philipp15b/go-steam),
forked/adopted by [FACEIT](https://github.com/faceit/go-steam) and forked by @13k.

## Installation

    go get github.com/13k/go-steam

## Usage

You can view the documentation [online](https://pkg.go.dev/github.com/13k/go-steam).

You should also take a look at the following sub-packages:

- [`gsbot`](https://pkg.go.dev/github.com/13k/go-steam/gsbot): utilites that make writing bots easier
- [`gsbot` command](https://pkg.go.dev/github.com/13k/go-steam/cmd/gsbot): example of using `gsbot`
- [`economy/inventory`](http://pkg.go.dev/github.com/13k/go-steam/economy/inventory): inventories
- [`economy/trade`](https://pkg.go.dev/github.com/13k/go-steam/economy/trade): trading
- [`economy/trade/tradeoffer`](https://pkg.go.dev/github.com/13k/go-steam/economy/trade/tradeoffer): trade offers

## Working with go-steam

Whether you want to develop your own Steam bot or directly work on go-steam itself, there are are
few things to know.

- If something is not working, check first if the same operation works (under the same conditions!)
  in the Steam client on that account. Maybe there's something go-steam doesn't handle correctly or
  you're missing a warning that's not obviously shown in go-steam. This is particularly important
  when working with trading since there are
  [restrictions](https://support.steampowered.com/kb_article.php?ref=1047-edfm-2932), for example
  newly authorized devices will not be able to trade for seven days.
- Since Steam does not maintain a public API for most of the things go-steam implements, you can
  expect that sometimes things break randomly. Especially the `economy/trade` and
  `economy/trade/tradeoffer` packages have been affected in the past.
- Always gather as much information as possible. When you file an issue, be as precise and complete
  as you can. This makes debugging way easier.
- If you haven't noticed yet, expect to find lots of things out yourself. Debugging can be
  complicated and Steam's internals are too. [SteamKit](https://github.com/SteamRE/SteamKit) is the
  reference implementation, so make sure to check it out when in doubt.
- Sometimes things break and other [SteamKit ports](https://github.com/SteamRE/SteamKit/wiki/Ports)
  are fixed already. Maybe take a look what people are saying over there? There's also the
  [SteamKit IRC channel](https://github.com/SteamRE/SteamKit/wiki#contact).

## License

Steam for Go is licensed under the New BSD License. More information can be found in LICENSE.txt.
