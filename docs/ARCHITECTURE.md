# Clinical Encounter Architecture

## Domain Scope

`clinical_encounter` owns the medical-visit record for the Velty ecosystem: the history of a
patient's attentions (`MedicalHistory`) and the args to open a new one (`CreateVisitArgs`). It is
the source of truth other modules would read from if they needed a patient's visit history — no
other module reads or writes it today.

## Core Entities

- **`MedicalHistory`** (table `medical_history`): one row per patient visit. Identity (`Id`,
  `PatientId`, `DoctorId`, soft-optional `ReservationId`), FSM state (`Status`, see below),
  scheduling (`AttentionAt` — Unix timestamp, date+time unified), clinical content (`Reason`,
  `Diagnostic`, `Prescription`, `Cie10Code`), timing (`StartedAt`, `FinishedAt`), immutable
  point-in-time snapshots taken at visit creation (`PatientNameSnapshot`, `PatientRutSnapshot`,
  `DoctorNameSnapshot`, `DoctorSpecialtySnapshot`) so the record reads correctly even if the
  patient/doctor's profile changes later, and `UpdatedAt`. It carries no UI widgets — it is shown,
  never edited through a form directly.
- **`CreateVisitArgs`**: the transport-facing args to open a new `MedicalHistory` — same clinical
  fields minus `Id`/`Status`/`StartedAt`/`FinishedAt`/`Cie10Code`/`UpdatedAt` (server-assigned or not
  yet collected at creation time). Every field carries a UI widget (`input.Text()`/
  `input.Number()`/`input.Textarea()` for `Diagnostic`/`Prescription`) so a `tinywasm/form`-rendered
  form is never silently empty.

## The Visit FSM (`const.go` — not modified by this migration)

`const.go` declares the FHIR-aligned lifecycle as plain string constants — this module does not
touch that file, only builds on top of it:

```go
StatusCreated    = "created"     // scheduled, not yet at clinic
StatusArrived    = "arrived"     // patient registered at reception
StatusTriaged    = "triaged"     // nurse took vitals, waiting for doctor
StatusInProgress = "in_progress" // doctor started the consultation
StatusCompleted  = "completed"
StatusCancelled  = "cancelled"
```

`MedicalHistory.Status` holds one of these. As of this writing `const.go` declares only the
`Status*` set — there is no `Event*` constant group and no transition table (`visitTransitions`) in
the repo; a prior doc referenced one but it does not exist in code. `CreateVisit` is the only method
that assigns a status today (`StatusCreated`, on creation) — no transition methods
(arrive/triage/start/complete/cancel) exist yet. `router.OpModule` in this module therefore exposes
only the operations backed by real methods (create/list/get) — it does not invent transition ops for
FSM states that have no corresponding business logic yet. Adding those is a future, separate change.

## Patterns

- **Reusable-module harness**: the module is coupled only to published contracts, not to concrete
  infrastructure — see `AGENTS.md` (this repo's root) for the full whitelist/blacklist:
    - `orm.DB` for storage (backend-agnostic — wraps whatever `storage.Conn` the app injects).
    - `ddl.CreateTable` (over `db.RawConn()`, guarded by a `ddl.Compiler` type assertion) for the
      module's own schema migration in `New()`.
    - `router.OpModule` (`ModelName()` + `MountOps(reg router.OpRegistry)`) for transport — the
      module never imports `tinywasm/mcp` and never implements `router.Router`/`router.APIModule`.
    - `model.IDGenerator` for identity (`Deps.IDs`, required).
    - `events.Publisher` for event-driven updates (`Deps.Publisher`, optional — `nil` disables
      publishing silently). This module's own `EventPublisher` interface (previously in
      `publish.go`) is retired in favor of this contract.
    - `view.Presenter` (`NewView(caller router.Caller) view.Presenter`) for UI, built with only
      `view`+`model`+`router` — the app chooses the renderer. Replaces the previous `web/` package,
      which built a concrete `tinywasm/dom` + `tinywasm/layout/rightpanel` UI directly inside the
      module and is removed — see `docs/PLAN.md` Stage 0/4 for the full reconciliation history.
    - Tests run against `storage/mem` (`orm.New(mem.New())`), never a concrete driver, and never a
      hand-rolled `orm.Executor`/`orm.Compiler` fake (the retired `web/mockdb` package did this
      redundantly — `storage/mem` is the upstream, maintained equivalent).
- **Snapshot fields, not live joins**: `*Snapshot` fields on `MedicalHistory` freeze patient/doctor
  identity at visit-creation time. This module never queries a `patient`/`doctor` table — those
  belong to other modules this one does not import.
- **No multi-tenancy field today**: unlike `item_catalog`, `MedicalHistory`/`CreateVisitArgs` carry
  no `TenantId` field in the current schema. This migration ports the existing columns mechanically
  (same table, same columns, same behavior) and does not add one — see `docs/PLAN.md` Stage 1. If
  this module needs tenant isolation later, that is a separate, explicit schema change, not a
  side-effect of the harness migration.

## Ops (via `MountOps`)

| Op | Action | Resource | Description |
|---|---|---|---|
| `create_visit` | `c` | `medical_history` | Create a new visit from `CreateVisitArgs` — assigns `Id`, sets `Status = StatusCreated` |
| `get_medical_history` | `r` | `medical_history` | Get one visit record by ID |
| `list_medical_history` | `r` | `medical_history` | List a patient's visit history |

No status-transition ops (`arrive`/`triage`/`start`/`complete`/`cancel`) exist yet — see "The Visit
FSM" above. Adding them is a future change with its own plan, once the corresponding `*Module`
methods exist.

## Composition Root Example

```go
enc, _ := clinicalencounter.New(db, clinicalencounter.Deps{
    IDs:       idGenerator,   // model.IDGenerator
    Publisher: eventPublisher, // events.Publisher, nil disables publishing
})
enc.MountOps(opRegistry) // router.OpRegistry
view := enc.NewView(caller) // router.Caller -> view.Presenter, rendered by whatever the app chooses
```

## Out of Scope

- `const.go` (FSM constants) — described above, never edited by this module's harness migration.
- `dicom.go` — currently an empty stub (`package clinical_encounter`, no code); left untouched.
- `clinical_encounterOld` — a legacy sibling directory with no importers anywhere in the ecosystem.
  It is not part of this repo and must never be read, referenced, or migrated by any change here.
