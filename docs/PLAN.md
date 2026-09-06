---
PLAN: "refactor!: migrate github.com/webtyp -> webtyp.com + adopt view.NewCallerLister"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 17700638390423483376
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — `clinical_encounter`: WebTyp rename + new `view.New` API

The framework moved from `github.com/webtyp/*` to the vanity path
`webtyp.com/*`, and **every framework module is now published** under the new
path. `origin/main` of this repo is still entirely on `github.com/webtyp/*`.
Two jobs:

- **A. The mechanical rename** `github.com/webtyp/*` → `webtyp.com/*`.
- **B.** Adopt the new `webtyp.com/view` `view.New` signature.

The module import path **stays** `github.com/veltylabs/clinical_encounter`. The
Go package name stays `clinical_encounter`. Do **not** touch
`docs/PLAN_MODEL_MIGRATION.md` / `docs/PLAN_UI_INITIAL_VIEW.md` (unrelated
historical plans).

---

## A. Rename `github.com/webtyp` → `webtyp.com`

### A1. Go source (`*.go`, including any `tests/*.go`)

Replace import-path prefix **`github.com/webtyp/`** → **`webtyp.com/`** in
every `.go` file. Grep:

```
grep -rln 'github.com/webtyp' --include='*.go' .
```

Package selectors do not change — only the path in the `import` block.
`view.go` legitimately uses `webtyp.com/time` (`time.FormatCompact`,
`time.Weekday`) — that import **stays** (renamed, not removed); `webtyp.com/time`
is the framework clock, not stdlib.

### A2. `go.mod`

`origin/main` `require` block:

```
github.com/webtyp/ddl v0.0.7
github.com/webtyp/events v0.0.2
github.com/webtyp/fmt v0.25.5
github.com/webtyp/form v0.3.29
github.com/webtyp/input v0.0.3
github.com/webtyp/model v0.1.4
github.com/webtyp/orm v0.11.4
github.com/webtyp/router v0.1.19
github.com/webtyp/storage v0.0.2-0.20260717121821-7e528006807f
github.com/webtyp/time v0.5.0
github.com/webtyp/view v0.1.15
github.com/webtyp/dom v0.13.5 // indirect
github.com/webtyp/json v0.5.17 // indirect
github.com/webtyp/widget v0.6.6 // indirect
```

For **each** `github.com/webtyp/<X>` (including the `storage` pseudo-version):

```
go mod edit -droprequire=github.com/webtyp/<X>
go get webtyp.com/<X>@latest
```

Then `go mod tidy`. `@latest` is authoritative; current published tags for
reference: `ddl v0.0.15`, `events v0.0.3`, `fmt v1.0.0`, `form v0.4.7`,
`input v0.0.6`, `model v0.1.8`, `orm v0.12.1`, `router v0.1.31`,
`storage v0.0.7`, `time v0.5.5`, `view v0.5.2`, `dom v0.13.10`, `json v0.5.25`,
`widget v0.6.24`. No `github.com/webtyp/*` left in `go.mod`; no
`replace … => ../…` pointing outside this module.

### A3. Docs / config text

`*.md` / `*.yml` / `*.yaml`: `github.com/webtyp/` → `github.com/webtyp/`.
Prose `WebTyp`/`WebTyp` → `WebTyp`. Leave `LICENSE` and upstream "TinyGo"
references untouched.

---

## B. New `view.New` API — use `view.NewCallerLister`

### The change

`webtyp.com/view`'s `view.New` no longer takes a `router.Caller`, an op-name
string, a slice factory, or `view.WithSaveOp` (removed):

```go
func New(l Lister, record model.Model, opts ...Option) Presenter
type Lister interface{ List() ([]model.Model, error) }
```

The framework ships the adapter that reproduces the old op/caller behaviour
exactly:

```go
func NewCallerLister(c router.Caller, ops Ops, newList func() model.ModelSlice) Lister
type Ops struct{ List, Save, Update, Delete string }  // Ops.List required;
// returned Lister carries exactly the write caps whose op name is non-empty
```

### Reference implementations (read first)

- **`webtyp.com/auth` → `auth/view.go`** (canonical — a full List+Save+Delete
  view).
- **`github.com/veltylabs/business_hours`** `main` — the same migration on a
  sibling.

### Do — rewrite `view.go`

`NewView` keeps its **exact current signature**
(`func(caller router.Caller) view.Presenter`). The old body used only
`view.WithSaveOp(OpCreateVisit)` — **no delete** — so `view.Ops` gets `List` +
`Save` and **no `Delete`**. Imports `webtyp.com/model`, `webtyp.com/router`,
`webtyp.com/time`, `webtyp.com/view` all stay. Body:

```go
// NewView construye el Presenter del historial médico — el motor agnóstico de
// tecnología que envuelve un renderer (rightpanel, o cualquier otro). Este
// módulo lo construye (view + model + router); la app decide qué renderer lo
// dibuja.
func NewView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{List: OpListVisitsByPatient, Save: OpCreateVisit},
		func() model.ModelSlice { return &MedicalHistoryList{} })
	return view.New(b, &MedicalHistory{}, view.WithTitle(titleMedicalHistory))
}
```

`titleMedicalHistory` is a new unexported constant in this package
(`const titleMedicalHistory = "Historial clínico"`) — do not inline the literal.
`OpListVisitsByPatient` and `OpCreateVisit` already exist in `ops.go` — reuse
them. `leadFromUnix`, `weekdayAbbr`, `monthAbbr`, and `(*MedicalHistory).Item()`
stay exactly as they are.

### Tests

`NewView`'s signature is unchanged, so tests calling `NewView(fakeCaller)` keep
working. Adapt any test referencing a removed symbol (`view.WithSaveOp`, old
`view.New` arity) minimally, preserving the assertion intent (lists rows, is
Saver, is **not** Deleter).

---

## Verify

```
grep -rn 'github.com/webtyp' --include='*.go' --include='go.mod' .   # empty
grep -rn 'view.WithSaveOp\|view.WithDeleteOp' .                        # empty
grep -rn '=> \.\./' go.mod                                            # empty
```

## Acceptance

- `go build ./...` → clean.
- `gotest ./...` → all green (vet, race, tests).
- No `github.com/webtyp/*` anywhere in `*.go` / `go.mod`.
- `view.go` no longer references `view.WithSaveOp` / the 5-arg `view.New`.
- `NewView` still has signature `func(caller router.Caller) view.Presenter`.
- The Presenter is a `view.Saver` but **not** a `view.Deleter` (unchanged from
  before).

## Constraints

- **No behaviour change** — path rename + adapter swap only. `view.Ops.Save` is
  the exact string the old `view.WithSaveOp` carried.
- **No hardcoded strings** — the view title is a named constant; op names
  already are (`ops.go`).
- Keep every `//go:build` tag exactly as-is (backend + WASM shared).
- Do **not** add `Delete` (or `Update`) to `view.Ops` — the old view was
  List + Save only; adding either would make the Presenter advertise a
  capability the backend never had.
- `webtyp.com/time` stays imported in `view.go` — do not "remove stdlib": it is
  the framework clock, deliberately used.
