[![Workflow Status](https://github.com/nixpig/syringe.sh/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/nixpig/syringe.sh/actions/workflows/build.yml?query=branch%3Amain)
[![Coverage Status](https://coveralls.io/repos/github/nixpig/syringe.sh/badge.svg?branch=main)](https://coveralls.io/github/nixpig/syringe.sh?branch=main)

# 🔐 `syringe.sh`

Distributed database-per-user encrypted secrets management over SSH protocol.

## TODO

- [x] Confirm authentication before calling cmd, e.g. with unregistered user calling project command results in NPE
- [x] Merge client and server codebases into one
- [x] Share command configuration for both cli client and server
- [x] Secret injection and run passed command
  - [x] Pass a `io.Writer` into `run` so that we can read secrets from it to inject instead of directly printing to terminal out
- [x] Update and enable the disabled unit tests
- [x] Utilise settings file to specify/save which identity to use (viper)
- [ ] Load ssh config from file
  - [ ] Single identity for host
  - [ ] Multiple identities for host - prompt user to choose which to use (have a default - add a flag to just use default (or specific) without prompt)
- [ ] Use SSH agent by default - check if key is in agent, if key isn't already in agent when loaded, prompt user to add to agent (add a flag for this too!)
- [ ] Encryption and signing of secrets
  - [ ] RSA
  - [ ] OpenSSH
  - [ ] ECDSA
  - [ ] ED25519
- [ ] Add 'syringe config' command to update config file
- [ ] Add unit tests for other areas
- [ ] E2E tests with the CLI (or SSH?) client, including a couple like trying to create secrets for a non-existent project or environmnet
  - Work out how to start/stop server asynchronously and run tests. Could be containerised using testcontainers?
  - Just use testcontainers??
- [ ] Add functionality to 'link' local directories/projects to specific project/environment - save in config file
- [x] Explicit (not implicit) user registration
- [ ] Improve error handling, errors and messaging
- [x] Exit codes on error
- [ ] Accept spaces in secret values
- [ ] Remove use of third-party package for SSH client (in CLI client)
- [ ] Proper good refactor and tidy-up (primarily of database stuff)
- [ ] Pull the Turso stuff out into separate SDK package??
- [ ] Create a wrapper package around the various SSH related stuff like config and known hosts

## Supported SSH key types

- RSA

## Usage

### Specifying an identity

An identity must be specified to connect over SSH and to encrypt/decrypt secrets.

The identity to use is selected with the following order of precedence.

1. The `--identity` flag.
1. The `identity` property in [settings file](#settings-file).
1. The running SSH agent, if available.

Note: when using the SSH agent directly, the syringe.sh host must also be configured in SSH config.

In any case, if you have an SSH agent running and the specified identity is not already loaded into the SSH agent, you will be prompted to do so.

### Settings file

syringe.sh uses a settings file located in your user config directory at `syringe/settings`, for example: `/home/nixpig/.config/syringe/settings`

The settings file uses a `key=value` format, with each key/value pair on a new line.

#### Settings

| Key                   | Value          | Description                                                                                                                                                       |
| --------------------- | -------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `identity`            | `string`       | Path to the SSH identity file to use. Equivalent to the `-i` flag to `ssh` or the `IdentityFile` parameter in SSH config. For example: `/home/nixpig/.ssh/id_rsa` |
| `add_to_agent`        | `true` `false` | Whether to add the identity to the running SSH agent when loaded.                                                                                                 |
| `add_to_agent_prompt` | `true` `false` | Whether to prompt before adding loaded identity to agent.                                                                                                         |
