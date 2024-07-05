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
- [ ] Add unit tests for other areas
- [ ] Encryption and signing of secrets
  - [ ] RSA
  - [ ] OpenSSH
  - [ ] ECDSA
  - [ ] ED25519
- [ ] E2E tests with the CLI (or SSH?) client, including a couple like trying to create secrets for a non-existent project or environmnet
  - Work out how to start/stop server asynchronously and run tests. Could be containerised using testcontainers?
  - Just use testcontainers??
- [ ] Add functionality to 'link' local directories/projects to specific project/environment
- [x] Explicit (not implicit) user registration
- [ ] Improve error handling, errors and messaging
- [x] Exit codes on error
- [ ] Accept spaces in secret values
- [ ] Remove use of third-party package for SSH client (in CLI client)
- [ ] Proper good refactor and tidy-up (primarily of database stuff)
- [ ] Pull the Turso stuff out into separate SDK package??
- [ ] Make it work also with a local database
  - [ ] Add syncing of local and remote databases
