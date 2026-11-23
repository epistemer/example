# Findings: building Confab on neozp

What it was like to build this app as a framework user. Each entry says
what I tried, what happened, and what would have helped. Kinds:

- **bug**: a package does something its contract says it doesn't.
- **contradiction**: two packages, or a package and its docs, disagree.
- **friction**: works as documented, but costs the app time or clarity.
- **gap**: the app needed something no package offers.
- **doc**: the docs or skills misled or left out something needed.

Severity: **high** (blocks or forces a workaround that loses a promise of
the framework), **medium** (a workaround in the app), **low** (a papercut).

Versions: cfgx v0.2.0, i18n v0.3.1, genx v0.2.0, models v0.4.0, errors
v0.1.0, auth v0.3.0, handlers v0.3.0, ui v0.4.0, Go 1.27.4, sqlc v1.29.0.

## Iteration 1 (2026-11-23)

### F1 — cfgx init names the app after the module path (friction, low)

`cfgx init github.com/epistemer/example` made an app called `example`:
binary, cobra `Use`, env prefix `EXAMPLE_`, `example.example.yaml`,
`.gitignore`. There is no flag for the app's name, so naming it `confab`
meant editing the schema's `dsl.App`, `main.go` and `.gitignore`, then
deleting the generated `example.*` artifacts by hand.

*Would help:* `cfgx init -name confab <module>`.

### F2 — a table prefix leaks into every sqlc name; models strips it only from labels (friction, medium; models + sqlc)

With tables named `cf_event`, `cf_talk_view` and `table_prefix: cf_`,
models' labels read `L10nEvent_Name`, but sqlc's types are `CfEvent`,
`CfTalkView`, so the generated listing is `ListCfTalkView` and the column
set `CfTalkViewColumns`. sqlc's only way out is a `rename` per relation,
which models' views refuse to follow (skill, "Limits"). An app on zp's
`zp_` convention therefore reads `models.ZpTalkView` everywhere while its
labels and keys don't carry the prefix: two names for one thing.

*Workaround:* this app drops the prefix (`staff` instead of `user`, which
PostgreSQL reserves).

*Would help:* models documenting the trade-off where `table_prefix` is
introduced, or following sqlc's `rename` for views.

### F3 — a cfgx enum over an sqlc enum gets a second set of labels (friction, low; cfgx + models)

`dsl.Enum("default_level", …, models.TalkLevelBeginner, …)` works and its
getter returns `models.TalkLevel` (nice). But `App.Labels()` writes
`settings.cfp.default_level.beginner` etc., while models already wrote
`enum.talk_level.beginner`. A translator translates "beginner" twice, and a
settings page that shows cfgx's labels disagrees with every other page once
one of them is translated.

*Would help:* cfgx skipping value labels for an enum whose Go type already
has `Label()` (an `option.SelectOption`), or the settings page using the
type's own labels. Here the settings form uses `models.AllTalkLevelValues()`
and so models' labels; cfgx's value labels are dead keys.

### F4 — `handlers gen` refuses the handler package while it imports its own output (friction, high; handlers)

A handler package uses its generated builders everywhere (redirects, links,
actions, table links). On the first run, and every time a route is added
and used in the same edit, `handlers gen` stops with
`the package doesn't type-check: could not import …/adminroutes`. The skill
mentions it in passing ("internal/admin cannot import adminroutes before
adminroutes exists"), but the consequence is that **no real handler
package can be generated from scratch**: every page links to another page.

*Workaround:* `scratchpad/stub`, a throwaway program that replaces every
function body with `panic(0)`, drops package-level values that use the
routes package (transitively) and unused imports; generate on the stubbed
copy, then restore. It took four attempts to get right.

*Would help:* `handlers gen` ignoring type errors that come from the
`-out` package (go/types keeps checking after an error; the parameter types
it needs come from other packages), or reading signatures from the AST
alone when the package doesn't type-check, and reporting the type errors
as warnings.

### F5 — forms' nullable values are pointers; models' views require pgx's pgtype (contradiction, medium; ui/forms + models)

`Field.Ptr()` / `SetPtr()` read and write `*T` for nullable columns, as
the skill says. models' views require `sql_package: pgx/v5`, where sqlc
generates `pgtype.Text`, `pgtype.Date`, `pgtype.Numeric`, … and the obvious
fix, `emit_pointers_for_null_types`, breaks models' column sets (skill:
"Breaks with emit_pointers_for_null_types"). So every nullable column is
converted by hand, both ways: `internal/pg` is 110 lines of that, and every
`Fill`/`Input` calls it.

