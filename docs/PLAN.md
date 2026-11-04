---
PLAN: "feat!: decouple clinical_encounter from sibling modules via interfaces.go and OpMyAgenda"
EXECUTOR: jules
REVIEWER: none
STATUS: dispatch
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work.

# Plan — decouple `clinical_encounter` from all sibling modules

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/ormc/cmd/ormc@latest
go get webtyp.com/time@v0.6.4 webtyp.com/appkit@v0.2.7
```

Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`, `go mod tidy` before `ormc`.

## Why

Of the 18 modules in `veltylabs/modules/*`, 16 have zero cross-module imports. `clinical_encounter` still imports 6 sibling modules (`appointment_booking`, `patient_directory`, `staff_manager`, `service_catalog`, `business_calendar`, `device_manager`) across `model.go`, `ui/browser.go`, `seed/seed.go`, `web/browser.go`, and `tests/`. A domain module must never import another module in any package: cross-module references in `model.go` are soft IDs, contracts live in `interfaces.go` at the module root with plain types, `ui/` calls only `clinical_encounter`'s own operations (`OpMyAgenda`), and the consuming application wires dependencies via `appkit.Find` in `module/`.

## Ecosystem search

- `webtyp.com/appkit` (v0.2.7): `appkit.Module`, `appkit.ModuleServer`, `appkit.ModuleBrowser`, `appkit.Find[T](d)` → **reuse** in `module/`.
- `webtyp.com/time` (v0.6.4): `time.Unix`, `time.Zone` (`Day`, `Date`, `Clock`) → **reuse** (import without alias as `time`, never `tinytime` or `tintime`).
- Reference module layout: `veltylabs/modules/room_layout` (`interfaces.go`, `module/{module.go,server.go,browser.go}`, `web/browser/main.go`).

## Design gate

1. Prior art: Hexagonal Architecture ports (Cockburn), Go standard library consumer-owned narrow interfaces (`io.Reader`), `veltylabs/modules/room_layout`.
2. Novice-name test: `DoctorReader{StaffByUser}`, `AgendaReader{ConfirmedByStaff}`, `PatientReader{PatientIdentity}` (uses `documentNumber`, never `rut`), `OpMyAgenda` (`"my_agenda"`).
3. Complexity ledger:
   - Sibling module imports in `go.mod`: 0 (−6)
   - Cross-module RPC calls from browser `ui/`: 0 (−3; replaced by 1 call to `clinical_encounter.my_agenda`)
   - Ways to wire the module in an app: 1 (`module.Module` via `appkit`)
4. Where it belongs: `interfaces.go` at root of `clinical_encounter`; `module/` subpackage for `appkit.Module`.
5. Deletes: all 6 `github.com/veltylabs/*` imports in `go.mod`, cross-module `Ref` in `model.go`, direct calls to `staff_manager`, `appointment_booking`, and `patient_directory` in `ui/browser.go`, and `web/browser.go` (moved to `web/browser/main.go` with in-memory fakes).

## Stage 1 — `interfaces.go` and `module.go`

Create `interfaces.go` at the root of `clinical_encounter` (move `RecipeReader`, `RoomFinder`, and `SupplyStock` out of `module.go` into `interfaces.go`, and add `DoctorReader`, `AgendaReader`, and `PatientReader`):

