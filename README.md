[![Workflow Status](https://github.com/syringe-sh/server/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/syringe-sh/server/actions/workflows/build.yml?query=branch%3Amain)
[![Coverage Status](https://coveralls.io/repos/github/syringe-sh/server/badge.svg?branch=main)](https://coveralls.io/github/syringe-sh/server?branch=main)

# `syringe.sh/server`

```
internal/
    - database/ # package database
        - database.go
    - screens/ # package screens
        - register.go
    - stores/ # package stores
        - stores.go
    - models/ # package models
        - user_model.go # etc...
    - services/ # package services
        - users_service.go
        - keys_service.go
    - handlers/ # package handlers
        - http/
            - http.go
        - ssh/
            - ssh.go
```

```go
f, err := tea.LogToFile("tmp/debug.log", "debug")
if err != nil {
    t.Fatalf("unable to open log file:\n%s", err)
}

defer f.Close()
```
