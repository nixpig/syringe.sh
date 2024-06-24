[![Workflow Status](https://github.com/syringe-sh/server/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/syringe-sh/server/actions/workflows/build.yml?query=branch%3Amain)
[![Coverage Status](https://coveralls.io/repos/github/syringe-sh/server/badge.svg?branch=main)](https://coveralls.io/github/syringe-sh/server?branch=main)

# 🔐 `syringe.sh/server`

Distributed database-per-user encrypted secrets management over SSH protocol.

## TODO

- [ ] E2E tests with an SSH client, including a couple like trying to create secrets for a non-existent project or environmnet
  - Work out how to start/stop server asynchronously and run tests. Could be containerised using testcontainers?
- [?] Explicit (not implicit) user registration
- [ ] Improve errors and messaging
- [ ] Secret injection
- [ ] Remove use of third-party package for SSH client (in CLI client)
- [?] Merge client and server codebases into one
- [ ] Proper good refactor and tidy-up
- [ ] Pull the Turso stuff out into separate SDK package??

## Proposed usage

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
- [x] `syringe environment list [-p <project>]`
- [x] `syringe environment rename [-p <project>] <old name> <new name>`

Manage your secrets.

- [x] `syringe secret set [-p <project> -e <environment>] <key> <value>`
- [x] `syringe secret get [-p <project> -e <environment>] <key>`
- [x] `syringe secret list [-p <project> -e <environment>]`
- [x] `syringe secret remove [-p <project> -e <environment>] <key>`

### `user`

Manage your user account(s).

- [ ] `syringe user register [-u <username>]`
- [ ] `syringe user list-keys [-u <username>]`
- [ ] `syringe user add-key [-k <ssh_public_key>] [-u <username>]`
- [ ] `syringe user remove-key [-k <ssh_public_key>] [-u <username>]`
- [ ] `syringe user delete [-u <username>]`

### `inject`

- [ ] `syringe inject [-p <project> -e <environment> -s <secret_key>] COMMAND`

#### Examples

Unlinked:

- [ ] `syringe inject -p galactic_takeover -e death_star -s ENV_LAUNCH_CODE my_cool_app`

Linked:

- [ ] `syringe inject my_cool_app`

### `link`

- [ ] `syringe link [-p <project> -p <environment>] <directory>`

#### Examples

- [ ] `syringe link -p galactic_takeover -e death_star .`
- [ ] `syringe link -p galactic_takeover -e death_star /home/nixpig/projects/joubini`