```go
package clinical_encounter

import "webtyp.com/time"

// RecipeReader es la parte de service_catalog que este módulo utiliza.
// Ejemplo: ids, qtys, err := r.ServiceRecipe("t1", "svc1")
type RecipeReader interface {
	ServiceRecipe(tenantId, serviceId string) (itemIds []string, quantities []float64, err error)
}

// RoomFinder es la parte de room_layout que este módulo utiliza.
// Ejemplo: roomID, ok, err := f.RoomOfOccupant("t1", "doc1", at)
type RoomFinder interface {
	RoomOfOccupant(tenantID, occupantID string, at time.Unix) (roomID string, found bool, err error)
}

// SupplyStock es la parte de supply_manager que este módulo utiliza.
// Ejemplo: cost, err := s.UseInRoom("t1", "room1", "item1", 2, "encounter", "enc1", "u1")
type SupplyStock interface {
	UseInRoom(tenantId, roomId, itemId string, quantity float64, docType, docId, by string) (int64, error)
}

// DoctorReader resuelve qué profesional clínico corresponde a una cuenta de usuario.
// Ejemplo: staffID, name, spec, ok, err := d.StaffByUser("t1", "u1")
type DoctorReader interface {
	StaffByUser(tenantID, userID string) (staffID, name, specialty string, found bool, err error)
}

// AgendaReader lista las reservas confirmadas de un profesional en un rango de tiempo [from, to).
// Ejemplo: resIDs, patIDs, svcIDs, starts, err := a.ConfirmedByStaff("t1", "doc1", from, to)
type AgendaReader interface {
	ConfirmedByStaff(tenantID, staffID string, from, to time.Unix) (reservationIDs, patientIDs, serviceIDs []string, startTimes []time.Unix, err error)
}

// PatientReader obtiene el nombre y número de documento de un paciente por su ID.
// Ejemplo: name, doc, ok, err := p.PatientIdentity("t1", "p1")
type PatientReader interface {
	PatientIdentity(tenantID, patientID string) (name, documentNumber string, found bool, err error)
}
```

In `module.go`:
- Remove `RecipeReader`, `RoomFinder`, and `SupplyStock` definitions (now in `interfaces.go`).
- Add to `Deps`:
  - `TenantID string`
  - `Zone time.Zone`
  - `Doctors DoctorReader`
  - `Agenda AgendaReader`
  - `Patients PatientReader`
- Store `tenantID`, `zone`, `doctors`, `agenda`, `patients` on `Module` in `New`.

## Stage 2 — `model.go` + `ormc`, `agenda.go`, and `ops.go`

In `model.go`:
- Delete imports of `github.com/veltylabs/appointment_booking`, `github.com/veltylabs/patient_directory`, `github.com/veltylabs/service_catalog`, and `github.com/veltylabs/staff_manager`.
- In `MedicalHistoryModel` (lines 18–21) and `CreateVisitArgsModel` (lines 50–51, 57–58), remove `Ref: &...` and `DB: &model.FieldDB{RefColumn: "id"}` from `patient_id`, `doctor_id`, `reservation_id`, and `service_id`:
  ```go
  {Name: "patient_id", Type: input.Text(), NotNull: true},
  {Name: "doctor_id", Type: input.Text(), NotNull: true},
  {Name: "reservation_id", Type: input.Text()},
  {Name: "service_id", Type: input.Text()},
  ```
- Append the transport definitions for `OpMyAgenda`:
  ```go
  var MyAgendaArgsModel = model.Definition{
  	Name: "my_agenda_args",
  	Fields: model.Fields{
  		{Name: "tenant_id", Type: model.Text()},
  		{Name: "user_id", Type: model.Text()},
  		{Name: "date", Type: model.Text()},
  	},
  }

  var AgendaEntryModel = model.Definition{
  	Name: "agenda_entry",
  	Fields: model.Fields{
  		{Name: "patient_id", Type: model.Text()},
  		{Name: "patient_name", Type: model.Text()},
  		{Name: "patient_document", Type: model.Text()},
  		{Name: "time_label", Type: model.Text()},
  		{Name: "reservation_id", Type: model.Text()},
  		{Name: "service_id", Type: model.Text()},
  	},
  }

  var MyAgendaResultModel = model.Definition{
  	Name: "my_agenda_result",
  	Fields: model.Fields{
  		{Name: "staff_id", Type: model.Text()},
  		{Name: "doctor_name", Type: model.Text()},
  		{Name: "doctor_specialty", Type: model.Text()},
  		{Name: "entries", Type: model.StructSlice(&AgendaEntryModel)},
  	},
  }
  ```
- Run `ormc` and `gofmt -w .` (never edit `model_orm.go` by hand).

