// Package db holds the schema, for the migrate command.
package db

import _ "embed"

//go:embed schema.sql
var Schema string
