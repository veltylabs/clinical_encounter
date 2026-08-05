---
PLAN: "feat: clinical_encounter joins the reusable-module harness (OpModule, IDGenerator, ddl, view.Presenter replaces web/, storage/mem tests)"
---

> This plan is dispatched via the CodeJob workflow. See skill: **agents-workflow**.

# PLAN — clinical_encounter joins the reusable-module harness

You are an agent with **zero prior context** and **only this repository**
(`github.com/veltylabs/clinical_encounter`). This plan is fully self-contained: every contract,
rule, and code snippet you need is inline below. Read `AGENTS.md` (this repo's root) first — it is
the canonical whitelist/blacklist of imports and is assumed as background for every stage below.

**Absolute rule, repeated because it matters more than anything else in this file:** there is a
sibling directory `../clinical_encounterOld`. It is legacy code with no importers anywhere in the
ecosystem, it is **not part of this repository**, and this plan never touches it. Do not read it, do
not migrate it, do not reference it beyond this warning.

---

## Stage 0 — Read this before touching any file: the view-reconciliation decision

This module currently has **two conflicting generations of UI code**, and this plan replaces the
older one. Read this stage in full before starting Stage 4 — it is the highest-risk part of this
plan and the reviewer will check it specifically.

### What exists in the repo right now

- `web/client.go` (`//go:build wasm`, `package main`) — a WASM entrypoint that imports
  `github.com/tinywasm/dom`, `github.com/tinywasm/layout/rightpanel`, and
  `github.com/tinywasm/unixid` **directly**, builds a `clinical_encounter.Module` by struct literal,
  and renders a `rightpanel.RightPanel` to `"body"`.
- `web/mockdb/mockdb.go` (`//go:build wasm`, `package mockdb`) — a hand-rolled `MemExecutor`/
  `MemCompiler` reimplementing `orm.Executor`/`orm.Compiler` in memory, used only so `web/client.go`
  has a non-nil `*orm.DB` to pass into `Module{}`.
- `view.go` (module root, **no build tag**, `package clinical_encounter`) — imports
  `github.com/tinywasm/dom` and `github.com/tinywasm/html` directly and declares four concrete UI
  components: `PatientHeadView`, `VisitFormView`, `HistoryListView`, `HistorySearchView`, each
  embedding `dom.Element` and implementing `Render() *dom.Element`.
- `viewCtrl.go` (`//go:build wasm`, `package clinical_encounter`) — a `ViewCtrl` that also imports
  `github.com/tinywasm/dom` directly, wiring click handlers between `HistoryListView` and
  `VisitFormView`.
- `ssr.go` (`//go:build !wasm`, `package clinical_encounter`) — `RenderCSS()`, a Go string literal
  of CSS classes (`.ce-patient-head`, `.ce-history-card`, `.ce-status-*`, …) consumed by the
  components above.

All five files were built under `docs/PLAN_UI_INITIAL_VIEW.md`, whose own "Estado de ejecución"
checklist marked **5 of 7 steps done** — including `[x] Verificar compilación (GOOS=js GOARCH=wasm go
build -o /tmp/app.wasm ./web/)`, i.e. a **verified, working WASM build**. Only "verify render in a
real browser" and "integrate `tinywasm/form`" were left open.

### Why this cannot stay as-is

Every one of the five files above imports a **concrete renderer** (`tinywasm/dom`, `tinywasm/html`,
`tinywasm/layout/rightpanel`) directly from module code with no build-tag exemption that would make
it acceptable — `AGENTS.md`'s blacklist is explicit: *"A concrete renderer: `tinywasm/layout` (or any
other UI kit). A module builds its `view.Presenter` with `view.New(caller, ...)` —
`view`+`model`+`router` only. The app picks the renderer that draws that `Presenter`."* `view.go` and
`viewCtrl.go` are not exempt just because they live at the module root instead of under `web/` — the
rule is about what gets imported, not which directory the file sits in.

Separately, `web/mockdb` reimplements `orm.Executor`/`orm.Compiler` as an in-memory store. That is
now **redundant**: `github.com/tinywasm/storage/mem` is exactly this, upstream, maintained, and
mandated by `AGENTS.md`'s Testing section for every module's own tests (`orm.New(mem.New())`).

### The decision — stated plainly, this discards verified work

**`web/` is deleted in full (both files). `view.go`, `viewCtrl.go`, and `ssr.go` are deleted in full
and `view.go` is rewritten from scratch as `NewView(caller router.Caller) view.Presenter`**, built
only from `view`+`model`+`router`, matching `veltylabs/item_catalog/view.go`. The APP — not this
module — will choose which concrete renderer (`tinywasm/layout/rightpanel` again, or anything else)
draws the resulting `Presenter`.

