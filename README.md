[![Workflow Status](https://github.com/syringe-sh/server/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/syringe-sh/server/actions/workflows/build.yml?query=branch%3Amain)
[![Coverage Status](https://coveralls.io/repos/github/syringe-sh/server/badge.svg?branch=main)](https://coveralls.io/github/syringe-sh/server?branch=main)

# `syringe.sh/server`

> **⚠️ THIS PROJECT IS A WORK IN PROGRESS**
>
> **⚠️ SOME OR ALL OF THE FUNCTIONALITY HERE MAY NOT ACTUALLY WORK CORRECTLY**

## Proposed usage

`syringe` executed without a subcommand should connect to a TUI.

### `inject`

- `syringe inject [-p <project> -e <environment> -s <key>]`

#### Examples

- `syringe inject`
- `syringe inject -p galactic_takeover -e death_star -s launch_code .`

### `link`

- `syringe link [-p <project> -p <environment>] <directory>`

#### Examples

- `syringe link -p galactic_takeover -e death_star .`

### `user`

Manage your user account(s).

- `syringe user register [-u <username>]`
- `syringe user delete [-u <username>]`

### `project`

Manage your projects.

- `syringe project add <name>`
- `syringe project remove <name>`
- `syringe project list`
- `syringe project rename <old name> <new name>`

### `environment`

Manage your environments.

- `syringe environment add [-p <project>] <name>`
- `syringe environment remove [-p <project>] <name>`
- `syringe environment list [-p <project>]`
- `syringe environment rename [-p <project>] <old name> <new name>`

### `secret`

Manage your secrets.

- `syringe secret set [-p <project> -e <environment>] <key> <value>`
- `syringe secret get [-p <project> -e <environment>] <key>`
- `syringe secret list [-p <project> -e <environment>]`
- `syringe secret remove [-p <project> -e <environment>] <key>`

### `key`

Manage SSH public keys associated with your account.

- `syringe key add`
- `syringe key remove [-k <public key>]`
