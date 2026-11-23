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
