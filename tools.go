//go:build genx

// The packages whose generators genx.yaml switches on, kept in go.mod.
package main

import (
	_ "github.com/epistemer/handlers"
	_ "github.com/epistemer/i18n"
	_ "github.com/epistemer/models"
)
