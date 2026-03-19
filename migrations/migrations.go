package migrations

import "embed"

// Путь "." означает текущую директорию (папку migrations)
//
//go:embed *.sql
var MigrationsDir embed.FS
