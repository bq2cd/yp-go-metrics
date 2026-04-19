// Package migrations embeds all SQL files in the current package directory into compiled Go binary.
// This makes it easier to distribute database migrations alongside the binary, thus ensuring their
// proper version.
package migrations

import "embed"

// FS is the embedded filesystem with SQL files from the current directory.
//
//go:embed *.sql
var FS embed.FS
