# neozp after Confab's first iteration: where it stands, and what next

2026-11-23. A review of the eight neozp packages, of Confab as the first app
built on all of them, and of zp as the record of what the framework must do.
It ends with next steps and the questions they need answered.

zp is described here by what it does, never by its business, in the way
ui's [inventory](https://github.com/epistemer/ui/blob/main/docs/ui-inventory.md)
does it. Nothing here plans moving zp onto neozp, which is out of neozp's
scope (decided 2026-11-16).

## 1. Summary

1. **The packages are deep and the seams between them are shallow.** In 16
   days, eight packages reached v0.x. Each one was designed, reviewed, locked
   and released on its own. The first app on all eight found 32 problems,
   and 15 of them sit between two packages, or between a package and a
   tool it builds on. Only 17 sit inside one package's own surface.
2. **The ui layer met handlers and auth for the first time in Confab.**
   - handlers' example combines seven packages but not ui.
   - ui's example uses errors, i18n and models, but not handlers, auth or
     cfgx.

   Five findings sit where ui meets handlers (F7, F16, F17, F18, F25),
   three of them medium. F16 and F25 together caused the iteration's worst
   bug: a POST that had already committed was sent again.
3. **Nobody owns a seam.** A fix that touches two packages needs two
   amendments under two sets of terms, and some needs wait on each other in
   a loop. For example, handlers waits on auth for API keys, while auth
   waits for an app to ask. neozp-wide decisions have no home of their own:
   neozp ADR 0001 lives in genx.
4. **Confab shows the promise holds:** routes, permissions, labels, tables
   and forms are declarations, links are typed and settings change at run
   time.
   - It also shows three layers no package owns:
     - a request's lifecycle (recovery, flash, error page, status codes,
       htmx);
     - values that cross package boundaries (enums, nulls);
     - a generation pipeline that loops back on itself.
   - About 460 lines of Confab are workarounds every app would copy.
5. **zp needs things no package plans:** jobs, file storage, message
   delivery and HTTP plumbing. Some are listed in a backlog but never
   scheduled. Some are explicitly disowned ("another library's job") with
   no other library named. The rest are mentioned nowhere.
6. **Next:**
   1. Stop adding depth to the packages and close the seams.
   2. Count a finding as done when Confab deletes its workaround.
   3. Make integration continuous through Confab's CI.
   4. Then widen Confab in three iterations, not one: consolidation, then
      people, then background work.
   5. New packages come from what Confab hand-rolled, with zp as the
      corroboration.

## 2. What this is based on

- The eight repositories at their latest tags: docs, roadmaps, status logs,
  ADRs and each one's CI workflow. Also the neozp project board, with 42
  items.
- Confab ([README](../README.md)), with the 32 findings in
  [findings.md](findings.md).
- zp's code, read as a record of needs: its framework-level folders,
  `go.mod`, templates and Makefile. Its wish lists (`ARCHITECTURE.md`,
  `roadmap.md`) are used only where the code confirms them.

## 3. neozp so far

### 3.1 What exists

Rough counts. Go is hand-written `.go` files, excluding tests, `zz_*`,
`*_templ.go` and ui's spike. Docs are every `docs/*.md`.

| Package | First commit | Latest | Go | Tests | Docs | ADRs | Reviews | Commits |
|---|---|---|---:|---:|---:|---:|---:|---:|
| cfgx | 2026-11-07 | v0.2.0 | 23,799 | 16,964 | 5,427 | 9 | 39 | 189 |
| i18n | 2026-11-12 | v0.3.1 | 2,581 | 3,091 | 2,135 | 3 | 10 | 50 |
| genx | 2026-11-13 | v0.2.0 | 13,103 | 12,189 | 2,714 | 9 | 10 | 58 |
| models | 2026-11-14 | v0.4.0 | 1,906 | 2,126 | 2,465 | 8 | 13 | 42 |
| errors | 2026-11-14 | v0.1.0 | 647 | 848 | 2,189 | 2 | 3 | 13 |
| auth | 2026-11-14 | v0.3.0 | 14,601 | 14,556 | 5,264 | 10 | 27 | 89 |
| handlers | 2026-11-18 | v0.3.0 | 9,010 | 6,831 | 2,719 | 7 | 7 | 71 |
| ui | 2026-11-21 | v0.4.0 | 27,668 | 24,025 | 10,727 | 14 | 23 | 152 |
| **Total** | | | **93,315** | **80,630** | **33,640** | **62** | **132** | **664** |