Create `agenda.go` at the root of `clinical_encounter`:
- Implement `func (m *Module) MyAgenda(args MyAgendaArgs) (*MyAgendaResult, error)`:
  - If `args.TenantId == ""`, use `m.tenantID`.
  - If `m.doctors == nil` or `args.UserId == ""`, return `&MyAgendaResult{}, nil`.
  - Call `staffID, docName, docSpec, found, err := m.doctors.StaffByUser(args.TenantId, args.UserId)`. If `err != nil`, return `nil, err`; if `!found || staffID == ""`, return `&MyAgendaResult{}, nil`.
  - Determine `date := args.Date`; if `date == ""` and `m.zone.Name() != ""`, set `date = m.zone.Date(time.NowUnix())`.
  - Compute `from, to time.Unix`: if `m.zone.Name() != ""`, call `from, to, err = m.zone.Day(date)`; otherwise parse `dayKey, err := time.ParseDateUTC(date)` and set `from, to = dayKey, dayKey+time.Day`. If `err != nil`, return `nil, ErrMissingArgs`.
  - Build `out := &MyAgendaResult{StaffId: staffID, DoctorName: docName, DoctorSpecialty: docSpec}`.
  - If `m.agenda == nil`, return `out, nil`.
  - Call `resIDs, patIDs, svcIDs, starts, err := m.agenda.ConfirmedByStaff(args.TenantId, staffID, from, to)`. If `err != nil`, return `nil, err`.
  - For each index `i` in `resIDs`:
    - If `m.patients == nil`, continue.
    - Call `patName, patDoc, ok, err := m.patients.PatientIdentity(args.TenantId, patIDs[i])`. If `err != nil || !ok`, continue (an unresolvable patient does not break the whole agenda).
    - Compute `timeLabel := ""` (if `m.zone.Name() != ""` and `i < len(starts)`, `timeLabel = m.zone.Clock(starts[i])`).
    - Append `AgendaEntry{PatientId: patIDs[i], PatientName: patName, PatientDocument: patDoc, TimeLabel: timeLabel, ReservationId: resIDs[i], ServiceId: svcIDs[i]}` to `out.Entries`.
  - Return `out, nil`.

In `ops.go`:
- Add `OpMyAgenda = "my_agenda"` to the `Op*` constants.
- Register in `MountOperations`:
  `reg.Operation(OpMyAgenda, m.opMyAgenda).Requires(model.ResourceOf(&MedicalHistoryModel), model.Read).Accepts(&MyAgendaArgs{})`
- Add `opMyAgenda(ctx router.Context)`: decode `MyAgendaArgs`, if `args.UserId == ""` set `args.UserId = ctx.UserID()`, call `m.MyAgenda(args)`, and encode the result.

## Stage 3 — `ui/browser.go`, `seed/seed.go`, `module/`, and `web/browser/main.go`

In `ui/browser.go`:
- Delete imports of `github.com/veltylabs/appointment_booking`, `github.com/veltylabs/patient_directory`, `github.com/veltylabs/staff_manager`, and replace `tintime "webtyp.com/time"` with `"webtyp.com/time"`.
- Replace `todayRange`, `doctorIdentity`, `loadDoctorIdentity`, and `loadTodayAgenda` with a single helper that calls `clinicalencounter.ModelName + "." + clinicalencounter.OpMyAgenda` passing `&clinicalencounter.MyAgendaArgs{TenantId: tenantID, UserId: userID, Date: zone.Date(time.NowUnix())}` into `var myAgenda clinicalencounter.MyAgendaResult`.
- If `myAgenda.StaffId != ""`, iterate over `myAgenda.Entries`: for each entry `e`, call `clinicalencounter.ModelName + "." + clinicalencounter.OpListVisitsByPatient` with `&clinicalencounter.ListVisitsArgs{TenantId: tenantID, PatientId: e.PatientId}` to build `items []model.Item`, and append `agendaPatient{ID: e.PatientId, Label: e.PatientName, Sublabel: e.PatientDocument, Description: e.TimeLabel, ReservationID: e.ReservationId, ServiceID: e.ServiceId, History: items}`.
- Use `myAgenda.StaffId`, `myAgenda.DoctorName`, and `myAgenda.DoctorSpecialty` in `cfg.NewRecord`.

In `seed/seed.go`:
- Delete imports of `github.com/veltylabs/patient_directory/seed` and `github.com/veltylabs/staff_manager/seed`.
- Replace `Upstream` with plain types owned by `seed`:
  ```go
  type PatientSeed struct {
  	ID             string
  	Name           string
  	DocumentNumber string
  }

  type DoctorSeed struct {
  	ID        string
  	Name      string
  	Specialty string
  }

  type LoadOptions struct {
  	Patients []PatientSeed
  	Doctors  []DoctorSeed
  }
  ```
