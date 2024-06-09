[![Workflow Status](https://github.com/syringe-sh/server/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/syringe-sh/server/actions/workflows/build.yml?query=branch%3Amain)
[![Coverage Status](https://coveralls.io/repos/github/syringe-sh/server/badge.svg?branch=main)](https://coveralls.io/github/syringe-sh/server?branch=main)

# `syringe.sh/server`

## `user`

Manage your user account(s).

- `syringe user register`
- `syringe user delete`

## `project`

Manage your projects.

- `syringe project add <name>`
- `syringe project remove <name>`
- `syringe project list`
- `syringe project rename <old name> <new name>`

## `environment`

Manage your environments.

- `syringe environment add [-p <project>] <name>`
- `syringe environment remove [-p <project>] <name>`
- `syringe environment list [-p <project>]`
- `syringe environment rename [-p <project>] <old name> <new name>`

## `secret`

Manage your secrets.

- `syringe secret set [-p <project> -e <environment>] <key> <value>`
- `syringe secret get [-p <project> -e <environment>] <key>`
- `syringe secret list [-p <project> -e <environment>]`
- `syringe secret remove [-p <project> -e <environment>] <key>`

## `key`

Manage your SSH auth keys.

- `syringe key add`
- `syringe key remove [-k <public key>]`

---

- Register user
- Add public key
- Create user database

```go
f, err := tea.LogToFile("tmp/debug.log", "debug")
if err != nil {
    t.Fatalf("unable to open log file:\n%s", err)
}

defer f.Close()
```
