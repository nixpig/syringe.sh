[![Workflow Status](https://github.com/nixpig/syringe.sh/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/nixpig/syringe.sh/actions/workflows/build.yml?query=branch%3Amain)
[![Coverage Status](https://coveralls.io/repos/github/nixpig/syringe.sh/badge.svg?branch=main)](https://coveralls.io/github/nixpig/syringe.sh?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/nixpig/syringe.sh)](https://goreportcard.com/report/github.com/nixpig/syringe.sh)

# 🔐 `syringe.sh`

Encrypted, passwordless, embeddable key-value store.


##



- `syringe set foo bar` - Set an item named 'foo' with value 'bar'
- `syringe get foo` - Get the item named 'foo'
- `syringe delete foo` - Delete the item named 'foo'
- `syringe list` - List all items

- `--identity` - Specify path the to SSH key to use, defaults to: 
- `--store` - Specify path to the store to use, defaults to: 
- `--config` - Specify path to a config file, defaults to: 

Precedence of config and flags.
- Config _not_ required, but if it's present then use it.
- Environment variables not required, but if they're present then use them.

## ONLY DEAL WITH FLAGS AND DEFAULT VALUES FOR THE MOMENT!!
- Flags overrides all.

## Default values
- identity = whatever is currently loaded in the ssh keyring
- store = $HOME/.syringe/{id}.db

Config: 
```env
identity=
store=
```
