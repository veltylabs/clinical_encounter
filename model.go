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
		// status: los valores válidos son ÚNICAMENTE las constantes Status* en const.go — los literales
		// viven solo en esas constantes y en ningún otro lugar (regla anti magic-string; ver la revisión de item_catalog).
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

// NotNull refleja exactamente el conjunto de argumentos requeridos que CreateVisit exige (visit.go) —
// el Definition es el único lugar donde se declara el contrato; las validaciones manuales del servicio
// son un duplicado de defensa en profundidad del mismo conjunto, nunca uno distinto.
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

// GetVisitArgsModel y ListVisitsArgsModel son solo de transporte (Field.DB es nil en todos los campos) —
// ver ops.go para las operaciones que los consumen.
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

// Topics de eventos de dominio — <módulo>.<entidad>.<verbo-en-pasado>, los datos de tenant/id van en
// el payload, nunca en el nombre del topic (no aplica aquí — ver la nota "no TenantId field" en ARCHITECTURE.md).
const (
	TopicVisitCreated = "clinical_encounter.visit.created"
)
