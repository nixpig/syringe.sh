[![Workflow Status](https://github.com/syringe-sh/server/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/syringe-sh/server/actions/workflows/build.yml?query=branch%3Amain)
[![Coverage Status](https://coveralls.io/repos/github/syringe-sh/server/badge.svg?branch=main)](https://coveralls.io/github/syringe-sh/server?branch=main)

# `syringe.sh/server`

> **⚠️ THIS PROJECT IS A WORK IN PROGRESS**
>
> **⚠️ SOME OR ALL OF THE FUNCTIONALITY HERE MAY NOT ACTUALLY WORK AS DOCUMENTED, OR EVEN AT ALL**

## TODO

- [ ] Improve errors/messaging

## Proposed usage

`syringe` executed without a subcommand should connect to a TUI.

### `project`

Manage your projects.

- [x] `syringe project add <name>`
- [x] `syringe project remove <name>`
- [x] `syringe project list`
- [x] `syringe project rename <old name> <new name>`

### `environment`

Manage your environments.

- [x] `syringe environment add [-p <project>] <name>`
- [x] `syringe environment remove [-p <project>] <name>`
- [-] `syringe environment list [-p <project>]`
- [x] `syringe environment rename [-p <project>] <old name> <new name>`

Manage your secrets.

- [x] `syringe secret set [-p <project> -e <environment>] <key> <value>`
- [x] `syringe secret get [-p <project> -e <environment>] <key>`
- [-] `syringe secret list [-p <project> -e <environment>]`
- [ ] `syringe secret remove [-p <project> -e <environment>] <key>`

### `user`

Manage your user account(s).

- [ ] `syringe user register [-u <username>]`
- [ ] `syringe user delete [-u <username>]`

### `inject`

- [ ] `syringe inject [-p <project> -e <environment> -s <key>]`

#### Examples

- [ ] `syringe inject`
- [ ] `syringe inject -p galactic_takeover -e death_star -s launch_code .`

### `link`

- [ ] `syringe link [-p <project> -p <environment>] <directory>`

#### Examples

- [ ] `syringe link -p galactic_takeover -e death_star .`

### `key`

Manage SSH public keys associated with your account.

- [ ] `syringe key add`
- [ ] `syringe key remove [-k <public key>]`
