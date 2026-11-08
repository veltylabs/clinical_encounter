---
PLAN: "refactor!: clinical_encounter on the closed crud contract — server tenant, transactional batch, single declarative validation, seed.Demo"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 4384197113549986892
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work. Commit after every stage.
> Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.
> Never edit `model_orm.go` or `model_crud.go`: a build error there means the hand-written code is wrong.

# Plan: `clinical_encounter` on `crud.Registry`

Context: master `webtyp/docs/ARCHITECTURE_MASTER_PLAN.md` D2–D18 (phases A4.2 + L5). Follows the template in `patient_directory`.
After this plan the module mounts operations through generated helpers on `crud.Registry`,
tenant comes from `Deps`, validation of the schema lives in one place (`crud`), and `seed.Demo` provides the demo without `webtyp.com/layout`.

## Tools and local dependencies

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go install webtyp.com/ormc/cmd/ormc@v0.1.37
```
```bash
go get webtyp.com/model@v0.2.14 webtyp.com/router@v0.4.19 webtyp.com/crud@v0.0.3 webtyp.com/rpc@v0.2.17 webtyp.com/mcp@v0.3.23 \
  webtyp.com/dbsync@v0.3.47 webtyp.com/view@v0.6.67 webtyp.com/appkit@v0.2.28
```
Remove every `replace` from `go.mod`; `go mod tidy`; run `ormc` (generates `model_crud.go`).
After the delivery check, `go.mod` does not list `webtyp.com/layout`.

## APIs used

| Symbol | Signature |
|---|---|
| generated `Mount<Entity>Read/Create/Update/Delete` | typed signatures on `crud.Registry` |
| `crud.ReadArgs` | `struct{ ID, Search, After string; Limit int64; Filter model.Model }` |
| `crud.Caller` | `struct{ UserID string }` |
| `crud.Loopback` / `crud.LoopbackActingAs` | `func Loopback(mods ...crud.Module) router.Caller` / `func LoopbackActingAs(userID string, mods ...crud.Module) router.Caller` |
| `crud.Op` | `func Op(m model.ModuleNaming, action model.Action) string` |
| `model.ValidateFields` | `func ValidateFields(action byte, f Fielder) error` |
| `orm.DB.Tx` | `func (d *DB) Tx(fn func(tx *DB) error) error` |

## Explicit deletions (grep must be empty)

- `ops.go`: raw `reg.Operation` calls, `var _ router.OperationModule`, imports `webtyp.com/router`, `webtyp.com/rpc`.
  `grep -rn "reg.Operation\|router.Context\|rpc\." --include=*.go . | grep -v _test` → empty.
- `model.go`: manual argument models (`CreateVisitArgsModel`, `GetVisitArgsModel`, `ListVisitsArgsModel`, etc.) superseded by `crud.ReadArgs`; errors covered by schema/crud.
- `web/`: directory deleted (`test ! -d web`).

## Stage 1 — `model.go` and `errors.go`

- `tenant_id` and `updated_at` get `ServerOnly: true` on models (D11).
- Remove manual argument models that are covered by CRUD arguments.
- Verify `errors.go` has doc comments in Spanish with usage examples (`// Ejemplo:`).
- Install `ormc@v0.1.37` and run `ormc generate .`.

## Stage 2 — `crud.go` (writes) and `queries.go` (reads)

- `crud.go`: Create, Update, Delete methods for MedicalHistory, EncounterSupply wrapped in `m.db.Tx`.
  Method signatures accept `(c crud.Caller, list <Entity>List) error`.
- `queries.go`: Read methods accepting `(c crud.Caller, args crud.ReadArgs, filter model.Fielder) (model.ModelSlice, error)`.
  Filter always scopes by `tenant_id` from `m.tenantID`.

## Stage 3 — `ops.go` and `module.go`

- `ops.go`: mount operations via generated `Mount<Entity>*` helpers on `crud.Registry`.
- `Module` implements `crud.Module` (`model.ModuleNaming` + `MountOperations(reg crud.Registry)`).
  `ModelName() string { return "clinical_encounter" }`.

## Stage 4 — `view.go`, `seed/demo.go`, and remove `web/`

- `view.go`: update `view.Ops` to closed contract: `Allow: model.Read | model.Create | model.Update` (no string op names). Do NOT add `//go:build browser`.
- Note L6: `clinical_encounter` retains its custom UI components in `ui/` as per master plan L6.
- `seed/demo.go`: provide `Demo() (view.Demo, error)` with `view.Demo{Label: "Clinical Encounter", Timezone: "America/Santiago", ActingAs: "u1", Modules: []crud.Module{mod}, Screens: func(d view.BrowserDeps) []view.Screen { return ui.Screens(d) }}`.
- Delete `web/` directory (`rm -rf web`).

## Stage 5 — tests and delivery check

- Update tests to use `crud.Loopback(m)` or `crud.LoopbackActingAs(userID, m)` or direct methods.
- Run `gotest`.
- Run `langc sync .`.
- Verify explicit deletions grep is empty.
