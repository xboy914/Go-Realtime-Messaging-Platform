package migrations

import "embed"

// Files contains the ordered SQL migrations.
//
//go:embed *.sql
var Files embed.FS
