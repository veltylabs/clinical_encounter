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
