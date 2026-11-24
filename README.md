# Confab — the neozp example app

Confab is a conference organiser's back office built on all eight neozp
packages: events with venues and ticket tiers, speakers with photos, talks
with slides, a review workflow, a JSON API for partners, settings that
change in the running app, and translations edited in the app (English and
Arabic, right to left).

It has two jobs:

1. **Show what an app on neozp looks like**: declarative (schemas, routes,
   forms and tables are declarations), configurable at run time (theme,
   call-for-papers rules, every text) and typed end to end (an event's id
   can't be passed where a talk's is expected; a route link is a value, not
   a string; settings are getters the compiler checks).
2. **Find what the packages get wrong.** Every difficulty met while
   building it is in [docs/findings.md](docs/findings.md), with a
   workaround and what would have helped.

Where neozp stands after the first iteration, and what comes next:
[docs/neozp-review-2026-11-23.md](docs/neozp-review-2026-11-23.md).

## Run it

Needs Go 1.27 and Docker.

```bash
make db
```

```bash
make migrate
```

```bash
make seed
```

```bash
make serve
```

Open http://localhost:8080 and sign in as `ada@confab.example` with the
password in `dev.env`. The emailed code is written to `outbox/` (auth never
logs codes). Tick "Remember this device" to skip the code next time.

The API lives at `/api/v1` (spec: [api/openapi.yaml](api/openapi.yaml)):

```bash
curl -s localhost:8080/api/v1/events/gophercon-delta/schedule
```

`make test` runs the tests (some need the database), `make check` runs
`genx check`, `handlers vet` and `go vet`.

## How it's built

| Path | What | Package it shows |
|---|---|---|
| `internal/settings/schema` | the configuration: groups, a store, mutable keys, secrets, commands with their own scope | cfgx (+ authschema, themeschema, i18nschema) |
| `db/schema.sql`, `db/query.sql` | tables, enums and three `_view` relations; comments are the English labels | sqlc + models |
| `internal/models` | sqlc's code plus models' labels, enum options, listings and column sets | models, i18n |
| `internal/confab` | the domain: rules between events, speakers and talks; errors people can read | errors |
| `internal/admin` | the back office: `//h:` routes, forms, tables, actions, pages | handlers, ui, auth |
| `internal/admin/adminroutes` | generated: typed URL builders, route labels, permissions | handlers, i18n |
| `api/openapi.yaml`, `internal/api` | the JSON API, its server and generated routes | handlers (openapi), oapi-codegen |
| `genx.yaml` | every generator, in one manifest | genx |

`go generate .` runs templ then genx. genx runs sqlc, cfgx, `handlers gen`
and `handlers openapi`, then models' and handlers' generators, then i18n's
labels generator over every `labels/v1` file.

## What it uses

### Every `//h:` directive

| Directive | Where |
|---|---|
| `key`, `method`, `path`, `desc` | every route |
| `method` GET, POST, PUT, PATCH, DELETE | `talks.move` / `talks.move_put` (PUT), `events.status` (PATCH), `events.delete`, `speakers.delete`, `talks.delete`, `translations.reset` (DELETE) |
| `path` `/` and `{name...}` | `dashboard` (`/`), `file` (`/files/{key...}`), `translations.*` (`/translations/{locale}/keys/{key...}`) |
| `auth public` | sign-in pages, `event_page`, `locale` |
| `auth session` | `sign_out`, `search`, `file` |
| `auth unverified` with a shared `perm` | `dashboard`, `account` (`staff`) |
| `perm` shared by several routes | `events.manage`, `speakers.manage`, `talks.manage`, `talks.review`, `settings.manage`, `translations.manage` |
| `view` | `events`, `speakers`, `talks`, `translations` |
| `search` | `dashboard`, `events`, `events.new`, `speakers`, `speakers.new`, `speakers.import`, `talks`, `talks.new`, `settings`, `translations` |
| `entity` | `events.new`, `events.edit` → `events`; `speakers.*` → `speakers`; `talks.*` → `talks` |
| `body` | `speakers.create` / `.update` (3MiB), `speakers.import_post` (64MiB), `talks.create` / `.update` (110MiB) |
| `param <name> label` | every parameter |

