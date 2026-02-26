package db

import "embed"

// MigrationsFS embeds the SQL migration files into the binary.
// This allows migrations to be run without needing external files deployed.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS
