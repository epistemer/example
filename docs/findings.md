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

*Workaround:* `scripts/stub`, a program that replaces every
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

### F21 — a shared permission's text is its key made readable (friction, low; handlers)

Permissions shared by several routes get the readable key as text:
"Events Manage", "Talks Review", "Api Talks Read". These are the names an
admin sees when granting permissions to a group. There's no directive or
extension to name a permission (`//h:desc` names the route).

*Would help:* `//h:perm events.manage Manage events` (a text after the key)
and the same in `x-permission`.

### F22 — the API and the admin can't share a permission (contradiction, medium; handlers)

"One routes package per permission": declaring `talks.review` in both the
admin's `//h:perm` and the spec's `x-permission` fails generation. So a
reviewer needs `talks.review` *and* `api.talks.review`, granted separately
(seed.go grants both). Both surfaces feed one `routes/v1` domain, so the
permission could be declared once and used by both.

*Would help:* permissions generated once per app (a merged permissions
package), or a route allowed to name a permission another routes package
declares.

### F23 — hidden labels escape a table's scroller and widen the page (bug, medium; ui)

A wide table scrolls inside `.ui-table-scroll`, but the visually hidden
spans in its cells are absolutely positioned and the scroller isn't their
containing block, so they stretch the document: 1747 px of scroll width
in a 1265 px window. In RTL the whole page shifts and the nav goes off
screen. `.ui-table-scroll { position: relative }` fixes it (in
`internal/admin/assets/confab.css`).

### F24 — after a boosted language change, `<html lang dir>` stay the old ones (bug, medium; ui + htmx)

`layout.Document` sets `lang` and `dir` from the locale, and `BodyAttrs`
takes `hx-boost`. With boost on, switching to Arabic swapped the body,
showed Arabic text, and kept `<html dir="ltr" lang="en">`: an Arabic page
laid out left to right until a full reload.

*Would help:* ui's script copying `lang` and `dir` from a boosted
response's `<html>` (htmx swaps the body only), or the docs saying to mark
the language switcher `hx-boost="false"`.

### F25 — `Reply.Go` after a toast panics, after the save committed (friction, medium; ui)

