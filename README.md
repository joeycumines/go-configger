# go-configger

A tool for merging and modifying config files, intended for development use.

Mostly complete, tests are sub par, but functionality is good.

- merging config files works great, supports both arrays and maps
- env, json and yaml are all supported (including merging together)
- output format may be any of the three above, though env only supports flat
  maps
- blacklisting (exclusion) of nodes using dot notation works well, 
- whitelisting doesn't do much, since some effort is required to make it work
  the way I originally intended

## Install

```bash
go get -u github.com/joeycumines/go-configger/cmd/goconfigger
```

## Usage

```bash
# see the command's help for more info
goconfigger help
```

## LICENSE

See the `LICENCE` file.
