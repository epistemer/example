// Package schema declares confab's settings. After a change, run
// go generate ./... to update the generated code, the example config,
// its JSON Schema and docs/CONFIG.md.
package schema

import (
	"time"

	"github.com/epistemer/auth/authschema"
	"github.com/epistemer/cfgx/dsl"
	"github.com/epistemer/i18n/i18nschema"
	"github.com/epistemer/ui/theme/themeschema"

	"github.com/epistemer/example/internal/models"
)

var Log = dsl.Group("log", "Logging",
	dsl.Enum("level", "Minimum log level", "debug", "info", "warn", "error").Default("info"),
	dsl.Enum("format", "Output format", "text", "json").Default("text"),
)

var Server = dsl.Group("server", "HTTP server",
	dsl.String("addr", "Address to listen on").Default("localhost:8080"),
	dsl.String("public_url", "Address people reach the app at").Default("http://localhost:8080"),
	dsl.Duration("shutdown_timeout", "How long to wait for requests to finish when stopping").Default(10*time.Second),
)

var bolt = dsl.Path("bolt", "Store file: settings, translations and sign-ins").Default("confab.db")

var Storage = dsl.Group("storage", "Storage",
	bolt,
	dsl.Path("files", "Directory of uploaded files").Default("files"),
)

var DB = dsl.Group("db", "Database",
	dsl.Secret("url", "PostgreSQL connection URL").Required(),
).GoName("DB")

var Mail = dsl.Group("mail", "Outgoing mail",
	dsl.Path("outbox", "Directory each message is written to, one file per message").Default("outbox"),
)

// The library groups: each pins its own names, so the generated settings
// satisfy auth.Config, passkey.Config, theme.Config and i18n.Config.
var (
	Auth  = authschema.Group(authschema.Passkey())
	Theme = themeschema.Group()
	I18n  = i18nschema.Group()
)

// CFP is the call for papers: settings an organiser changes in the running
// app (Mutable), on the settings page.
var CFP = dsl.Group("cfp", "Call for papers",
	dsl.Bool("closed", "Refuse new talks").Mutable(),
	dsl.Int("max_talks", "Talks a speaker may give at one event").Default(3).Range(1, 10).Mutable(),
	dsl.Enum("default_level", "Level a new talk starts at",
		models.TalkLevelBeginner, models.TalkLevelIntermediate, models.TalkLevelAdvanced).
		Default(models.TalkLevelBeginner).Mutable(),
	dsl.Duration("default_duration", "Length a new talk starts at").Default(30*time.Minute).Range(5*time.Minute, 4*time.Hour).Mutable(),
).GoName("CFP")

// Demo seeds an organiser who holds every permission.
var Demo = dsl.Group("demo", "Demo data",
	dsl.String("email", "Organiser's email address").Default("ada@confab.example"),
	dsl.String("name", "Organiser's name").Default("Ada Organiser"),
	dsl.Secret("password", "Organiser's password").Required(),
)

var App = dsl.App("confab", "Confab, a conference organiser's back office",
	Log, Server, Storage, DB, Mail, Auth, Theme, I18n, CFP, Demo).
	ConfigFile("confab", ".").
	Global(Log).
	Store(bolt).
	Labels()

var (
	Serve   = dsl.Command("Serve").Uses(Server, Storage, DB, Mail, Auth, Theme, I18n, CFP)
	Migrate = dsl.Command("Migrate").Uses(DB)
	Seed    = dsl.Command("Seed").Uses(Storage, DB, Mail, Auth, Demo).WithStore()
)
