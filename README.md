[![Workflow Status](https://github.com/syringe-sh/server/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/syringe-sh/server/actions/workflows/build.yml?query=branch%3Amain)
[![Coverage Status](https://coveralls.io/repos/github/syringe-sh/server/badge.svg?branch=main)](https://coveralls.io/github/syringe-sh/server?branch=main)

# `syringe.sh/server`

```go
f, err := tea.LogToFile("tmp/debug.log", "debug")
if err != nil {
    t.Fatalf("unable to open log file:\n%s", err)
}

defer f.Close()
```