- `func Load(m *clinicalencounter.Module, tenantID string, opts ...LoadOptions) (Data, error)`: if `opts` is empty or has no `Patients`/`Doctors`, use default demo values (2 patients and 1 doctor); if an explicit `LoadOptions` was passed by a caller with `< 2` patients or `< 1` doctor, return an error as before.

Create `module/module.go`, `module/server.go` (`//go:build !browser`), and `module/browser.go` (`//go:build browser`):
- `module/module.go`:
  ```go
  package module

  import "webtyp.com/appkit"

  // Module conecta clinical_encounter en App.Modules de una aplicación.
  // Ejemplo: Modules: []appkit.Module{clinical.Module}
  var Module = appkit.Module{Name: "clinical_encounter", Server: server, Browser: browser}
  ```
- `module/server.go` (`//go:build !browser`):
  Resolves `DoctorReader`, `AgendaReader`, `PatientReader` via `appkit.Find` (and optional `RecipeReader`, `RoomFinder`, `SupplyStock` if registered in `d`), and calls `clinicalencounter.New(d.DB, clinicalencounter.Deps{IDs: d.IDs, Publisher: d.Events, TenantID: d.TenantID, Zone: d.Zone, Doctors: doc, Agenda: ag, Patients: pat, ...})`.
- `module/browser.go` (`//go:build browser`):
  `var browser = appkit.ModuleBrowser{Screens: func(d appkit.BrowserDeps) []view.Screen { return ui.Screens(d.Caller, d.Zone, nil, "", "") }}`.

Replace `web/browser.go` with `web/browser/main.go` (delete `web/browser.go`; no build tag on `web/browser/main.go`):
- Build `clinicalencounter.New` with in-memory fakes implementing `DoctorReader`, `AgendaReader`, and `PatientReader`, seed with `seed.Load(ceMod, demoTenantID)`, mount `ceMod` on `loopback.WithTenant`, and render `ui.Screens(caller, zone, ids, demoTenantID, "demo-user-ana")`. Zero imports of other `veltylabs/*` modules.

## Stage 4 — `tests/`, docs, and `go.mod`

- Update `tests/seed_test.go`: remove `patient_directory` and `staff_manager` imports; pass `ceseed.LoadOptions{Patients: ..., Doctors: ...}` and also test calling `ceseed.Load(m, "t1")` with default options and with insufficient options.
- Update `tests/ui_agenda_browser_test.go`: remove `appointment_booking`, `patient_directory`, and `staff_manager` imports; mock `clinicalencounter.ModelName + "." + clinicalencounter.OpMyAgenda` returning `clinicalencounter.MyAgendaResult` and `clinicalencounter.ModelName + "." + clinicalencounter.OpListVisitsByPatient`.
- Add `tests/agenda_test.go`: test `MyAgenda` and `OpMyAgenda` on the server with fake `DoctorReader`, `AgendaReader`, and `PatientReader` (doctor found with confirmed reservations, unknown user returns empty agenda, unresolvable patient is skipped).
- Update `README.md` and `docs/ARCHITECTURE.md` (in Spanish) to document `interfaces.go` (`DoctorReader`, `AgendaReader`, `PatientReader`, `RecipeReader`, `RoomFinder`, `SupplyStock`), `OpMyAgenda`, and `module.Module`.
- Run `go mod tidy` so all `github.com/veltylabs/*` dependencies disappear from `go.mod` and `go.sum`.

## Delivery check

1. `gotest` passes.
2. `grep -rn "github.com/veltylabs/" --include="*.go" go.mod . | grep -v "github.com/veltylabs/clinical_encounter"` → empty.
3. `grep -rn "tinytime\|tintime" --include="*.go" .` → empty.

Commit after every stage. Never edit `model_orm.go` by hand.

## Stages

| Stage | Files | Done when |
|---|---|---|
| 1 | `interfaces.go`, `module.go` | compiles |
| 2 | `model.go`, `model_orm.go` (ormc), `agenda.go`, `ops.go` | ormc + compiles |
| 3 | `ui/browser.go`, `seed/seed.go`, `module/*`, `web/browser/main.go` (delete `web/browser.go`) | compiles |
| 4 | `tests/*`, `README.md`, `docs/ARCHITECTURE.md`, `go.mod`, `go.sum` | `gotest` passes & zero sibling imports |
