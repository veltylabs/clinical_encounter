---
PLAN: "feat: export ui.Screens, drop layout dependency, resolve langc reuse warnings"
TAG: v0.2.24
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 12319542210229332773
---

> Read AGENTS.md at the repo root before any change.

## Tools

```
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go install webtyp.com/ormc/cmd/ormc@latest
```

## Why

Phase M of webtyp app composition (`webtyp/docs/APP_COMPOSITION_MASTER_PLAN.md`): modules declare their screens via `Screens(caller router.Caller) []view.Screen` instead of legacy `Browser(...) (platformd.UIModule, error)`, and drop `webtyp.com/layout` from `go.mod`.
Additionally, `langc sync .` detected a reuse warning on `doctor_name_snapshot` and `patient_name_snapshot` sharing `name snapshot`; distinct labels must be assigned in `model.go`, `view.Presenter` updated with options, and Spanish translations completed in `lang.json`.

## APIs used

| symbol | exact signature | tag | file |
|---|---|---|---|
| `view.Screen` | `type Screen struct { ID string; Label lang.Text; Icon string; Permission model.Resource; Presenter Presenter; Lookups []Lookup; Help lang.Text; Children []Screen }` | `webtyp.com/view@v0.6.45` | `view/screen.go` |
| `view.WithDescription` | `func WithDescription(d lang.Text) Option` | `webtyp.com/view@v0.6.45` | `view/options.go` |
| `view.WithTitle` | `func WithTitle(t string) Option` | `webtyp.com/view@v0.6.45` | `view/options.go` |
| `view.New` | `func New(source Lister, target model.Model, opts ...Option) Presenter` | `webtyp.com/view@v0.6.45` | `view/view.go` |
| `router.Caller` | `type Caller interface { Call(op string, args, out any, done func(error)) }` | `webtyp.com/router@v0.4.8` | `router/caller.go` |

If a signature in the published tag differs from this table, stop and name the symbol in the PR; never adapt it by guessing.

## Stage 1: Explicit labels in model.go and ormc generation

In `model.go`:
1. In `MedicalHistoryModel`, give explicit labels to fields that otherwise trigger reuse warnings:
```go
		{Name: "patient_name_snapshot", Type: input.Text(), Label: "Patient name", NotNull: true},
		{Name: "patient_rut_snapshot", Type: input.Rut(), Label: "Patient RUT", NotNull: true},
		{Name: "doctor_name_snapshot", Type: input.Text(), Label: "Doctor name", NotNull: true},
		{Name: "doctor_specialty_snapshot", Type: input.Text(), Label: "Doctor specialty"},
```
2. In `CreateVisitArgsModel`, apply the matching labels:
```go
		{Name: "patient_name_snapshot", Type: input.Text(), Label: "Patient name", NotNull: true},
		{Name: "patient_rut_snapshot", Type: input.Text(), Label: "Patient RUT", NotNull: true},
		{Name: "doctor_name_snapshot", Type: input.Text(), Label: "Doctor name", NotNull: true},
		{Name: "doctor_specialty_snapshot", Type: input.Text(), Label: "Doctor specialty"},
```
3. Run `ormc` to regenerate `model_orm.go`.

Files: `model.go`, `model_orm.go`

## Stage 2: Update NewView in view.go

In `view.go`:
1. Define constant:
```go
const ResourceMedicalHistory = "medical_history"
```
2. Update `NewView` to accept `opts ...view.Option`:
```go
// NewView construye el Presenter del historial médico.
func NewView(caller router.Caller, opts ...view.Option) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListVisitsByPatient, Save: OpCreateVisit},
		func() model.ModelSlice { return &MedicalHistoryList{} })
	allOpts := append([]view.Option{view.WithTitle(titleMedicalHistory)}, opts...)
	return view.New(b, &MedicalHistory{}, allOpts...)
}
```

Files: `view.go`

## Stage 3: Declarative Screens and drop layout from go.mod

1. In `ui/browser.go`:
Replace with:
```go
//go:build browser

package ui

import (
	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"webtyp.com/lang"
	"webtyp.com/router"
	"webtyp.com/view"
)

const (
	ID    = "clinical_encounter"
	Label = "Historial clínico"
)

const (
	descHistory lang.Text = "Medical history, patient clinical visits, consultations and diagnoses."
)

// Screens declara las pantallas que este módulo ofrece al menú de la app.
func Screens(caller router.Caller) []view.Screen {
	return []view.Screen{
		{
			ID:         ID,
			Label:      Label,
			Icon:       ID,
			Permission: clinicalencounter.ResourceMedicalHistory,
			Presenter:  clinicalencounter.NewView(caller, view.WithDescription(descHistory)),
		},
	}
}
```
2. Remove `webtyp.com/layout` from `go.mod`. Run `go mod tidy`.

Files: `ui/browser.go`, `go.mod`, `go.sum`

## Stage 4: Test and Translations

1. Replace `tests/ui_agenda_browser_test.go` with `tests/ui_screens_test.go`:
```go
package tests

import (
	"testing"

	"github.com/veltylabs/clinical_encounter/ui"
	"webtyp.com/router/mock"
)

func TestUIScreens(t *testing.T) {
	caller := mock.NewCaller()
	screens := ui.Screens(caller)
	if len(screens) == 0 {
		t.Fatal("expected at least 1 screen")
	}
	for _, s := range screens {
		if err := s.Validate(); err != nil {
			t.Fatalf("screen %s invalid: %v", s.ID, err)
		}
		if s.Presenter == nil {
			t.Fatalf("screen %s expected Presenter", s.ID)
		}
		if s.Presenter.Description() == "" {
			t.Fatalf("screen %s expected non-empty description", s.ID)
		}
	}
}
```
Delete `tests/ui_agenda_browser_test.go`.

2. Run `langc sync .`.
3. Complete 100% of Spanish translations `(es)` in `lang.json`.
4. Ensure `langc sync .` reports:
   - 0 untranslated `(es)`
   - 0 reuse warnings
   - 0 identical values

Files: `tests/ui_screens_test.go`, `tests/ui_agenda_browser_test.go` (deleted), `lang.json`

## Delivery check

Run:
```
gotest
langc sync .
```
Both commands must exit 0 with 0 errors, 0 untranslated `(es)`, and 0 reuse warnings.
Commit after every stage. When all checks pass, push and open the PR.

## Stages

| # | Stage | Files |
|---|---|---|
| 1 | Model labels and ormc | `model.go`, `model_orm.go` |
| 2 | NewView with options | `view.go` |
| 3 | Declarative screens & drop layout | `ui/browser.go`, `go.mod`, `go.sum` |
| 4 | Test and translations | `tests/ui_screens_test.go`, `lang.json` |