None of the eight repositories has an open issue or PR. Every backlog is
either "as apps ask" or gated on a decision.

### 3.2 What went right

- **Dependencies point the right way.**
  - cfgx, i18n and genx depend on no other neozp package.
  - models and errors depend only on i18n.
  - auth, handlers and ui sit on top.
  - ui depends on neither handlers nor auth. It takes `interface{ URL()
    string }` and a gate instead.

  The rule "a package that needs something a UI gave it in zp declares it
  as an interface it owns" made this possible.
- **Labels as data work across producers.** Four generators write
  `labels/v1` files: models, handlers (two of them) and cfgx. One
  translations page in Confab edits all of them, along with forms' texts
  and errors.
- **Each package has an example in CI**, and some combine several packages.
  handlers' shop uses seven, and auth's uses four.
- **The design process catches design problems early:** a brief, a cross-model
  review, a spike, a locked interface, then amendments. ui's rich inputs went
  from design to a locked contract to a release in two days without a
  breaking change.

### 3.3 What the process missed: integration came last

Every interface was locked before a real app used it. The examples are toy
shops, and each one tests its own package's story. handlers' shop does
combine seven packages, but ui's junction with the rest, and every junction
under a real app's needs, were first exercised by Confab:

| Where the finding sits | Findings | Count |
|---|---|---:|
| Between two neozp packages | F3, F5, F8, F12, F14, F18, F28, F29 | 8 |
| Between neozp and a tool it builds on (sqlc and pgx, htmx, templ) | F2, F6, F15, F24, F31 | 5 |
| Between two halves of one package (ui's forms and its script; handlers' admin and API sides) | F17, F22 | 2 |
| Inside one package | the other 17 | 17 |

This has happened before. genx's [friction log](https://github.com/epistemer/genx/blob/main/docs/friction.md)
(M8.4) recorded the same kind of problem when cfgx and i18n first
generated through genx. Every first meeting between packages has
produced seam findings. Integration is something to run continuously, not
an event at the end.

Release skew points the same way. models and errors require i18n v0.2.0,
while auth, handlers and ui require v0.3.1. Go picks v0.3.1 for every app,
so models' and errors' CI never tests the version apps actually run.

### 3.4 Nobody owns a seam

- Each package's terms say a change another package needs "goes through
  that package's own terms". So a seam fix is two amendments, reviewed
  separately, and each package can reasonably wait for the other:
  - **API keys:** auth's interface §9 leaves them out "as apps ask", and
    handlers' backlog item 3 waits "once auth covers them".
  - **The recovering wrapper:** ui's brief §8 hands it to handlers, and it
    has sat as an open Decision on the board since ui's design.