`components.Reply(w, r).Toast(saved).Go(link)` panics by design ("the
redirect drops them; the app's flash carries a message"), but:

- it panics after the handler saved, and with no recovering wrapper (F16)
  the connection dropped; Chrome sent the POST again, and the second save
  failed the version check with "someone changed this event while you
  edited it". The real cause took a while to find.
- there is no flash: "the app's flash" is the app's to write
  (`internal/admin/flash.go`, a cookie with an i18n key).

*Would help:* make it a compile error: `Toast`, `Replace`, `Remove` and
`Refresh` return a type without `Go`; and a flash in ui (the toast survives
the redirect).

### F26 — `KindToggle` is a three-state select, not a switch (doc, low; ui/forms)

The name reads as an on/off switch. It's `Bool`'s Yes/No/unset select.
`Checkbox().Hint(KindToggle)` is accepted and silently turns a two-state
checkbox into a select with an empty "Choose…" choice.

*Would help:* naming it `KindYesNo`, and Hint refusing it on a Checkbox.

### F27 — left-to-right values read backwards in right-to-left pages (bug, low; ui/inputs)

In Arabic the settings page shows `c62828#` for `#c62828`. Colours, emails,
URLs, phone numbers, codes and slugs are left to right whatever the page;
the rules know which fields hold them (`Email`, `URL`, `Phone`, `Matches`)
but inputs get no `dir`.

*Would help:* `dir="ltr"` (or `dir="auto"`) from those rules through
`Control.Attrs`.

### F28 — an enum path parameter takes any string (gap, medium; handlers + models)

`EventsStatus(id ids.EventID, status models.EventStatus)` gives a typed
builder, which is excellent; but reading the path, handlers converts any
string to the named type, so `/events/3/status/bogus` reaches PostgreSQL
and answers 500. The API side validates against the spec's enum; the HTML
side doesn't.

*Would help:* models generating `UnmarshalText` on enum types (refusing
other values), which handlers already prefers.

### F29 — i18n's errors aren't public (gap, medium; i18n + errors)

`/admin/translations/xx` answers 500 "something went wrong":
`i18n.ErrUnknownLocale` has no kind and no message for people (i18n can't
import errors, which imports i18n). The same goes for `ErrFormat`: a
translation whose verbs don't fit is refused with "something went wrong"
on the field, when the person needs to be told to keep `%d`.

*Workaround:* none yet in the app (the page should map them).

*Would help:* i18n's sentinels implementing `i18n.Ter` and a `Kind()`
(errors reads them structurally, no import needed).

### F30 — `AssertLabelled` wants a label on a hidden field (friction, low; ui/forms)

The event's `version` is a hidden input nobody sees; `formstest.AssertLabelled`
fails until it has a label.

### F31 — templ isn't part of genx (friction, low; genx + templ)

`handlers gen` type-checks the admin package, templates included, so
`templ generate` must run first; `genx check` can't see a stale
`_templ.go`. main.go runs `templ generate` then `genx` from `go:generate`.

### F32 — genx can't run one producer (friction, low; genx)

Bootstrapping (F4) meant commenting producers out of `genx.yaml`. A
`genx run <producer>` would have helped, as would a way to declare that a
producer's input may not type-check yet.

## What worked well

Not every note is a problem. These were better than expected:

- **Typed links everywhere.** `adminroutes.EventsStatus(id,
  models.EventStatusPublished)` can't be built with a talk's id or a
  misspelled status; actions, nav items, table links and redirects all take
  the same value, and a route the user can't reach makes its action vanish
  without a second permission check.
- **An sqlc enum as a cfgx setting.** `CFPDefaultLevel() models.TalkLevel`
  flows straight into the talk form's `Set`.
- **`Parsed[ids.Slug]`** reports the type's own error (`errors.confab.bad_slug`),
  so a domain type validates itself in forms and in route parameters.
- **`As`** turned an `Int` into a `time.Duration` and a `Texts` into
  `[]models.TalkTopic`, with rules on the first stage and the typed value
  on the second.
- **One translations page for everything.** Route labels, table and enum
  labels, settings labels, form texts and errors all land in one catalogue;
  the page is a ui table over `i18n.Search` with hand-made columns.
- **The API's errors.** Domain errors declared once come out of the API as
  `{"code": "errors.confab.not_scheduled", "message": …}` with a 409, and
  the spec's validation answers 400 before the handler runs.
- **Conditions run on both sides.** The workshop's seats appear in the
  browser and are cleared on the server from one `When`.

## Summary

| # | Package | Kind | Severity |
|---|---|---|---|
| F1 | cfgx | friction | low |
| F2 | models + sqlc | friction | medium |
| F3 | cfgx + models | friction | low |
| F4 | handlers | friction | **high** |
| F5 | ui/forms + models | contradiction | medium |
| F6 | ui | gap | medium |
| F7 | ui | friction | low |
| F8 | models + ui/forms | friction | low |
| F9 | ui/forms | bug | medium |
| F10 | ui/forms | friction | low |
| F11 | ui/tables | friction | medium |
| F12 | cfgx + errors | contradiction | medium |
| F13 | handlers | friction | low |
| F14 | cfgx init | friction | low |
| F15 | models | gap | medium |
| F16 | handlers | gap | medium |
| F17 | ui | contradiction | medium |
| F18 | ui + handlers | contradiction | low |
| F19 | models | friction | low |
| F20 | ui | gap | low |
| F21 | handlers | friction | low |
| F22 | handlers | contradiction | medium |
| F23 | ui | bug | medium |
| F24 | ui | bug | medium |
| F25 | ui | friction | medium |
| F26 | ui/forms | doc | low |
| F27 | ui/inputs | bug | low |
| F28 | handlers + models | gap | medium |
| F29 | i18n + errors | gap | medium |
| F30 | ui/forms | friction | low |
| F31 | genx + templ | friction | low |
| F32 | genx | friction | low |

## Closing iteration 1 (2026-11-25)

Confab moved to handlers v0.4.0, ui v0.5.0, models v0.5.1, cfgx v0.3.1,
i18n v0.4.0 and errors v0.2.0, which build epistemer's answers of
2026-11-24 ([neozp decisions](https://github.com/epistemer/neozp/blob/main/docs/decisions.md)),
and deleted the workarounds. A finding is closed when its workaround is
gone (neozp roadmap I1).

| # | Status | How |
|---|---|---|
| F1 | closed | cfgx v0.3.0: `cfgx init -name` (A17) |
| F2 | closed | models documents `table_prefix`'s trade-off |
| F3 | closed | cfgx v0.3.0 takes an enum's labels from its Go type (A16); the dead `settings.cfp.default_level.*` keys are gone |
| F4 | closed, low | handlers v0.4.0's error names the missing builder; the order is handler, `handlers gen`, then the link. `scripts/stub` deleted |
| F5 | closed | models v0.5.0 reads `*T` and overridden types; sqlc's config now uses pointers and overrides, as zp's does. `internal/pg` deleted |
| F6 | **open** | the combobox now upgrades when htmx arrives late (ui v0.5.0); serving htmx waits for the discussion of forking htmx (U2). The vendored copy stays |
| F7 | closed | `NavItem` asks the actions gate (ui A9); `navItem` only marks the current page |
| F8 | closed | `option.Static`; `dataLabel` deleted |
| F9 | closed | `PostCode` passes unknown countries (ui A9); the recover deleted |
| F10 | closed | documented |
| F11 | closed | documented: build such tables at startup, as `H.Check` does |
| F12 | closed | `Problem.Err` (cfgx A15): the settings page shows themeschema's own error on the field; the repeated checks deleted |
| F13 | **open** | deferred with epistemer's parameter-form idea (H5) |
| F14 | closed | cfgx's skeleton calls `slog.SetDefault` |
| F15 | closed | `models.RegisterTypes`; the hand-written `LoadTypes` deleted |
| F16 | closed | `Register` and `Mount` recover; `Recover` and `ErrPanic` deleted |
| F17 | closed | documented: 422 whenever htmx may be involved |
| F18 | **open** | deferred to the discussion of forking htmx (H6) |
| F19 | closed | enum values read as words ("Submitted"), neozp D7 |
| F20 | closed | Lucide generated; edit, delete, calendar and external-link actions carry icons |
| F21 | closed | permissions carry their text (`//h:perm events.manage Manage events`) |
| F22 | closed | the API takes `talks.review from adminroutes`; the seed grants one permission, and `TestReviewPermissionShared` holds it |
| F23 | closed | ui v0.5.0; the CSS rule deleted from `confab.css` |
| F24 | closed | checked: a boosted switch to Arabic sets `dir="rtl" lang="ar"` |
| F25 | closed | `Reply(...).Toast(m).Go(l)` carries the toast through handlers' flash; `flash.go` deleted (see F34) |
| F26 | closed | `<ui-tristate>` renders `KindToggle` (ui A9) |
| F27 | closed | text inputs carry `dir="auto"` |
| F28 | closed | enums refuse unknown text: `/events/1/status/bogus` answers 400 |
| F29 | closed | i18n v0.4.0 and errors v0.2.0: `/translations/xx` answers 404 "there's no locale with the code xx" (see F35) |
| F30 | closed | the hidden version field has no label |
| F31 | closed | documented in genx |
| F32 | dropped, then met again | see F33 |

### F33 — a change to sqlc's options breaks the pipeline in two places (friction, medium; genx + models + handlers + cfgx)

Switching sqlc to `emit_pointers_for_null_types` and overrides changed
every nullable type. `go tool genx` then failed twice before any of models'
generators ran:

1. cfgx's producer compiles the settings schema, which imports the models
   package, and models' `zz_models_views.go` was still written for pgtype.
2. With that file deleted, `handlers gen` type-checks the admin package,
   which uses the column sets models hadn't written yet.

genx runs every producer before any trivial generator, so a producer
can't depend on a generator's output. Taking handlers out of `genx.yaml`
to break the loop (F32's workaround) made genx delete handlers' outputs
and drop them from `genx.lock`. Once restored, genx refused to overwrite
them ("exists and isn't in genx.lock"), until the lock was mended by hand.

*Would help:* genx ordering producers after the generators whose outputs
their inputs import, or a way to declare that dependency; and
`genx run <producer>…` (G1, dropped on 2026-11-24) after all.

### F34 — a toast carried across a redirect is always "info" (friction, low; ui)

`Reply.Toast(msg)` takes no tone, and Reply flashes it at level `info`.
"The event was saved" used to be a success alert; now it's an
"Information:" toast.

*Would help:* a tone on `Toast` (or `Reply.Success(msg)`), passed to the
flash's level.

### F35 — an app that finds a locale missing has no public error to return (friction, low; i18n)

`Translator.Locale(code)` answers with a bool. The documented
`ErrUnknownLocale` is the bare variable, which isn't public (i18n §8.5),
so returning it still shows a 500. Confab now asks `Search`, which returns
the public error. For an unknown key it declares i18n's key again with
`errors.Key1` (`ErrNoSuchKey`), as §8.5 allows.

*Would help:* `Translator.Locale` returning the public error, or i18n
exporting constructors for its public errors.
