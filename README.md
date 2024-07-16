[![Workflow Status](https://github.com/nixpig/syringe.sh/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/nixpig/syringe.sh/actions/workflows/build.yml?query=branch%3Amain)
[![Coverage Status](https://coveralls.io/repos/github/nixpig/syringe.sh/badge.svg?branch=main)](https://coveralls.io/github/nixpig/syringe.sh?branch=main)

# 🔐 `syringe.sh`

Self-hostable distributed database-per-user encrypted secrets management over SSH.

> [!CAUTION]
>
> This project is a work in progress and not yet ready for general use.
>
> Feel free to browse the code while it's being developed, but use at your own risk.

SSH (Secure Shell) is a cryptographic network protocol for secure communication between computers over an unsecured network that uses keys for secure authentication. If you've ever `ssh`'d into a remote machine or used CLI tools like `git` then you've used SSH.

syringe.sh uses SSH as the protocol for communication between the client (your machine) and the server (in the cloud).

Your public key is uploaded to the server. Your private key is then used to authenticate when you connect.

Secrets are encrypted locally using your key before being sent to the server and stored in a separate database tied to your SSH key.

Secrets can only be decrypted locally using your private key. Without your private key, nobody can decrypt and read your secrets. It's important you don't lose this, else your secrets will be lost forever.

```
┌────────────────────────────────┐
│ STDIN                          │
│ syringe secret set SKEY s3cr3t │
└─────┬──────────────────────────┘
      │
  ┌───▼────────────────┐                        ┌─────────────────┐
  │      ┌────────────┐│       Encrypted        │┌───────┐        │
  │ CLI  │ 🔐 Encrypt ├─────────────────────────►│ Store │ Server │
  │      └────────────┘│          SSH           │└───┬───┘        │
  └────────────────────┘                        └────│────────────┘
                                                ┌────▼────┐
                                                │ User DB │┐  K: SKEY
                                                └┬────────┘│  V: <encrypted>
                                                 └─────────┘

┌─────────────────────────┐
│ STDIN                   │
│ syringe secret get SKEY │
└─────┬───────────────────┘
      │
  ┌───▼────────────────┐                        ┌─────────────────┐
  │      ┌────────────┐│       Encrypted        │┌───────┐        │
  │ CLI  │ 🔓️ Decrypt │◄────────────────────────►│ Store │ Server │
  │      └────┬───────┘│          SSH           │└───────┘        │
  └───────────│────────┘                        └─────────────────┘
         ┌────▼─────┐
         │ STDOUT   │
         │ s3cr3t   │
         └──────────┘

```

Secrets can be managed using 'projects' and 'environments'.

## TODO

### P1

- [ ] Build and publish artifact on GitHub
- [ ] Install script that downloads cli binary into path.
- [ ] E2E tests with the CLI (or SSH?) client, including a couple like trying to create secrets for a non-existent project or environmnet
  - Work out how to start/stop server asynchronously and run tests. Could be containerised using testcontainers?
  - Just use testcontainers??
- [ ] Genericise storage solution so whole thing can self-hosted and backed by sqlite databases
- [ ] Build and publish deployable Docker image for server

### P2

- [ ] Update syringe.sh domain
- [ ] Set up demo server on syringe.sh
- [ ] Email confirmation on new user registration?
- [ ] Accept spaces in secret values when quoted
- [ ] Improve error handling, errors and messaging
- [ ] Proper good refactor and tidy-up!! Best practices around configuration management with Viper.

### P3

- [ ] Formatted and --plain output of commands, e.g. table when listing secrets
- [ ] Add functionality to 'link' local directories/projects to specific project/environment - save in config file
- [ ] Remove use of third-party package for SSH client (in CLI client)
- [ ] Pull the Turso stuff out into separate SDK package??
- [ ] Create a wrapper package around the various SSH related stuff like config and known hosts
- [ ] Add multiple keys for the same user
- [ ] Allow deletion of user and data
- [ ] Add 'syringe config' command to create/update config file, e.g. `syringe config set hostname localhost`?

## Installation

### CLI

1. Download the corresponding binary from the [releases](https://github.com/nixpig/syringe.sh/releases) page and put it into your path.
1. Rename the binary to `syringe`.
1. Run `syringe` from your terminal.

> [!NOTE]
>
> In future a more simple install script will be available.

### Server

1. Download the corresponding binary from the [releases](https://github.com/nixpig/syringe.sh/releases) page and put it into your path.
1. Package and deploy per your requirements.

> [!NOTE]
>
> In future this will be packaged for easy configuration and deployment.

## Supported SSH key types

- RSA

## Usage

### Specifying an identity

An _identity_ is a path to an SSH key, for example `~/.ssh/id_rsa`.

An identity must be specified to connect over SSH and to encrypt/decrypt secrets.

The identity to use is selected with the following order of precedence.

1. The `--identity` flag.
1. The `identity` property in [settings file](#settings-file).
1. The running SSH agent, if available.

If you have an SSH agent running and the specified identity is not already loaded into the SSH agent, it will be added.

Note: when using the SSH agent directly (i.e. identity not specified as flag or in settings), the syringe.sh host must also be configured in SSH config.

### Settings file

syringe.sh uses a settings file located in your user config directory, for example: `/home/nixpig/.config/syringe/settings`. If this doesn't exist, it will be created for you when you run any `syringe` command.

The settings file uses a `key=value` format, with each key/value pair on a new line.

| Key        | Type     | Description                                                                                                                                                        |
| ---------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `identity` | `string` | Path to the SSH identity file to use. Equivalent to the `-i` flag to `ssh` or the `IdentityFile` parameter in SSH config. For example: `/home/nixpig/.ssh/id_rsa`. |
| `hostname` | `string` | (default: `syringe.sh`)                                                                                                                                            |
| `port`     | `number` | (default: `22`)                                                                                                                                                    |

#### Example settings file

```bash
identity=$HOME/.ssh/id_rsa
hostname=localhost
port=23234
```