Parameter types: named integers (`ids.EventID`, `ids.SpeakerID`,
`ids.TalkID`), text unmarshalers (`ids.Slug`, `ids.Locale`), enums from
another package (`models.EventStatus`, `models.TalkState`) and a rest
string.

The API covers `security: []`, bearer, `x-auth: session`, `x-permission`,
`x-route-key`, `x-label`, `x-max-body` and the `Error` schema.

### Every form control kind

| Kind | Constructor | Where |
|---|---|---|
| text | `Text` | event name, speaker name, title |
| multiline | `Text().Hint(KindMultiline)` | summary, abstract, biography, translation |
| password | `Text().Raw().Hint(KindPassword)` | sign-in |
| hidden | `Int().Hint(KindHidden)` | the event's version (optimistic locking) |
| number | `Int`, `Float`, `Number[int32]`, `Int().As(minutes)` | quantity, latitude, capacity, duration |
| checkbox | `Checkbox` | online, consent (required = ticked), remember device |
| toggle | `Bool` (three states) | recorded: yes, no or not decided |
| date | `Date` | first and last day, on sale until |
| datetime | `Time(name, layout)` | a talk's start |
| select | `Enum`, `Number[int64]().Options(rows)` | status, level, format, state; the event of a talk |
| multiselect | `Texts().Options(enum options).As(conv)` | topics, read as `[]models.TalkTopic` |
| file | `File` | speakers CSV |
| files | `Files` | speaker photos |
| group | `Group` | venue, active only when not online |
| list | `List` | ticket tiers |
| remote | `Number[int64]().Remote(search, labeler)` | a talk's speaker |
| upload | `Upload` | a speaker's photo |
| uploads | `Uploads` | a talk's slides |

Also `Parsed[T]` (`ids.Slug`, `decimal.Decimal`), `As`, `Placeholder`,
`Help`, `RequiredWith`, `InvalidWith`, `Options`, conditions with `Eq`,
`In`, `Filled`, `Not`, `All` and `Any`, group checks (`Form.Check`), field
errors from the domain (`Field.Fail`), and the rules `MinLen`, `MaxLen`,
`Matches`, `Email`, `URL`, `Phone`, `CountryCode`, `CurrencyCode`,
`PostCode`, `Min`, `Max`, `Between`, `AtLeast`, `AtMost`, `OneOf` (with
`.Err`), `NotBefore`, `NotAfter`, `MinItems`, `MaxItems`, `Each`, `MinSize`,
`MaxSize`, `ContentType`, `Files`, `MaxFiles` and `rule.New`.

### And the rest

- **ui**: the app shell with nav sections, color scheme menu and a
  language switcher; the auth shell; tables over views (search, sort,
  filters, a remote filter, paging, row actions, bulk actions) and over
  i18n's catalogue; actions that are absent when the guard refuses them and
  unavailable with a reason; the action dialog; toasts; replies that
  replace or remove a row; badges, alerts, cards, detail lists, tabs; a
  typed context value (`CurrentTheme`).
- **auth**: password plus emailed code, remembered devices, unverified
  sessions for the dashboard, bearer tokens with refresh for the API,
  permissions and groups (`access.Root`, a "Reviewers" group).
- **i18n**: generated labels from four producers, an Arabic locale seeded
  once, translations edited in the app, white-label branding
  (`branding.title`).
- **errors**: domain errors with kinds and arguments (`Key1`, `Key2`),
  `Wrap`, public messages on pages and in the API's `{"code","message"}`.
- **cfgx**: mutable settings edited on a page and applied at once (theme
  colours, call-for-papers rules), an sqlc enum as a setting, secrets,
  three commands with their own scope.

Not yet: passkeys, an OpenID Connect provider, authenticator apps,
step-up before a dangerous action (`auth.Within`), invitations and
recovery, ui's catalogue of stories, an end-to-end test through a browser
client. They are the next iteration.