- neozp-wide decisions have no home. neozp ADR 0001 is
  [in genx](https://github.com/epistemer/genx/blob/main/docs/adr/0001-generation-pipeline.md),
  there is no neozp-wide roadmap, and the board tracks work without
  holding decisions.
- **The v1 gates disagree, and some are stale.**
  - errors'
    [roadmap](https://github.com/epistemer/errors/blob/main/docs/roadmap.md?plain=1#L20-L21)
    and auth's
    [roadmap](https://github.com/epistemer/auth/blob/main/docs/roadmap.md?plain=1#L20-L21)
    still say "`v1.0.0` waits until zp's business code runs on it".
    genx's [roadmap](https://github.com/epistemer/genx/blob/main/docs/roadmap.md?plain=1#L309-L310)
    says the same of models ("zp's move"). The 2026-11-16 rule says no
    release waits for that.
  - cfgx waits until "a real application has proven the API", and genx
    until "the sqlc generators run in a real app". Confab is now that app.
  - ui is "gated on neozp alone": its visual packages built without a
    breaking amendment.
  - handlers, i18n and models leave v1 to epistemer, with no criterion.

## 4. Confab as the product

### 4.1 What it proves

From [findings.md](findings.md), "What worked well":

- Typed links end to end. An action whose route the guard refuses
  disappears, with no second permission check.
- An sqlc enum can be a cfgx setting.
- Domain types validate themselves in forms and in route parameters.
- `As` chains typed values. Conditions run on both client and server from
  one `When`.
- One catalogue holds every text.
- The API's errors come out as `{code, message}` from the same declarations.

The "declarative, configurable, typed" promise holds wherever a package
owns the whole path.

### 4.2 What it costs

Confab has 3,891 hand-written lines of Go: the 5,237 non-test,
non-generated lines minus sqlc's 1,346. On top of that come 575 lines of
templ, 275 of SQL and an 80-line cfgx schema. The 12,834 generated lines
are free.

Of the hand-written Go, about 460 lines are workarounds and plumbing that
every app would copy:

| Lines | What | Why |
|---:|---|---|
| 120 | `scripts/stub` | F4: `handlers gen` can't generate a package that uses its own routes |
| 111 | `internal/pg` | F5: forms' `*T` against pgx's `pgtype` |
| 90 | `internal/confab/files.go` | file storage: ui's uploads store nothing (ui ADR 0012), and no package does |
| ~30 | the outbox in `internal/confab/people.go` | message delivery: auth's `Sender` has no implementation (auth §9) |
| ~85 | `flash.go`, `Recover`, `ErrorPage`, `Check` | F16, F25, F11: the request's lifecycle |
| 25 | `migrate.go` | migrations: nothing in neozp, and Confab executes `schema.sql` |
| — | vendored htmx, `confab.css` | F6, F23 |

Beyond those, `serve.go` is 194 lines of composition, and each entity's
admin file runs 420 to 525 lines. The save lifecycle is written three
times:

1. bind
2. answer 422 with the form
3. call the domain
4. map domain errors to fields
5. delete files stored for a save that failed
6. flash and redirect

Step 4 is a `fail` switch on every form because errors v1 leaves field
paths out (errors' interface §8). Step 5 exists because nothing ties a
stored file to a committed transaction.

### 4.3 Three layers no package owns

1. **A request's lifecycle.**
   - Panics drop the connection (F16).
   - A toast before a redirect panics after the save has committed (F25).
   - Flash messages are the app's to write (F25).
   - Forms may answer 400 but the script swaps only 422 (F17).
   - Nothing serves htmx, and its load order matters (F6).
   - `<html dir>` goes stale after a boosted page (F24).
   - Non-POST actions need script (F18).

   zp kept all of this in its own middleware and event code. In neozp it
   falls between handlers (which wraps handlers) and ui (which says what a
   response looks like).
2. **Values that cross boundaries.** An enum goes SQL, then sqlc, models,
   forms, handlers, cfgx and OpenAPI, and each step converts it its own way:
   - path parameters accept any string (F28);
   - pgx needs the type registered (F15);
   - cfgx writes a second set of labels (F3);
   - the English labels are raw SQL values (F19).

   Nulls have the same problem (F5), as do display labels for data (F8).
3. **The generation pipeline is not one acyclic graph.**
   - A handler package needs its generated routes to type-check, and the
     generator needs the package to type-check (F4).
   - templ runs outside genx, so `genx check` can't see a stale
     `_templ.go` (F31).
   - genx can't run one producer to break the loop (F32).

### 4.4 What Confab hasn't tested

- **auth had no findings because it was the least exercised**, not the best.
  - Only these ran: password plus emailed code, remembered devices, bearer
    tokens with refresh, and groups.
  - Not yet run: passkeys, OIDC, TOTP, step-up (`Within`), invitations,
    recovery and the `Recorder`. auth's events go nowhere in Confab.
  - CSRF is covered: `Load(Cookie())` applies `http.CrossOriginProtection`
    to every `/admin/` request.
- **Detail pages.** zp has 16, some with an embedded table. Confab has
  only a public event page.
- **A second population of users** with their own sign-in and their own
  records (in Confab, speakers).
- **Background work:** schedules, emails, retries.
- **ui's catalogue of stories** for an app's own components, an end-to-end
  browser test, and CI. Confab has no `.github` yet.

### 4.5 The findings, sorted by who acts

The test for each finding: does fixing it change what an app writes, or a
documented behaviour? If so, it is an amendment and the API decision is
epistemer's (every package's ownership.md, "What comes back to you").

**Fixes within each package's terms** (no API change):

| Package | Findings |
|---|---|
| ui | F23 (CSS), F24 (the script copies `lang` and `dir` from a boosted response), F27 (ui/inputs renders `dir="auto"` on text-like inputs; the rules' documented `Attrs` stay as they are), F6 (the combobox upgrades when htmx appears later), F17 (the skill says 422; server-state-interface already does), F10 and F26 (docs) |
| cfgx | F14 (the `init` skeleton calls `slog.SetDefault`) |
| models | F2 (document `table_prefix`'s trade-off) |
| genx | F31 (document templ's order before genx) |

**Amendments** (questions for epistemer, each with a default in §6.4):

- handlers: F4, F13, F16, F18, F21, F22.
- ui: F6, F7, F9, F11, F20, F25, F26, F30.
- models: F5, F8, F15, F19, F28.
- cfgx: F1, F3, F12.
- i18n: F29.
- genx: F32.

## 5. zp as the record of needs

### 5.1 What neozp already covers

| zp area | neozp |
|---|---|
| settings in bbolt, config | cfgx and its store |
| translations in bbolt, locale middleware | i18n |
| errors and their generator | errors |
| enum, view and meta generators (sqlc plugins) | models |
| the routes generator and router | handlers |
| sessions, invitations, password resets, one-time codes, groups, permissions | auth, auth/access |
| templ components, layout, icons, theme middleware | ui |
| validators: email, phone, post codes, country and currency codes | ui/forms/rule |
| body limit | handlers (`//h:body`, `x-max-body`) |
| the generator entry point | genx |

### 5.2 What no package covers

Each item below is confirmed in zp's code, except where it is only in a
neozp backlog (the shared store, the command palette). *Italics* mark what
Confab hand-rolled.

**Listed in a backlog, not scheduled:**

- **API keys for machine clients.** zp has key middleware and a command to
  issue keys. Listed in auth §9 and handlers' backlog item 3, each waiting
  on the other.
- **Resource-scoped permissions.** zp has a second population of users,
  with their own API and their own records. Listed in auth §9.
- **A store shared between replicas.** In cfgx's backlog and i18n's Later
  list.
- **A command palette for global search.** In ui's inventory coverage list;
  not built.

**Explicitly disowned, with nobody named:**

- **Message delivery.** zp sends SMTP email from six templ templates, SMS
  through providers, and push notifications. auth §9 leaves "delivering
  messages" out, and *Confab's outbox* is a stand-in.
- **File storage.** zp stores on local disk or S3, from a path, bytes, an
  upload or a URL. ui ADR 0012 stores nothing, and *Confab's `files.go`*
  plus its cleanup in every save fills the gap.
- **General HTTP rate limiting.** zp keeps its limiter in bbolt. auth's
  brief §6.8 calls it "another library's job".

**Mentioned nowhere:**

- **Jobs.** zp's scheduler is about 3,000 lines:
  - cron and ISO 8601 schedules;
  - executions stored in the database;
  - overlap strategies;
  - recovery of missed runs;
  - a reaper for stale tasks;
  - worker pools;
  - idempotency keys for side effects that must not run twice.

  neozp mentions a scheduler only as a group in cfgx's example config.
  *Confab runs `auth.Sweep` in a bare goroutine.*
- **HTTP plumbing.** zp has middleware for request ids, access logs,
  compression, security headers, health checks and static caching.
  *Confab has none of these.*
- **Migrations.** zp uses golang-migrate; *Confab executes `schema.sql`.*
- **An audit trail people can read.** zp records session activity, and
  auth emits `Event`s to a `Recorder`, but nothing stores or shows them.
- **Remembered table state per user.** zp keeps it in cookies; ui's tables
  keep their state in the URL only.

**Not counted, because the code doesn't confirm them:** exports to Excel
or CSV, charts, and PDF documents. zp has five templ PDF templates, but no
code outside their generated files calls them. zp's server-sent events
stream only table rows, which ui ADR 0010 already settled.

### 5.3 One stance to state once

auth, i18n and cfgx's store all keep their state in bbolt, so a neozp app
is a single process. zp also keeps its settings, translations and a rate
limiter in bbolt, which points to one process as well, so nothing is in
conflict today. But the stance is recorded three times, as cfgx
interface §7.3, cfgx's backlog and i18n's Later list. It should be one
neozp-wide decision.

## 6. Next steps

The packages are ahead of their first user. The next gains are in the
seams, not in more depth. Two rules for the next stretch:

- **A finding is done when Confab deletes its workaround.** A release that
  fixes it isn't enough on its own.
- **Integration runs continuously.** Confab's CI checks every package,
  rather than waiting for the next app.

### 6.1 Now: bookkeeping

- Correct the board.
  - The findings items say Confab isn't on GitHub. It is.
  - The milestone gives the wrong severity counts: it should say 1 high,
    15 medium and 16 low.
- Align errors' and auth's roadmaps with the 2026-11-16 rule, replacing the
  stale v1 lines with the gate decided in Q2.

### 6.2 Close iteration 1

1. Make the fixes in §4.5 that need no API change, one PR per repository,
   each with a patch release.
2. epistemer answers the amendment questions in §6.4. Unanswered ones take
   their defaults.
3. The packages release the amendments. Batch them into one minor release
   per package, not a release per finding.
4. Confab moves to the new versions and deletes the workarounds:
   - `scripts/stub`
   - `internal/pg`
   - `flash.go`
   - `Recover`
   - the vendored htmx
   - `confab.css`
   - the `postCode` recover
   - `dataLabel`
   - `H.Check`
   - the double permission grants

   Each deletion closes its finding.

### 6.3 Widen Confab, in three iterations

| Iteration | Scope | Done when |
|---|---|---|
| **2: consolidation** | §6.2's deletions; CI with a Postgres service; a nightly job that builds Confab against every package's `main`; one end-to-end browser test (sign in, create and publish an event, review a talk); a detail page with tabs and an embedded table (a speaker and their talks); ui catalogue stories for Confab's components | iteration 1's workarounds are gone, and CI is green on tags and on `main` |
| **3: people** | a speaker portal, where a second population signs in and edits only its own talks (resource scope); staff invited into groups; recovery; passkeys; OIDC against a local provider; TOTP; step-up before deleting an event (`auth.Within`); auth's `Recorder` behind an activity page; partner API keys | every auth flow runs in Confab, and the resource-scope and API-key questions have a real app asking |
| **4: background work** | a call for papers that closes at a date (a job); reminder emails from templates (delivery); speaker photos on a second storage backend (files) | each runs on a neozp package, or that package's brief cites this iteration's code as its need |

Iteration 3 is the "app asking" that auth §9 and handlers' backlog item 3
are waiting for, which breaks their loop.

### 6.4 Questions for epistemer

Each question has a default. They are mirrored on the board as Decision
items, one per package and one for neozp as a whole. No amendment starts
before epistemer has replied to its item; a reply that answers only some
questions takes the defaults for the rest.

**⚠ marks a default that reverses a locked or documented decision.** The
mark says which, so that accepting the default is a deliberate choice.

**neozp-wide**

- **Q1. A home for seams.** *Default:* Confab's repository is neozp's
  integration home. neozp-wide ADRs and an integration roadmap live in its
  `docs/`, and its CI is the cross-package check. (The alternative is a new
  `epistemer/neozp` repository.)
- **Q2. The v1 gate.** *Default:* a package's v1.0.0 waits until Confab uses
  its whole surface, with no open bug or contradiction finding against it.
  One criterion for all eight. It replaces the zp lines in errors, auth and
  genx, makes cfgx's and genx's "a real app" concrete, and joins ui's
  "without a breaking amendment".
- **Q3. Single instance.** *Default:* neozp is single-process until v1. That
  is stated once, and the shared-store items in cfgx and i18n stay parked.
- **Q4. Who owns the request's lifecycle.** *Default:*
  - handlers owns what wraps a handler, which is recovery (the open
    Decision).
  - ui owns what a response says: a flash that carries a toast across a
    redirect, and htmx served by `theme`.
- **Q5. Which new package first.** *Default:* message delivery (Confab can't
  sign anyone in for real without it), then files (Confab already
  hand-rolls it), then jobs (zp's largest need). Jobs gets its brief after
  iteration 4 has produced a real job. Each brief starts from Confab's
  code, with zp as corroboration.
- **Q6. HTTP plumbing.** *Default:* no package yet. Confab documents a
  standard-library recipe in iteration 2, and a package follows only if
  the recipe gets copied.

**handlers**

- **H1 (F4).** ⚠ Reverses the interface's error table: "the package doesn't type-check" is a documented refusal. *Default yes:* `handlers gen` reads signatures when the only
  type errors come from its own `-out` package, and reports them as
  warnings. This changes the documented refusal.
- **H2 (F16, the open Decision).** *Default:* `Register` and `Mount`
  recover a panic into an error passed to the app's error handler, and log
  it with its stack.
- **H3 (F22).** ⚠ Reverses interface's rule that "a permission belongs to one routes package", which §16 left for later. *Default:* permissions are generated once per app, a
  `permissions/v1` domain (backlog item 6, which now has its second user).
  The admin and the API can then share `talks.review`.
- **H4 (F21).** *Default yes:* a permission can carry its text:
  `//h:perm talks.review Review talks`, and the same in `x-permission`.
- **H5 (F13).** *Default yes:* a search `Hit` carries a `Link` when its
  route has no parameters.
- **H6 (F18).** *Default yes:* handlers accepts a form POST with `_method`
  for a route declared DELETE, PATCH or PUT, so ui's actions work without
  script, as tables promise.

**ui**

- **U1 (F25).** ⚠ Reverses server-state-interface's reply rules: `Go` after a toast panics by design, and the flash is the app's. *Default yes:* `Toast`, `Replace`, `Remove` and `Refresh`
  return a type without `Go`. A toast set before `Go` survives the redirect
  as a flash.
- **U2 (F6).** *Default yes:* `theme`'s head serves a pinned htmx and its
  config, unless the app opts out.
- **U3 (F7).** *Default yes:* `NavItem` consults the actions gate.
- **U4 (F11).** ⚠ Reverses amendment A7's `Column.Remote(search, labeler)`, which takes the labeler at declaration. *Default yes:* a table reaches its remote `Labeler` per
  request, through `Bind` or the context, so tables stay package-level
  declarations.
- **U5 (F9).** ⚠ Reverses interface §6: `PostCode` documents that it panics for an unknown country. *Default yes:* `PostCode` passes a country it doesn't know,
  and `rule.HasPostCode` says which countries it knows.
- **U6 (F26).** ⚠ Renames a member of `Kind`, which ownership.md reserves for epistemer. *Default yes:* `KindToggle` becomes `KindYesNo`, and
  `Hint` refuses it on a `Checkbox`.
- **U7 (F30).** ⚠ Changes `AssertLabelled`'s documented check of every control. *Default yes:* `AssertLabelled` skips hidden fields.
- **U8 (F20).** *Default yes:* add icons for edit, delete, external link,
  upload, download and calendar.

**models**

- **M1 (F5).** ⚠ Reverses interface's documented refusal of `emit_pointers_for_null_types` on `_view` relations (A4). *Default yes:* column sets accept sqlc's
  `emit_pointers_for_null_types`, so an app uses `*T` from the database to
  the form.
- **M2 (F28).** *Default yes:* enums get `UnmarshalText` and `MarshalText`,
  which refuse unknown values. handlers already prefers them.
- **M3 (F15).** *Default yes:* a generated `RegisterTypes(ctx, conn)` for
  enums and enum arrays.
- **M4 (F8).** *Default:* add `option.Static(value, text)` for data labels,
  rather than changing `SelectOption.Label()`.
- **M5 (F19).** ⚠ Reverses neozp's labels convention that an enum value's English text "is itself". *Default yes:* enum values take their English text from
  `COMMENT ON TYPE`, in the form `value: Text; …`.

**cfgx**

- **C1 (F12).** *Default yes:* `Problem.Err` keeps the validator's error.
- **C2 (F3).** ⚠ Reverses ADR 0008 (A14), under which cfgx labels every enum value. *Default yes:* no value labels for an enum whose Go type
  already has `Label()`.
- **C3 (F1).** *Default yes:* `cfgx init -name`.

**i18n**

- **I1 (F29).** *Default yes:* `ErrUnknownLocale`, `ErrUnknownKey` and
  `ErrFormat` implement `T(ctx)` and `Kind()`, so errors shows them to
  people without importing i18n.

**genx**

- **G1 (F32).** *Default yes:* `genx run <producer>…` runs the named
  producers and what they need.

### 6.5 Open, not yet a question

- **A resource declaration.** Confab writes the same save lifecycle three
  times, and zp has 13 index, 20 create, 21 edit and 16 detail pages. One
  declaration per resource could generate those pages,
  but it is too early to design one. Iteration 2 extracts helpers inside
  Confab first, and only what keeps repeating becomes a proposal.
- **Exports, charts and PDFs** wait for a real need. Neither zp's code nor
  Confab has one yet.

## 7. Risks to watch

- **Building faster than it is checked.** 93,000 lines in 16 days, with one
  app. Every amendment now touches three or four packages, and one
  iteration produced 24 API questions. The remedy is continuous integration
  (§6.3) and batched amendments (§6.2).
- **Skills are contracts too.** Four findings started in the skills: F4's
  passing mention, F10, F17 and F26. Agents read the skills, not the
  interface docs. Confab's CI should catch a skill that contradicts the
  code, for example through a story or test built from each skill's
  examples.
- **One file, three owners.** auth, i18n and cfgx share the app's bbolt
  file. bbolt locks the file, so two commands of one app, such as `seed`
  and `serve`, can't hold it at once. This is untested in Confab, and
  iteration 2 should check it.
