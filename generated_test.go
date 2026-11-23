package main

import (
	"testing"

	"github.com/epistemer/genx/genxtest"
)

// Every generated file is current: sqlc's request, models' labels, enums
// and views, cfgx's settings, both route surfaces, and i18n's labels.
func TestGenerated(t *testing.T) { genxtest.AssertGenerated(t) }
