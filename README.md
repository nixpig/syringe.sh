[![Workflow Status](https://github.com/nixpig/syringe.sh/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/nixpig/syringe.sh/actions/workflows/build.yml?query=branch%3Amain)
[![Coverage Status](https://coveralls.io/repos/github/nixpig/syringe.sh/badge.svg?branch=main)](https://coveralls.io/github/nixpig/syringe.sh?branch=main)

# 🔐 `syringe.sh`

Distributed database-per-user encrypted secrets management over SSH protocol.

> **⚠️ This project is a work in progress and not yet ready for general use.**
>
> &nbsp;&nbsp; Feel free to browse the code while it's being developed, but use at your own risk.

SSH (Secure Shell) is a cryptographic network protocol for secure communication between computers over an unsecured network that uses keys for secure authentication. If you've ever `ssh`'d into a remote machine or used CLI tools like `git` then you've used SSH.

syringe.sh uses SSH as the protocol for communication between the client (your machine) and the server (in the cloud).

Your public key is uploaded to the server. Your private key is then used to authenticate when you connect.

Secrets are encrypted locally using your key before being sent to the server and stored in a separate database tied to your SSH key.

Secrets can only be decrypted locally using your private key. Without your private key, nobody can decrypt and read your secrets. It's important you don't lose this, else your secrets will be lost forever.

## TODO

- [x] Confirm authentication before calling cmd, e.g. with unregistered user calling project command results in NPE
- [x] Merge client and server codebases into one
- [x] Share command configuration for both cli client and server
- [x] Secret injection and run passed command
  - [x] Pass a `io.Writer` into `run` so that we can read secrets from it to inject instead of directly printing to terminal out
- [x] Update and enable the disabled unit tests
- [x] Utilise settings file to specify/save which identity to use (viper)
- [x] Load ssh config from file
  - [x] Single identity for host
  - [?] Multiple identities for host - prompt user to choose which to use (have a default - add a flag to just use default (or specific) without prompt)
- [x] Use SSH agent by default - check if key is in agent, if key isn't already in agent when loaded, prompt user to add to agent (add a flag for this too!)
- [ ] Encryption and signing of secrets
  - [ ] RSA
  - [x] Encrypt on client before sending
  - [ ] Decrypt on client
    - [ ] secret get
    - [ ] secret list
    - [ ] inject
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

- [ ] Genericise storage solution so whole thing can be run locally backed by sqlite databases

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

In any case, if you have an SSH agent running and the specified identity is not already loaded into the SSH agent, you will be prompted to add it.

### Settings file

syringe.sh uses a settings file located in your user config directory, for example: `/home/nixpig/.config/syringe/settings`. If this doesn't exist, you will be prompted to create it any time you run a `syringe` command.

The settings file uses a `key=value` format, with each key/value pair on a new line.

#### Settings

| Key                   | Value          | Description                                                                                                                                                        |
| --------------------- | -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `identity`            | `string`       | Path to the SSH identity file to use. Equivalent to the `-i` flag to `ssh` or the `IdentityFile` parameter in SSH config. For example: `/home/nixpig/.ssh/id_rsa`. |
| `add_to_agent`        | `true` `false` | Whether to add the identity to the running SSH agent when loaded.                                                                                                  |
| `add_to_agent_prompt` | `true` `false` | Whether to prompt before adding the identity to the running SSH agent.                                                                                             |