**This is not a mechanical rename.** Someone did real, in-browser-verified UI work in this repo — a
WASM binary that actually compiled and (per the checklist) was one step from being confirmed
rendering in a real browser. This plan asks you to delete that code and rebuild the equivalent
capability (list + select + save a visit) behind a different, narrower contract (`view.Presenter`)
that the module can unit-test itself without a browser, but that a browser-rendered UI still has to
be re-wired on top of, in the app repo, later. The `RenderCSS()` styling in `ssr.go` and the exact
`rightpanel.RightPanel` slot wiring in `web/client.go` are useful references for whoever does that
app-side wiring later — but they do not belong in this module, so they are deleted here, not moved.
If you want a record of what is being discarded, `git log` on this repo (before this plan's commit)
has it; this plan does not preserve a copy inside the module.

---

## Stage 1 — `model.go`: plain structs → `model.Definition`

### Current `model.go` (verbatim, to be replaced)

```go
package clinical_encounter

type MedicalHistory struct {
	ID                      string
	PatientID               string `db:"not_null"`
	DoctorID                string `db:"not_null"`
	ReservationID           string // soft ref — optional
	Status                  string `db:"not_null"` // see Status* constants
	AttentionAt             int64  `db:"not_null"` // Unix timestamp (date + time unified)
	Reason                  string `db:"not_null"`
	Diagnostic              string
	Prescription            string
	Cie10Code               string
	StartedAt               int64
	FinishedAt              int64
	PatientNameSnapshot     string `db:"not_null"`
	PatientRutSnapshot      string `db:"not_null"`
	DoctorNameSnapshot      string `db:"not_null"`
	DoctorSpecialtySnapshot string
	UpdatedAt               int64 `db:"not_null"`
}

type CreateVisitArgs struct {
	PatientID               string
	DoctorID                string
	AttentionAt             int64
	Reason                  string
	PatientNameSnapshot     string
	PatientRutSnapshot      string
	DoctorNameSnapshot      string
	ReservationID           string
	Diagnostic              string
	Prescription            string
	DoctorSpecialtySnapshot string
}
```

### The `model.Definition` contract (summary — full contract in `AGENTS.md`)

`Field.Type` is the `model.Kind` interface, filled by a constructor (`model.Text()`, `model.Int()`,
…) — never a bare enum literal. A field that needs a UI widget uses a `Kind` from
`github.com/tinywasm/form/input` instead (`input.Text()`, `input.Textarea()`, `input.Number()`) —
those also satisfy `model.Kind`. Do not re-derive this contract from first principles; if something
here is unclear, `AGENTS.md`'s whitelist section and
[`veltylabs/item_catalog/model.go`](https://github.com/veltylabs/item_catalog/blob/main/model.go) are
the concrete reference.

### New `model.go` (full replacement)

`MedicalHistoryModel` keeps plain `model.Text()`/`model.Int()` Kinds — it has no widget today (it is
the historical record: shown, never edited through a form directly) and this migration does not give
it one. `CreateVisitArgsModel` gains a widget on **every** field, proactively: `view.go`'s old
placeholder comment already said *"In production, use tinywasm/form.New(parentID,
CreateVisitArgs{})"* — without a widget, that `form.New` call renders empty the day something wires
it up. `diagnostic`/`prescription` use `input.Textarea()` (long clinical text); everything else uses
`input.Text()`/`input.Number()`.

Two new **transport-only** `Definition`s are added (`Field.DB` is `nil` throughout — they persist
nothing) to carry the args of the two read ops introduced in Stage 3.

```go
package clinical_encounter

import (
	"github.com/tinywasm/form/input"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/model"
)

var MedicalHistoryModel = model.Definition{
	Name: "medical_history",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "patient_id", Type: model.Text(), NotNull: true},
		{Name: "doctor_id", Type: model.Text(), NotNull: true},
		{Name: "reservation_id", Type: model.Text()},
		// status: valid values are ONLY the Status* constants in const.go — the literals live
		// in those constants and nowhere else (anti magic-string rule; see item_catalog review).
		{Name: "status", Type: model.Text(), NotNull: true},
		{Name: "attention_at", Type: model.Int(), NotNull: true},
		{Name: "reason", Type: model.Text(), NotNull: true},
		{Name: "diagnostic", Type: model.Text()},
		{Name: "prescription", Type: model.Text()},
		{Name: "cie10_code", Type: model.Text()},
		{Name: "started_at", Type: model.Int()},
		{Name: "finished_at", Type: model.Int()},
		{Name: "patient_name_snapshot", Type: model.Text(), NotNull: true},
		{Name: "patient_rut_snapshot", Type: model.Text(), NotNull: true},
		{Name: "doctor_name_snapshot", Type: model.Text(), NotNull: true},
		{Name: "doctor_specialty_snapshot", Type: model.Text()},
		{Name: "updated_at", Type: model.Int(), NotNull: true},
	},
}

// NotNull mirrors exactly the required-argument set CreateVisit enforces (visit.go) — the
// Definition is the single place the contract is declared; the service's manual checks are a
// defense-in-depth duplicate of the same set, never a different one.
var CreateVisitArgsModel = model.Definition{
	Name: "create_visit_args",
	Fields: model.Fields{
		{Name: "patient_id", Type: input.Text(), NotNull: true},
		{Name: "doctor_id", Type: input.Text(), NotNull: true},
		{Name: "attention_at", Type: input.Number(), NotNull: true},
		{Name: "reason", Type: input.Text(), NotNull: true},
		{Name: "patient_name_snapshot", Type: input.Text(), NotNull: true},
		{Name: "patient_rut_snapshot", Type: input.Text(), NotNull: true},
		{Name: "doctor_name_snapshot", Type: input.Text(), NotNull: true},
		{Name: "reservation_id", Type: input.Text()},
		{Name: "diagnostic", Type: input.Textarea()},
		{Name: "prescription", Type: input.Textarea()},
		{Name: "doctor_specialty_snapshot", Type: input.Text()},
	},
}

// GetVisitArgsModel and ListVisitsArgsModel are transport-only (Field.DB is nil throughout) —
// see Stage 3 for the ops that consume them.
var GetVisitArgsModel = model.Definition{
	Name: "get_visit_args",
	Fields: model.Fields{
		{Name: "id", Type: model.Text()},
	},
}

var ListVisitsArgsModel = model.Definition{
	Name: "list_visits_args",
	Fields: model.Fields{
		{Name: "patient_id", Type: model.Text()},
	},
}

var (
	ErrNotFound    = fmt.Err("visit not found")
	ErrMissingArgs = fmt.Err("missing required arguments")
)

// Domain event topics — <module>.<entity>.<past-tense-verb>, tenant/id data goes in the
// payload, never in the topic name (n/a here — see "no TenantId field" note in ARCHITECTURE.md).
const (
	TopicVisitCreated = "clinical_encounter.visit.created"
)
```

Do **not** add a `TenantId` field to either `Definition` — the current schema has none, and this
migration ports existing columns mechanically (same table, same columns, same behavior). Adding
tenant scoping is a separate, explicit future change, not a side effect of this one.

Do **not** touch `const.go` — `StatusCreated`/`StatusArrived`/`StatusTriaged`/`StatusInProgress`/
`StatusCompleted`/`StatusCancelled` stay exactly as they are; they are referenced by value from
`visit.go`, not migrated into the `Definition`.

### `go get` before regenerating

```
go get github.com/tinywasm/model@v0.1.2 github.com/tinywasm/orm@v0.11.4 github.com/tinywasm/form@v0.3.13 github.com/tinywasm/fmt@v0.25.5
```

### Regenerate `model_orm.go`

```
go install github.com/tinywasm/ormc/cmd/ormc@latest
ormc   # run from the module root — regenerates model_orm.go, DO NOT hand-edit it
```

**Casing warning:** the regenerated struct uses **pure casing** — `MedicalHistory.ID` becomes
`MedicalHistory.Id`, and every `..._id` field becomes `...Id` (never `...ID`). This is a breaking
rename inside this module. Update every consumer:

- `visit.go`: `record.ID` → `record.Id`, `args.PatientID` → `args.PatientId`, `args.DoctorID` →
  `args.DoctorId`, `args.ReservationID` → `args.ReservationId`.
- Stage 3's new `ops.go` and Stage 4's new `view.go`: write them with the new casing directly (they
  do not exist yet at this point in the plan).

The DB columns and wire field names (`"patient_id"`, `"doctor_id"`, …) do **not** change — only the Go
struct field casing does.

### `model_test.go` — widget-regression test (required)

Create `tests/model_test.go` (package `tests`, external — see Stage 6 for why tests live there):

```go
package tests

import (
	"testing"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"github.com/tinywasm/form"
	"github.com/tinywasm/model"
)

var _ model.Model = (*clinicalencounter.CreateVisitArgs)(nil)

// TestCreateVisitArgsHasWidgets guards against CreateVisitArgsModel silently losing its
// widgets again — form.New must never render an empty form for it.
//
// API note (`tinywasm/form@v0.3.13`, verified against the real published module — this plan
// originally targeted the now-stale `v0.2.16`, where `form.New` returned a single `*Form`):
// `form.New` now returns `(*Form, error)`.
func TestCreateVisitArgsHasWidgets(t *testing.T) {
	f, err := form.New("clinical_encounter", &clinicalencounter.CreateVisitArgs{})
	if err != nil {
		t.Fatalf("form.New: %v", err)
	}
	if len(f.Inputs) == 0 {
		t.Fatal("CreateVisitArgsModel has no widgets — form.New returned an empty form")
	}
}
```

### Stage 1 acceptance criteria

- `grep -n "db:\"not_null\"" model.go` → empty (no struct tags left in `model.go`).
- `grep -n "type MedicalHistory struct" model.go` → empty (no hand-written struct — only
  `model.Definition` literals; the concrete struct lives in the regenerated `model_orm.go`).
- `grep -n "MedicalHistoryModel\|CreateVisitArgsModel\|GetVisitArgsModel\|ListVisitsArgsModel" model.go`
  → 4 matches.
- `grep -rn "\.ID\b" --include="*.go" . | grep -v _test.go` → empty (casing migrated everywhere).
- `go build ./...` succeeds.
- `gotest ./...` green, including the new `tests/model_test.go`.

---

## Stage 2 — Identity and events: retire `unixid.UnixID` and the local `EventPublisher`

### Current `module.go` (verbatim)

```go
package clinical_encounter

import (
	"github.com/tinywasm/orm"
	"github.com/tinywasm/unixid"
)

type Module struct {
	DB  *orm.DB
	UID *unixid.UnixID
	Pub EventPublisher
}

func (m *Module) ModelName() string {
	return "clinical_encounter"
}
```

### Current `publish.go` (verbatim, delete this file)

```go
package clinical_encounter

// EventPublisher is implemented by the host application (SSE, websocket, etc.).
// Pass nil to disable event publishing (no-op).
type EventPublisher interface {
	Publish(event string, payload any) error
}
```

`EventPublisher` is a **self-declared port that duplicates an ecosystem contract** —
`AGENTS.md`'s blacklist forbids exactly this: *"no local `EventPublisher`, `UIAdapter`, `IDGenerator`
… interface that intersects `events.Publisher`/… /`model.IDGenerator`/`router.OpModule`."* Delete
`publish.go` entirely; nothing replaces it in this file.

### Target `module.go` for this stage (ddl wiring comes in Stage 5 — this is the intermediate step)

```go
package clinical_encounter

import (
	"github.com/tinywasm/events"
	"github.com/tinywasm/model"
	"github.com/tinywasm/orm"
)

// Deps are the module's infrastructure ports — never a concrete implementation.
type Deps struct {
	IDs       model.IDGenerator // required — the module never builds its own
	Publisher events.Publisher  // optional — nil disables publishing silently
}

type Module struct {
	db  *orm.DB
	ids model.IDGenerator
	pub events.Publisher
}

func (m *Module) ModelName() string {
	return "clinical_encounter"
}
```

`New(db *orm.DB, deps Deps) (*Module, error)` is added in Stage 5 together with the `ddl` schema
migration — do not add it yet in this stage if you are executing stages strictly in order; if you are
applying this plan in one pass, Stage 5's `New()` supersedes the snippet above.

### Stage 2 acceptance criteria

- `grep -rn "tinywasm/unixid" .` → empty, repo-wide, tests included.
- `grep -rn "EventPublisher" .` → empty.
- `test -f publish.go` → does not exist.
- `grep -n "Deps struct" module.go` → 1 match, with `IDs model.IDGenerator` and
  `Publisher events.Publisher` fields.

---

## Stage 3 — Transport: author the first `router.OpModule`

### What exists today: nothing

There is **no existing transport layer** in this module — no `mcp.go`, no `ops.go`, no
`router.OpModule` implementation, not even a stub. `CreateVisit` in `visit.go` is a plain Go method
called only from `web/client.go` (which this plan deletes in Stage 4) via a direct struct literal.
This stage is **not a migration** — it is authoring the module's first transport surface from
scratch, so it is scoped tightly to what the current business logic actually supports.

### What NOT to invent

`const.go`'s FSM declares six `Status*` values, but **no method in this repo transitions a record
between them** other than `CreateVisit` setting `StatusCreated` on creation. There is no
`Event*` constant group and no transition table in the code (a prior doc referenced one; it does not
exist). Do **not** add ops like `arrive_visit`/`triage_visit`/`start_visit`/`complete_visit`/
`cancel_visit` — those would require designing new business logic this plan does not specify. Only
these three ops are added, each backed by a real (or, for the two reads, newly-written but trivial)
`*Module` method:

| Op | Action | Resource | Backing method |
|---|---|---|---|
| `create_visit` | `model.Create` | `medical_history` | `Module.CreateVisit` (existing, adjusted for new `Deps`) |
| `get_medical_history` | `model.Read` | `medical_history` | `Module.GetVisit` (new, trivial — single `Query`) |
| `list_medical_history` | `model.Read` | `medical_history` | `Module.ListVisitsByPatient` (new, trivial — single `Query`) |

### `visit.go` — add `GetVisit`/`ListVisitsByPatient`, adjust `CreateVisit`

Current `CreateVisit` (verbatim, for reference — casing/field-source lines change per Stage 1/2):

```go
func (m *Module) CreateVisit(args CreateVisitArgs) (*MedicalHistory, error) {
	if args.PatientID == "" || args.DoctorID == "" || args.Reason == "" ||
		args.PatientNameSnapshot == "" || args.PatientRutSnapshot == "" || args.DoctorNameSnapshot == "" {
		return nil, fmt.Err("missing", "required", "arguments")
	}
	if args.AttentionAt == 0 {
		return nil, fmt.Err("missing", "attention_at")
	}
	record := &MedicalHistory{
		ID:                      m.UID.NewID(),
		PatientID:               args.PatientID,
		DoctorID:                args.DoctorID,
		ReservationID:           args.ReservationID,
		Status:                  StatusCreated,
		AttentionAt:             args.AttentionAt,
		Reason:                  args.Reason,
		Diagnostic:              args.Diagnostic,
		Prescription:            args.Prescription,
		PatientNameSnapshot:     args.PatientNameSnapshot,
		PatientRutSnapshot:      args.PatientRutSnapshot,
		DoctorNameSnapshot:      args.DoctorNameSnapshot,
		DoctorSpecialtySnapshot: args.DoctorSpecialtySnapshot,
		UpdatedAt:               time.Now(),
	}
	if err := m.DB.Create(record); err != nil {
		return nil, err
	}
	return record, nil
}
```

New `visit.go` (full replacement):

```go
package clinical_encounter

import (
	"github.com/tinywasm/events"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/time"
)

func (m *Module) CreateVisit(args CreateVisitArgs) (*MedicalHistory, error) {
	if args.PatientId == "" || args.DoctorId == "" || args.Reason == "" ||
		args.PatientNameSnapshot == "" || args.PatientRutSnapshot == "" || args.DoctorNameSnapshot == "" {
		return nil, ErrMissingArgs
	}
	if args.AttentionAt == 0 {
		return nil, ErrMissingArgs
	}

	record := &MedicalHistory{
		Id:                      m.ids.NewID(),
		PatientId:               args.PatientId,
		DoctorId:                args.DoctorId,
		ReservationId:           args.ReservationId,
		Status:                  StatusCreated,
		AttentionAt:             args.AttentionAt,
		Reason:                  args.Reason,
		Diagnostic:              args.Diagnostic,
		Prescription:            args.Prescription,
		PatientNameSnapshot:     args.PatientNameSnapshot,
		PatientRutSnapshot:      args.PatientRutSnapshot,
		DoctorNameSnapshot:      args.DoctorNameSnapshot,
		DoctorSpecialtySnapshot: args.DoctorSpecialtySnapshot,
		UpdatedAt:               time.Now(),
	}

	if err := m.db.Create(record); err != nil {
		return nil, err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicVisitCreated, Payload: record})
	}
	return record, nil
}

func (m *Module) GetVisit(id string) (*MedicalHistory, error) {
	var rec MedicalHistory
	qb := m.db.Query(&rec).Where(MedicalHistory_.Id).Eq(id)
	_, err := ReadOneMedicalHistory(qb, &rec)
	if err != nil {
		if err == orm.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &rec, nil
}

func (m *Module) ListVisitsByPatient(patientId string) ([]*MedicalHistory, error) {
	var rec MedicalHistory
	qb := m.db.Query(&rec).
		Where(MedicalHistory_.PatientId).Eq(patientId).
		OrderBy(MedicalHistory_.AttentionAt).Desc()
	results, err := ReadAllMedicalHistory(qb)
	if err != nil {
		return nil, err
	}
	return results, nil // ReadAllX returns []*MedicalHistory directly (same as item_catalog) — never dereference it
}
```

`fmt.Err("missing", "required", "arguments")` is replaced by the typed `ErrMissingArgs` from
`model.go` (Stage 1) — no hardcoded error strings duplicated across files. `MedicalHistory_` is the
typed-column helper `ormc` generates automatically for every DB-backed model (always-on, no directive
needed) — do not hand-write it.

### New `ops.go`

```go
package clinical_encounter

import (
	"github.com/tinywasm/model"
	"github.com/tinywasm/router"
)

const (
	OpCreateVisit          = "create_visit"
	OpGetVisit             = "get_medical_history"
	OpListVisitsByPatient  = "list_medical_history"
)

func (m *Module) MountOps(reg router.OpRegistry) {
	reg.Op(OpCreateVisit, m.opCreateVisit).Requires("medical_history", model.Create).Accepts(&CreateVisitArgs{})
	reg.Op(OpGetVisit, m.opGetVisit).Requires("medical_history", model.Read).Accepts(&GetVisitArgs{})
	reg.Op(OpListVisitsByPatient, m.opListVisits).Requires("medical_history", model.Read).Accepts(&ListVisitsArgs{})
}

var _ router.OpModule = (*Module)(nil)

func (m *Module) opCreateVisit(ctx router.Context) {
	var args CreateVisitArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	record, err := m.CreateVisit(args)
	if err != nil {
		// Status convention (ecosystem-wide): 400 = invalid input, 404 = not found,
		// 500 = genuine internal error only — never collapse client errors into 500.
		if err == ErrMissingArgs {
			ctx.WriteStatus(400)
			return
		}
		ctx.WriteStatus(500)
		return
	}
	if err := ctx.Encode(record); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opGetVisit(ctx router.Context) {
	var args GetVisitArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	record, err := m.GetVisit(args.Id)
	if err != nil {
		if err == ErrNotFound {
			ctx.WriteStatus(404)
			return
		}
		ctx.WriteStatus(500)
		return
	}
	if err := ctx.Encode(record); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opListVisits(ctx router.Context) {
	var args ListVisitsArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	records, err := m.ListVisitsByPatient(args.PatientId)
	if err != nil {
		ctx.WriteStatus(500)
		return
	}
	list := make(MedicalHistoryList, len(records))
	for i, r := range records {
		list[i] = r
	}
	if err := ctx.Encode(&list); err != nil {
		ctx.WriteStatus(500)
	}
}
```

No build tag on `ops.go` — `router`/`model` are isomorphic (README's Build Tags Rule).

### Stage 3 acceptance criteria

- `test -f ops.go` → exists.
- `grep -n "var _ router.OpModule" ops.go` → 1 match.
- `grep -n "func (m \*Module) MountOps" ops.go` → 1 match, taking `router.OpRegistry`.
- `grep -rn "tinywasm/mcp" .` → empty (no concrete transport ever imported).
- `gotest ./...` green.

---

## Stage 4 — Replace the concrete-renderer UI with `view.Presenter`

Executes the Stage 0 decision. Delete, then rewrite:

```
rm -rf web/
rm viewCtrl.go
rm ssr.go
```

(`viewCtrl.go` and `ssr.go` are deleted outright — no replacement file. `view.go` is rewritten, not
deleted, since the module still needs a `view.go`.)

### New `view.go` (full replacement — old content shown in Stage 0)

**API note (`tinywasm/view@v0.1.12`, verified against the real published module — this plan
originally targeted the now-stale `v0.1.1`):** `view.New` dropped its 4th positional arg (the
`func(list model.FielderSlice) []view.Item` projector) and `view.WithFill`. `newList` is now
`func() model.ModelSlice` (not `FielderSlice`), and list-row projection is a method the record
itself implements — `view.Itemizer` (`Item() view.Item`). Selection lookup (formerly the manual
`byID` slice + `WithFill`) is now automatic: the `Presenter` builds its own id→record index from
`Itemizer` during `Reload`. This makes the "no Go map" `byID` workaround unnecessary — delete it,
not port it.

```go
package clinical_encounter

import (
	"github.com/tinywasm/model"
	"github.com/tinywasm/router"
	"github.com/tinywasm/view"
)

// Item implements view.Itemizer — the ONLY view-specific code this record carries. The
// Presenter indexes rows by ID from this during Reload; there is no manual byID/WithFill lookup.
func (it *MedicalHistory) Item() view.Item {
	return view.Item{ID: it.Id, Label: it.Reason, Description: it.Status}
}

// NewView builds the medical-history Presenter — the tech-agnostic engine a renderer
// (tinywasm/layout/rightpanel, or any other) wraps. This module builds it (view+model+router
// only); the app decides which renderer draws it.
func NewView(caller router.Caller) view.Presenter {
	record := &MedicalHistory{}

	return view.New(
		caller,
		record,
		OpListVisitsByPatient,
		func() model.ModelSlice { return &MedicalHistoryList{} },
		view.WithTitle("Historial clínico"),
		view.WithSaveOp(OpCreateVisit),
	)
}
```

No `view.WithDeleteOp(...)` — there is no delete op (Stage 3 deliberately does not add one; a
clinical record is not deleted through this module). This mirrors
[`veltylabs/item_catalog/view.go`](https://github.com/veltylabs/item_catalog/blob/main/view.go)
exactly in shape, adjusted to this module's fields and ops.

### Stage 4 acceptance criteria

- `test -d web` → does not exist.
- `test -f viewCtrl.go` → does not exist.
- `test -f ssr.go` → does not exist.
- `grep -rn --include=*.go "tinywasm/dom\|tinywasm/html\|tinywasm/layout" .` → empty, repo-wide,
  tests included. Scoped to `.go` files deliberately: `tinywasm/dom` legitimately shows up in
  `go.mod`/`go.sum` as an **indirect** entry once Stage 1's `tests/model_test.go` (`tinywasm/form`)
  is in place — see the note in Stage 7. `tinywasm/html`/`tinywasm/layout` must not appear anywhere,
  including `go.mod`.
- `grep -n "func NewView(caller router.Caller) view.Presenter" view.go` → 1 match.
- `grep -n "view.New(" view.go` → 1 match.
- `go build ./...` succeeds.

---

## Stage 5 — `New()`: schema migration via `ddl.CreateTable`

There is currently **no `New()` constructor at all** — `web/client.go` (deleted in Stage 4) built a
`Module{}` by struct literal, so no code path ever created the `medical_history` table (the old
`web/mockdb` executor's `Exec` was a no-op anyway). This stage adds the constructor every other
module in this harness has.

### Final `module.go` (supersedes Stage 2's intermediate version)

```go
package clinical_encounter

import (
	"github.com/tinywasm/ddl"
	"github.com/tinywasm/events"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/model"
	"github.com/tinywasm/orm"
)

// Deps are the module's infrastructure ports — never a concrete implementation.
type Deps struct {
	IDs       model.IDGenerator // required — the module never builds its own
	Publisher events.Publisher  // optional — nil disables publishing silently
}

type Module struct {
	db  *orm.DB
	ids model.IDGenerator
	pub events.Publisher
}

func New(db *orm.DB, deps Deps) (*Module, error) {
	if deps.IDs == nil {
		return nil, fmt.Err("clinical_encounter: Deps.IDs is required")
	}
	// ddl.Compiler is an optional capability — only SQL backends (sqlt, postgres) implement it.
	// storage/mem (this module's own tests, Stage 6) creates tables lazily and needs no DDL, so a
	// type assertion — not an unconditional call — is how the module stays backend-agnostic here.
	if ddlCompiler, ok := db.RawConn().(ddl.Compiler); ok {
		if err := ddl.New(db.RawConn(), ddlCompiler).CreateTable(&MedicalHistory{}); err != nil {
			return nil, err
		}
	}
	return &Module{db: db, ids: deps.IDs, pub: deps.Publisher}, nil
}

func (m *Module) ModelName() string {
	return "clinical_encounter"
}
```

### Stage 5 acceptance criteria

- `grep -n "func New(db \*orm.DB, deps Deps) (\*Module, error)" module.go` → 1 match.
- `grep -n "ddl.Compiler\|ddl.New\|ddl.CreateTable" module.go` → present.
- `grep -n "db.RawConn()" module.go` → present, guarded by a type assertion (`if ddlCompiler, ok :=
  db.RawConn().(ddl.Compiler); ok {`), never called unconditionally.
- `grep -rn "orm.DB.CreateTable\|db.CreateTable" .` → empty (the removed method is never called).
- `gotest ./...` green.

---

## Stage 6 — Tests: `tests/` package, `orm.New(mem.New())`, no concrete driver

The `tests/` directory currently exists but is empty. Populate it. Never import a concrete storage
driver (`tinywasm/sqlite`, `tinywasm/sqlt`, …) — not even here. Follow `AGENTS.md`'s Testing section
and `item_catalog@main`'s `tests/` (already migrated to the correct pattern):
`orm.New(mem.New())` from `github.com/tinywasm/storage/mem`, `package tests` inside this same
module (no nested `tests/go.mod`).

### `tests/setup_test.go`

```go
package tests

import (
	"testing"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"github.com/tinywasm/events"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/model"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/storage/mem"
)

type mockIDGen struct{ counter int }

func (g *mockIDGen) NewID() string {
	g.counter++
	return "test-id-" + fmt.Convert(g.counter).String() // tinywasm/fmt — stdlib strconv is banned, tests included
}

var _ model.IDGenerator = (*mockIDGen)(nil)

type mockPublisher struct{ Events []events.Event }

// events.Publisher.Publish is fire-and-forget: NO error return (verified against
// github.com/tinywasm/events@v0.0.2 — a `Publish(e Event) error` signature does not compile).
func (p *mockPublisher) Publish(e events.Event) {
	p.Events = append(p.Events, e)
}

var _ events.Publisher = (*mockPublisher)(nil)

func setup(t *testing.T) *clinicalencounter.Module {
	t.Helper()
	db := orm.New(mem.New())
	m, err := clinicalencounter.New(db, clinicalencounter.Deps{IDs: &mockIDGen{}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return m
}
```

### `tests/visit_test.go` — service-method coverage

```go
package tests

import (
	"testing"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
)

func TestCreateVisit_HappyPath(t *testing.T) {
	m := setup(t)
	rec, err := m.CreateVisit(clinicalencounter.CreateVisitArgs{
		PatientId:           "pat_1",
		DoctorId:            "doc_1",
		Reason:              "Control rutinario",
		AttentionAt:         1700000000,
		PatientNameSnapshot: "Juan Pérez",
		PatientRutSnapshot:  "1-9",
		DoctorNameSnapshot:  "Dr. Soto",
	})
	if err != nil {
		t.Fatalf("CreateVisit: %v", err)
	}
	if rec.Id == "" {
		t.Fatal("expected a non-empty Id")
	}
	if rec.Status != clinicalencounter.StatusCreated {
		t.Fatalf("expected Status %q, got %q", clinicalencounter.StatusCreated, rec.Status)
	}
}

func TestCreateVisit_MissingRequiredArgs(t *testing.T) {
	m := setup(t)
	_, err := m.CreateVisit(clinicalencounter.CreateVisitArgs{})
	if err != clinicalencounter.ErrMissingArgs {
		t.Fatalf("expected ErrMissingArgs, got %v", err)
	}
}

func TestGetVisit_NotFound(t *testing.T) {
	m := setup(t)
	_, err := m.GetVisit("does-not-exist")
	if err != clinicalencounter.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListVisitsByPatient(t *testing.T) {
	m := setup(t)
	_, err := m.CreateVisit(clinicalencounter.CreateVisitArgs{
		PatientId: "pat_1", DoctorId: "doc_1", Reason: "r", AttentionAt: 1,
		PatientNameSnapshot: "P", PatientRutSnapshot: "1-9", DoctorNameSnapshot: "D",
	})
	if err != nil {
		t.Fatalf("CreateVisit: %v", err)
	}
	list, err := m.ListVisitsByPatient("pat_1")
	if err != nil {
		t.Fatalf("ListVisitsByPatient: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 record, got %d", len(list))
	}
}
```

### `tests/ops_test.go` — `MountOps` against `router/mock`

`router/mock.Router.Op(name, handler)` registers the op under the synthetic method `"OP"` and path
`"/"+name` — drive it with `Router.Invoke("OP", "/"+opName, ctx)`. `router/mock.Context.Decode`
backs onto a real JSON codec internally (that is `router/mock`'s own implementation detail, not this
module importing an encoder) — so `ctx.InBody` must be valid JSON with the field names from the
`Definition`s in `model.go` (snake_case).

```go
package tests

import (
	"testing"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"github.com/tinywasm/model"
	"github.com/tinywasm/router"
	"github.com/tinywasm/router/mock"
)

func TestMountOps_CreateVisit(t *testing.T) {
	m := setup(t)
	reg := &mock.Router{}
	m.MountOps(reg)
	reg.Configure(mock.Config{
		Authn:     func(next router.HandlerFunc) router.HandlerFunc { return next },
		Authorize: func(userID string, resource model.Resource, action model.Action) bool { return true },
	})

	ctx := &mock.Context{
		InBody: []byte(`{"patient_id":"pat_1","doctor_id":"doc_1","reason":"Control","attention_at":1700000000,"patient_name_snapshot":"Juan","patient_rut_snapshot":"1-9","doctor_name_snapshot":"Dr. Soto"}`),
	}
	ctx.SetUserID("test-user")
	reg.Invoke("OP", "/"+clinicalencounter.OpCreateVisit, ctx)

	if ctx.Status != 0 && ctx.Status != 200 {
		t.Fatalf("expected success status, got %d, body=%s", ctx.Status, ctx.ResponseBody())
	}
	if len(ctx.ResponseBody()) == 0 {
		t.Fatal("expected a non-empty response body")
	}
}
```

If `mock.Config`'s exact field names differ from what is shown above once you check
`github.com/tinywasm/router/mock`'s current godoc, adjust to match — the intent (allow the guarded
`medical_history`/`Create` check to pass so the handler runs) does not change. Add the remaining
coverage from the checklist below directly modeled on this test: `get_medical_history`,
`list_medical_history`, and an RBAC-denial case (omit `Authorize`/`SetUserID` and assert `ctx.Status
== 403`).

### `tests/conformance_test.go` — `view.Presenter` round-trip

Ecosystem convention: the conformance-style view test lives in its **own file**
`tests/conformance_test.go` (keeps test files small and modular — same rule item_catalog adopted
after review). Use `github.com/tinywasm/view/conformance`'s `FakeCaller` (codec-free,
`view`+`model`+`router` only — no encoder import needed), and drive the Presenter with the same
`Reload()`/`Items()` calls `item_catalog/tests/conformance_test.go` uses in green (that file is the
working reference — copy its shape, not this prose):

`view.Presenter` has no `CanSave()`/`CanDelete()` method — `Saver`/`Deleter` are capabilities the
renderer (or, here, the test) discovers by type assertion (see the `Presenter`/`Saver`/`Deleter`
doc comments in `tinywasm/view@v0.1.12`):

```go
package tests

import (
	"testing"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"github.com/tinywasm/model"
	"github.com/tinywasm/view"
	"github.com/tinywasm/view/conformance"
)

func TestView_ListPopulatesItems(t *testing.T) {
	caller := &conformance.FakeCaller{
		Reply: func(op string, into model.Decodable) {
			if op != clinicalencounter.OpListVisitsByPatient {
				return
			}
			list := into.(*clinicalencounter.MedicalHistoryList)
			rec := list.Append().(*clinicalencounter.MedicalHistory)
			rec.Id, rec.Reason, rec.Status = "mh_1", "Control", clinicalencounter.StatusCompleted
		},
	}
	p := clinicalencounter.NewView(caller)
	if err := p.Reload(); err != nil {
		t.Fatalf("Reload: %v", err)
	}
	items := p.Items()
	if len(items) != 1 || items[0].ID != "mh_1" || items[0].Label != "Control" {
		t.Fatalf("unexpected items: %+v", items)
	}
	if _, ok := p.(view.Saver); !ok {
		t.Error("expected Saver capability (WithSaveOp is configured)")
	}
	if _, ok := p.(view.Deleter); ok {
		t.Error("expected no Deleter capability (no WithDeleteOp by design)")
	}
}
```

### Testing checklist (from `AGENTS.md` — confirm each is covered before declaring this stage done)

- [x] Happy path for each service method (`TestCreateVisit_HappyPath`, `TestListVisitsByPatient`)
- [ ] Tenant isolation — **not applicable**: this module has no `TenantId` field (see Stage 1 note)
- [x] Not-found errors (`TestGetVisit_NotFound`)
- [ ] Each op via `MountOps` against `router/mock` — routed correctly + RBAC enforced (extend
      `TestMountOps_CreateVisit` with the RBAC-denial case and the two remaining ops)
- [ ] `events.Publisher` fake receives the correct topic and typed payload (add
      `TestCreateVisit_PublishesEvent` using `mockPublisher` from `setup_test.go`)
- [ ] `view.Presenter` list/select/save round-trip against a fake `router.Caller`
      (`tests/conformance_test.go`)

### Stage 6 acceptance criteria

- `grep -rn "tinywasm/sqlite\|tinywasm/sqlt\|tinywasm/postgres" tests/` → empty.
- `grep -n "storage/mem" tests/setup_test.go` → 1 match.
- `gotest ./...` green, all checklist items above checked.

---

## Stage 7 — `go.mod` end state

### Current `go.mod` (verbatim)

```
module github.com/veltylabs/clinical_encounter

go 1.25.2

require (
	github.com/tinywasm/dom v0.10.1
	github.com/tinywasm/fmt v0.24.0
	github.com/tinywasm/layout v0.0.1
	github.com/tinywasm/orm v0.8.0
	github.com/tinywasm/time v0.5.0
	github.com/tinywasm/unixid v0.2.23
)

require (
	github.com/tinywasm/css v0.1.2 // indirect
	github.com/tinywasm/html v0.0.3 // indirect
)

replace github.com/tinywasm/layout => /home/cesar/Dev/Project/tinywasm/layout
```

### Target `go.mod` (versions pinned to `veltylabs/item_catalog`'s current `go.mod` — the reference
implementation; re-check that repo if this plan is executed much later than it was written. As of
this revision, re-verified by actually building every stage's code above against the real published
modules, not guessed)

```
module github.com/veltylabs/clinical_encounter

go 1.25.2

require (
	github.com/tinywasm/ddl v0.0.7
	github.com/tinywasm/events v0.0.2
	github.com/tinywasm/fmt v0.25.5
	github.com/tinywasm/form v0.3.13
	github.com/tinywasm/model v0.1.2
	github.com/tinywasm/orm v0.11.4
	github.com/tinywasm/router v0.1.19
	github.com/tinywasm/storage v0.0.2
	github.com/tinywasm/time v0.5.2
	github.com/tinywasm/view v0.1.12
)
```

Remove entirely: `github.com/tinywasm/dom` **as a direct dependency** (and `github.com/tinywasm/layout`
and its local `replace` directive — a `replace` pointing at a local path is itself a defect
`AGENTS.md`'s blacklist calls out), `github.com/tinywasm/unixid`, `github.com/tinywasm/html`,
`github.com/tinywasm/css` (drop once `go mod tidy` confirms nothing pulls it in transitively). Run
`go mod tidy` after applying every other stage — indirect entries will settle automatically; do not
hand-curate the indirect block. Expect `go mod tidy` to re-add `github.com/tinywasm/dom` as an
**indirect** entry (pulled transitively by `tinywasm/form`, via `tests/model_test.go`'s widget
regression test — the same pattern `item_catalog/tests/orm_test.go` already uses), plus
`github.com/tinywasm/json` and `github.com/tinywasm/widget` indirect — this is expected and correct;
the acceptance grep below only checks THIS module's own `.go` files, which never import `dom`
directly, so it stays green.

### Stage 7 acceptance criteria

- `grep -n "replace" go.mod` → empty.
- `grep -n "tinywasm/layout\|tinywasm/unixid\|tinywasm/html" go.mod` → empty (gone entirely, direct
  and indirect).
- `grep -n "tinywasm/dom" go.mod` → present, but only on a line ending `// indirect`. `go mod tidy`
  re-adds it transitively (via `tinywasm/form`, used by `tests/model_test.go`'s widget-regression
  test) — this module's own code never imports it directly; if the line does NOT say `// indirect`,
  something imports `tinywasm/dom` directly by mistake and that is a real defect to fix.
- `grep -n "tinywasm/ddl \|tinywasm/events\|tinywasm/router\|tinywasm/view\|tinywasm/model\|tinywasm/storage"
  go.mod` → all present.
- `go build ./...` and `gotest ./...` both succeed with the tidied `go.mod`/`go.sum`.

---

## Final acceptance criteria (repo-wide, run after all stages)

```
grep -rn --include=*.go "tinywasm/mcp\|tinywasm/json\|tinywasm/unixid\|tinywasm/sqlite\|tinywasm/sqlt\|tinywasm/postgres\|tinywasm/layout\|tinywasm/dom\|tinywasm/html" .
```
→ **empty, repo-wide, tests included** (per `AGENTS.md`'s blacklist and the master plan's §5
acceptance criteria for the harness pattern). `--include=*.go` matters here: `tinywasm/dom` and
`tinywasm/json` legitimately appear in `go.mod`/`go.sum` as **indirect** entries (`tinywasm/form`'s
own dependency graph — see Stage 7) — the rule is about what this module's `.go` files import, not
about the transitive closure recorded in `go.sum`. If this grep is run without `--include=*.go`, it
will find those two inside `go.mod`/`go.sum` and give a false failure even on a fully correct
implementation.

```
grep -rn "EventPublisher" .
```
→ empty (no self-declared port).

```
test -d web
```
→ does not exist.

```
grep -n "var _ router.OpModule" ops.go
```
→ 1 match.

```
grep -n "func NewView(caller router.Caller) view.Presenter" view.go
```
→ 1 match.

```
gotest ./...
```
→ green.

Not touched by any stage above, confirm still present and unmodified: `const.go`, `dicom.go` (empty
stub), `LICENSE`, `.gitignore`. Not touched at all, confirm no diff exists outside this repo:
`../clinical_encounterOld`.

---

## Stages summary

| # | Stage | Files | Output | Criterion |
|---|---|---|---|---|
| 0 | View reconciliation decision | (docs only — no code change) | Explicit record of what is deleted and why | Read before Stage 4 |
| 1 | `model.go` → `model.Definition` | `model.go`, `model_orm.go` (regenerated), `visit.go`/`module.go`/`viewCtrl.go` (casing only), `tests/model_test.go` | 4 `Definition`s, widget test | `gotest ./...` green |
| 2 | Identity/events ports | `module.go`, delete `publish.go` | `Deps{IDs, Publisher}` | no `unixid`/`EventPublisher` |
| 3 | Transport | new `ops.go`, `visit.go` (add `GetVisit`/`ListVisitsByPatient`) | `router.OpModule` | `var _ router.OpModule` compiles |
| 4 | View | delete `web/`, `viewCtrl.go`, `ssr.go`; rewrite `view.go` | `NewView(caller) view.Presenter` | no `dom`/`html`/`layout` imports |
| 5 | Schema migration | `module.go` (`New()`) | `ddl.CreateTable` type-assertion pattern | no `db.CreateTable` calls |
| 6 | Tests | `tests/setup_test.go`, `tests/visit_test.go`, `tests/ops_test.go`, `tests/conformance_test.go` | `storage/mem`-backed suite | `gotest ./...` green |
| 7 | `go.mod` | `go.mod`, `go.sum` | pinned to `item_catalog`'s current versions | no `replace`, no removed packages |
