[![Workflow Status](https://github.com/nixpig/syringe.sh/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/nixpig/syringe.sh/actions/workflows/build.yml?query=branch%3Amain)
[![Coverage Status](https://coveralls.io/repos/github/nixpig/syringe.sh/badge.svg?branch=main)](https://coveralls.io/github/nixpig/syringe.sh?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/nixpig/syringe.sh)](https://goreportcard.com/report/github.com/nixpig/syringe.sh)

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

- [ ] Proper good refactor and tidy-up!! Best practices around configuration management with Viper.
- [ ] Update syringe.sh domain
- [ ] Set up demo server on syringe.sh
- [ ] E2E tests with the CLI (or SSH?) client, including a couple like trying to create secrets for a non-existent project or environmnet

  - Tests for database package and migrations
  - Work out how to start/stop server asynchronously and run tests. Could be containerised using testcontainers?
  - Just use testcontainers??

## CLI

### Installation

1. Download the package for your operating system and architecture from the [releases](https://github.com/nixpig/syringe.sh/releases) page and extract to a directory in your path, e.g.

   ```
   $ wget -qO- https://github.com/nixpig/syringe.sh/releases/download/0.0.9/syringe.sh_syringe_0.0.9_linux_amd64.tar.gz | tar -xzvf - -C /usr/bin
   ```

1. Run the `syringe` command to get started.

> [!NOTE]
>
> Without additional configuration, the `syringe` command will connect to the demo server at syringe.sh.
>
> Feel free to have a play around there before you decide whether to spin up your own server.

### Usage

> [!TIP]
>
> Run `syringe help` to view documentation for all available commands and example usage.

### Supported SSH key types

The following key types are supported for the syringe client.

- RSA

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
| `hostname` | `string` | The hostname of the server to connect to (default: `syringe.sh`).                                                                                                  |
| `port`     | `number` | The port the server is running on (default: `22`).                                                                                                                 |

#### Example settings file

```bash
identity=$HOME/.ssh/id_rsa
hostname=localhost
port=23234
```

## Server

The recommended method of running the server is using Docker.

An example [`Dockerfile`](https://github.com/nixpig/syringe.sh/blob/main/Dockerfile) and [`docker-compose.yml`](https://github.com/nixpig/syringe.sh/blob/main/docker-compose.yml) are included in the repository.

## Disclaimers

The public syringe.sh server is for demo purposes and may not be actively monitored or maintained. You absolutely should **not** store any secret or private data there.

You are responsible for your own security. It is up to you to evaluate the suitability of this software before using it and to take any necessary measures to secure your data to prevent unauthorized access.
