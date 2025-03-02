package migrations

import "embed"

//go:embed sql/*.sql
var Migrations embed.FS
