package clinical_encounter

import (
	"webtyp.com/input"
	"webtyp.com/model"
)

var MedicalHistoryModel = model.Definition{
	Name:          "medical_history",
	Allow:         model.Read | model.Update,
	Rules:         model.Rules{Read: model.HaveNoBusinessRule, Update: model.AddMyBusinessRule},
	ExternalCalls: []model.ExternalCallDefinition{{Name: "DeductSupplies", Mode: model.RetryInBackground, MaxAttempts: 3}},
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "patient_id", Type: input.Text(), NotNull: true},
		{Name: "doctor_id", Type: input.Text(), NotNull: true},
		{Name: "reservation_id", Type: input.Text()},
		{Name: "service_id", Type: input.Text()},
		{Name: "status", Type: input.Select(), NotNull: true, Item: model.ItemDescription},
		{Name: "attention_at", Type: model.UnixTime(), NotNull: true, Item: model.ItemBadge},
		{Name: "reason", Type: input.Textarea(), NotNull: true, Item: model.ItemLabel},
		{Name: "diagnostic", Type: input.Textarea()},
		{Name: "prescription", Type: input.Textarea()},
		{Name: "cie10_code", Type: input.Text()},
		{Name: "started_at", Type: model.NanoTime()},
		{Name: "finished_at", Type: model.NanoTime()},
		{Name: "patient_name_snapshot", Type: input.Text(), Label: "patient name", NotNull: true},
		{Name: "patient_rut_snapshot", Type: model.Text()},
		{Name: "doctor_name_snapshot", Type: input.Text(), Label: "doctor name", NotNull: true},
		{Name: "doctor_specialty_snapshot", Type: input.Text()},
		{Name: "updated_at", Type: model.NanoTime(), NotNull: true, ServerOnly: true},
	},
}

var CreateVisitArgsModel = model.Definition{
	Name: "create_visit_args",
	Wire: true,
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, NotNull: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "patient_id", Type: input.Text(), NotNull: true},
		{Name: "doctor_id", Type: input.Text(), NotNull: true},
		{Name: "attention_at", Type: input.Number(), NotNull: true},
		{Name: "reason", Type: input.Text(), NotNull: true},
		{Name: "patient_name_snapshot", Type: input.Text(), Label: "patient name", NotNull: true},
		{Name: "patient_rut_snapshot", Type: input.Text()}, // vacío si el paciente no tiene documento
		{Name: "doctor_name_snapshot", Type: input.Text(), Label: "doctor name", NotNull: true},
		{Name: "reservation_id", Type: input.Text()},
		{Name: "service_id", Type: input.Text()},
		{Name: "diagnostic", Type: input.Textarea()},
		{Name: "prescription", Type: input.Textarea()},
		{Name: "doctor_specialty_snapshot", Type: input.Text()},
	},
}

var MyAgendaArgsModel = model.Definition{
	Name: "my_agenda_args",
	Wire: true,
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
	Name:  "my_agenda_result",
	Allow: model.Read,
	Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), OmitEmpty: true}, // filter
		{Name: "date", Type: model.Text(), OmitEmpty: true},      // filter
		{Name: "staff_id", Type: model.Text()},
		{Name: "doctor_name", Type: model.Text()},
		{Name: "doctor_specialty", Type: model.Text()},
		{Name: "entries", Type: model.StructSlice(&AgendaEntryModel)},
	},
}

const (
	ErrNotFound    model.NotFound = "visit not found"
	ErrMissingArgs model.Invalid  = "missing required arguments"
	ErrImportDate  model.Invalid  = "an imported visit needs the date it happened"
	ErrIdTaken     model.Conflict = "visit id already used by another patient"
)

var EncounterSupplyModel = model.Definition{
	Name:          "encounter_supply",
	Allow:         model.Create | model.Read | model.Update | model.Delete,
	Rules:         model.Rules{Create: model.AddMyBusinessRule, Read: model.HaveNoBusinessRule, Update: model.AddMyBusinessRule, Delete: model.AddMyBusinessRule},
	ExternalCalls: []model.ExternalCallDefinition{{Name: "Deduct", Mode: model.RetryInBackground, MaxAttempts: 3}},
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "encounter_id", Type: model.Text(), Ref: &MedicalHistoryModel, DB: &model.FieldDB{RefColumn: "id"}, NotNull: true},
		{Name: "item_id", Type: model.Text(), NotNull: true},
		{Name: "quantity", Type: input.Decimal(), Label: "Quantity", NotNull: true},
		{Name: "billable", Type: input.Checkbox(), Label: "Billable"},
		{Name: "from_recipe", Type: model.Bool()},
		{Name: "status", Type: model.Text(), NotNull: true},
		{Name: "room_id", Type: model.Text(), ServerOnly: true},
		{Name: "cost", Type: model.Int(), ServerOnly: true},
		{Name: "failure", Type: model.Text(), ServerOnly: true},
		{Name: "deducted_at", Type: model.NanoTime(), OmitEmpty: true, ServerOnly: true},
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true, ServerOnly: true},
	},
}

var UpsertSupplyUseArgsModel = model.Definition{
	Name: "upsert_supply_use_args",
	Wire: true,
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "encounter_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
		{Name: "item_id", Type: model.Text()},
		{Name: "quantity", Type: model.Float()},
		{Name: "billable", Type: model.Bool()},
	},
}

var RecentPatientsModel = model.Definition{
	Name:  "recent_patients",
	Allow: model.Read,
	Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), OmitEmpty: true}, // filter field
		{Name: "limit", Type: model.Int(), OmitEmpty: true},      // filter field
	},
}

var Definition = model.ModuleDefinition{
	Name: "clinical_encounter",
	Entities: []model.Definition{
		MedicalHistoryModel,
		EncounterSupplyModel,
		MyAgendaResultModel,
		RecentPatientsModel,
	},
}