*Would help:* models generating `Ptr`/`FromPtr` helpers per pgtype it
meets, or supporting `emit_pointers_for_null_types` in column sets; or forms
accepting `interface{ Value() (driver.Value, error) }`-style nullables.

### F6 — ui needs htmx but nothing serves it, and its order matters (gap, medium; ui)

Tables, actions, the remote select and lists need htmx; ui's `theme`
serves its CSS and script but not htmx, and the only copy is embedded in
`ui/catalog`'s private assets. The app vendors it
(`internal/admin/assets/htmx.min.js`) and writes the `htmx-config` meta
itself. Then the order bites: ui's script is a module (deferred), and
`<ui-combobox>` checks `globalThis.htmx` once, when it connects; with htmx
loaded `defer` after `th.Head()` the combobox silently stays a text input
of ids. htmx has to come first in `Document.Head(...)`.

*Would help:* `theme` serving a pinned htmx with the config (an option,
`th.Head(theme.WithHTMX())`), or the combobox upgrading when htmx appears.

### F7 — nav items ignore the actions gate (friction, low; ui)

`layout.NavItem(link)` shows a link whatever the request's gate says, while
`actions.Resolve` drops what the guard refuses. Every app will wrap NavItem
in `if actions.Allowed(ctx, link)` (here `navItem` in layout.templ).

*Would help:* NavItem consulting `actions.Gate` the way actions do, or a
`layout.NavItemIf`.

### F8 — `option.SelectOption` wants a key for data labels (friction, low; models + ui/forms)

The talk form's event select lists rows. `Field.Options` takes
`option.SelectOption`, whose `Label()` returns `i18n.TKer`, a key; an event's
name is data, and `i18n.Static` returns only a `Ter`. The app declares
`dataLabel` with an empty `K()` to fit.

*Would help:* `Label() i18n.Ter` in `SelectOption` (K is never needed for
display), or an `option.Static(value, text)` in models.

### F9 — `rule.PostCode(country)` panics for a country outside its table (bug-ish, medium; ui/forms)

A post code's pattern depends on the country the person picked, so the
rule must be built at check time from a field's value; for a country the
table doesn't have, `PostCode` panics. The app recovers from that panic
(`postCode` in events.go). There's no way to ask whether a country is
known.

*Would help:* `rule.PostCodeFor(func() string)` or a non-panicking lookup
(`rule.HasPostCode(country)`), and the rule passing unknown countries.

### F10 — a form struct can't name a field after a Form method (friction, low; ui/forms)

`type TranslationForm struct { *forms.Form; Text *forms.Field[string] }`
doesn't compile once `f.Text("text")` is called: the field `Text` hides the
embedded method `Text`. Natural field names are taken: Text, Date, Time,
File, Files, Int, Float, Bool, Label, Group, List, Check, Fail, Valid,
Err, Errors, When, Control.

*Would help:* documenting it in the skill; it's the cost of embedding.

### F11 — a table with a remote filter can't be a package-level declaration (friction, medium; ui/tables)

Tables are declared once at package level, like forms' rules, and seal at
first use. `Column.Remote(search, labeler)` takes the `listing.Labeler` at
declaration, and a labeler reads the database, which exists only at run
time. So the talks table is built lazily on the handler (`sync.Once`), and
a declaration mistake there (a `Sort` on a column not in the table) only
panics at the first request. The panic inside `Once` then left the table
nil, and every later request panicked too.

*Workaround:* `H.Check()` builds and checks every table at start.

*Would help:* the labeler reaching the table per request (on `Bind`, or
through the context like the actions gate), so the table stays a
declaration.

