package migrations

import "embed"

// FS embeds the versioned MySQL migration set into the AxisRelay binary.
//go:embed *.sql
var FS embed.FS
