---
PLAN: "feat: ImportVisit — store a closed visit from a previous system with its real dates"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 16068151884085356589
PR: https://github.com/veltylabs/clinical_encounter/pull/14
REVIEWED_COMMIT: 87e7ac7a63524c3b9159bc9cf8b3aadf0d770746
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work.

# Plan — `ImportVisit`: bring the clinic's visit history with its real dates

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
```

Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`.

## Why

The clinic's previous system holds 64,416 visit records since 2015. The only path to a closed visit today is
`CreateVisit` (`visit.go:10`) then `CompleteVisit` (`complete.go:14`), and `CompleteVisit` sets
`FinishedAt = time.NowNano()`: a visit from 12-03-2018 would be closed *today*, and every patient's history would
show tens of thousands of visits "finished on the import day". `CreateVisit` also expands the service recipe into
supply lines (`m.supplyOn`), which must not happen for history. The importer (`mjosefa-cms/migrate/pastdb`) needs
one method that stores a visit as it happened.

## Ecosystem search

- `CreateVisit` validation (required fields, idempotent by id, `ErrIdTaken`) → **reuse the same rules**.
- Same contract as `patient_directory.ImportPatient` and `appointment_booking.ImportReservation` (their own plans).

## Design gate

1. Prior art: HL7 FHIR `Encounter.period` (https://hl7.org/fhir/encounter.html) — start and end are data of the
   encounter, not of when it was recorded; followed.
2. Name: `ImportVisit(v MedicalHistory)`.
3. Ledger: +1 method, +1 error. Ways to open a visit for a doctor: unchanged.
4. Belongs here: owner of `medical_history`.
5. Deletes nothing.

## Stage 1 — `import.go` (new)

```go
// ImportVisit stores a visit from a previous system already closed, with its own id and dates: it creates it, or
// updates it when that id already exists, so importing twice leaves the same row. It requires what CreateVisit
// requires (id, tenant, patient, doctor, reason, patient name and document snapshots, doctor name snapshot) and a
// real moment (AttentionAt > 0). It never uses "now" for a date, never expands supply recipes, publishes no event and
// opens no transaction (the importer already runs inside one).
//
// Example: v, err := enc.ImportVisit(MedicalHistory{Id: "8000000000000000005", TenantId: t, PatientId: p,
//	DoctorId: d, AttentionAt: at, Reason: "control", PatientNameSnapshot: "Ana", PatientRutSnapshot: "12345678-5",
//	DoctorNameSnapshot: "Dra. Pérez"})
func (m *Module) ImportVisit(v MedicalHistory) (*MedicalHistory, error)
```

Body:
1. `v.TenantId == ""` → `&model.FieldError{Field: "tenant_id", Err: ErrTenantRequired}`; any of `Id`, `PatientId`,
   `DoctorId`, `Reason`, `PatientNameSnapshot`, `PatientRutSnapshot`, `DoctorNameSnapshot` empty → `ErrMissingArgs`;
   `v.AttentionAt <= 0` → `ErrImportDate`.
2. `v.Status = StatusCompleted`; when `v.StartedAt == 0` → `v.StartedAt = time.Nano(int64(v.AttentionAt) * 1e9)`;
   when `v.FinishedAt == 0` → `v.FinishedAt = v.StartedAt`. `v.UpdatedAt = time.NowNano()` (this one is "when the row
   was written", correct as now).
3. `existing, err := m.GetVisit(v.TenantId, v.Id)`: found and `existing.PatientId == v.PatientId` → `m.db.Update(&v,
   orm.Eq(MedicalHistory_.Id, v.Id), orm.Eq(MedicalHistory_.TenantId, v.TenantId))`; found with another patient →
   `ErrIdTaken`; not found → `m.db.Create(&v)`.

New error next to `ErrMissingArgs` in `model.go`, same style: `ErrImportDate model.Invalid = "an imported visit needs the date it happened"`.

## Stage 2 — tests (`tests/import_test.go`)

| Test | Call | Expected |
|---|---|---|
| `TestImportVisitKeepsDates` | `AttentionAt` = 2018-03-12 10:00 UTC | `Status == completed`, `FinishedAt` and `StartedAt` equal that instant (not today) |
| `TestImportVisitTwice` | same id twice, second with new diagnostic | 1 row, diagnostic updated |
| `TestImportVisitOtherPatient` | same id, other patient | `ErrIdTaken` |
| `TestImportVisitNoDate` | `AttentionAt: 0` | `ErrImportDate` |
| `TestImportVisitMissingReason` | empty reason | `ErrMissingArgs` |
| `TestImportVisitNoSupplies` | module with supply on, service with a recipe | `ListSupplyUse` returns 0 lines |
| `TestImportVisitNoEvent` | counting fake Publisher | 0 events |

## Stage 3 — translations and docs

`langc sync .`, Spanish value for the new error. `README.md` + `docs/ARCHITECTURE.md`: section "Importar historia"
**in Spanish**: why the dates come from the data and not from "now", with the example.

## Delivery check

1. `gotest` passes (a PR that does not compile, or whose tests could not run, is a failed delivery).
2. `grep -n "NowUnix\|m.db.Tx\|Publish\|recipes" import.go` → empty.
3. `langc sync .` → 0 untranslated `es`.

Commit after every stage. Never edit a generated `*_orm.go`. Never run `git reset`, `git restore`, `git checkout --
<file>` or `git stash` on your own work. Stuck 30 minutes: open the PR anyway listing what fails.

## Stages

| Stage | Files | Done when |
|---|---|---|
| 1 | `import.go`, `model.go` (error) | compiles |
| 2 | `tests/import_test.go` | tests pass |
| 3 | `lang.json`, `README.md`, `docs/ARCHITECTURE.md` | done |
