[![Workflow Status](https://github.com/nixpig/syringe.sh/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/nixpig/syringe.sh/actions/workflows/build.yml?query=branch%3Amain)
[![Coverage Status](https://coveralls.io/repos/github/nixpig/syringe.sh/badge.svg?branch=main)](https://coveralls.io/github/nixpig/syringe.sh?branch=main)

# 🔐 `syringe.sh`

Distributed database-per-user encrypted secrets management over SSH protocol.

## TODO

- [x] Merge client and server codebases into one
- [x] Share command configuration for both cli client and server
- [ ] Secret injection and run passed command
  - [ ] Pass a `io.Writer` into `run` so that we can read secrets from it to inject instead of directly printing to terminal out
- [ ] Encryption of secrets
- [ ] E2E tests with an SSH client, including a couple like trying to create secrets for a non-existent project or environmnet
  - Work out how to start/stop server asynchronously and run tests. Could be containerised using testcontainers?
- [x] Explicit (not implicit) user registration
- [ ] Improve error handling, errors and messaging
- [x] Exit codes on error
- [ ] Remove use of third-party package for SSH client (in CLI client)
- [ ] Proper good refactor and tidy-up (primarily of database stuff)
- [ ] Pull the Turso stuff out into separate SDK package??

## Proposed usage

### `syringe user`

Manage users.

### `syringe project`

Manage projects.

- [x] `syringe project add <name>`
- [x] `syringe project remove <name>`
- [x] `syringe project list`
- [x] `syringe project rename <old name> <new name>`

### `syringe environment`

Manage environments.

- [x] `syringe environment add [-p <project>] <name>`
- [x] `syringe environment remove [-p <project>] <name>`
- [x] `syringe environment list [-p <project>]`
- [x] `syringe environment rename [-p <project>] <old name> <new name>`

### `syringe secret`

Manage secrets.

- [x] `syringe secret set [-p <project> -e <environment>] <key> <value>`
- [x] `syringe secret get [-p <project> -e <environment>] <key>`
- [x] `syringe secret list [-p <project> -e <environment>]`
- [x] `syringe secret remove [-p <project> -e <environment>] <key>`

### `syringe inject`

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