### F12 — cfgx flattens a validator's error into English text (contradiction, medium; cfgx + errors)

The settings page saves through `Serve.Update`. cfgx validates with the
same checks as the form (`theme.CheckColor` returns `errors.Key` values),
but reports `cfgx.Errors{Problem{Key, Message: err.Error()}}`: the
translated, public error is gone, and the key is a cfgx path, not a form
field. The page can only show "the settings were refused: …" in English.
The form therefore repeats every check (`checked(theme.CheckColor)`) to get
messages people can read, and maps settings to fields by hand (12 lines of
`Fill`, 12 of `apply`).

*Would help:* `Problem.Err error` keeping the original error, and a
generated mapping from settings to a forms.Form (a settings form is
schema-shaped, like a table is view-shaped).

### F13 — route search returns routes, not links (friction, low; handlers)

`Table.Search` returns `Hit{Route, Text}`. To link a hit the app parses
`Route.Pattern` ("GET /events") and prepends `Prefix`; a route with
parameters can't be linked at all, which is why only parameterless routes
are marked `//h:search` here.

*Would help:* `Hit.Link` (a `handlers.Link`) when the route has no
parameters.

### F14 — the cfgx skeleton bridges slog for cfgx only (friction, low; cfgx init + auth + handlers)

`cfgx init` writes a logrus app with an slog bridge that it passes to
`cfgx.WithLogger`. auth, handlers and i18n log through `slog.Default()`,
which still writes Go's default text format to stderr, so the app's logs
come out in two formats.

*Fix in the app:* `slog.SetDefault(logging.Slog(log))` in main. *Would
help:* the skeleton doing that.

### F15 — an enum array column needs pgx type registration nobody does (gap, medium; models + sqlc)

`topics talk_topic[]` generates `[]models.TalkTopic`, and the first insert
fails: `unable to encode []models.TalkTopic … unknown type (OID 16425)`.
pgx must be told the types on every connection (`LoadTypes` in
`AfterConnect`). models knows every enum and could generate
`models.RegisterTypes(ctx, conn)`.

### F16 — `handlers` has no recovering wrapper: a panic drops the connection (gap, medium; handlers)

The F11 panic reached `net/http`, which logs it and closes the connection;
the browser shows a network error, not a 500 page. This is the open
decision "handlers: recovering wrapper so a render panic becomes a 500
page" on the project board; this app is a second user asking for it.

### F17 — forms may answer 400, but ui's script swaps only 422 (contradiction, medium; ui)

The forms skill: "re-render with the status you choose (400 from
errors.Invalid, or 422)". With `hx-boost` (or any htmx form), ui's script
swaps 4xx only for 422 (actions.js). The talk form re-rendered with
`handlers.Status(err)` — 400 for "the talk must start between …", 409 for
an overlap — and the page silently didn't change. Every admin form here now
answers 422.

*Would help:* the skill saying "422 whenever htmx may be involved", or the
script swapping any 4xx whose response is a page (`HX-Reswap`/a header the
server sets).

### F18 — non-POST actions need htmx, while tables promise to work without script (contradiction, low; ui + handlers)

ui renders a DELETE/PATCH/PUT action as a submit button with
`formmethod=post` (requestAttrs in components/actions.go), so with script
off the browser POSTs to a route handlers registered for DELETE only, and
ServeMux answers 405. handlers has no method override.

*Here:* delete, publish and state changes keep DELETE/PATCH/PUT (to use
every `//h:method`), so they need script.

### F19 — enum values' English labels are their SQL values (friction, low; models)

Badges and selects read "draft", "published", "workshop": lower-case SQL
values, because an enum value's English text "is itself" and SQL has no
comment per enum value. Until someone edits the English in the
translations page, every enum shows raw.

*Would help:* a convention models reads, such as `COMMENT ON TYPE
talk_state IS 'submitted: Submitted; accepted: Accepted'`.

### F20 — the icon set is small for an admin (gap, low; ui)

There is no pencil, trash, external-link, upload, download or calendar
icon; the actions here carry no icons except the public page's eye.
