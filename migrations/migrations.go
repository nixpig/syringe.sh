package migrations

import "embed"

//go:embed app/*.sql
var App embed.FS

//go:embed user/*.sql
var User embed.FS
