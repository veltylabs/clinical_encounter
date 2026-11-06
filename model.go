package clinical_encounter

import (
	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/model"
)

var MedicalHistoryModel = model.Definition{
	Name: "medical_history",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
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
		{Name: "patient_rut_snapshot", Type: model.Text()}, // el documento que tenga el paciente (RUT, pasaporte, extranjero) o vacío
		{Name: "doctor_name_snapshot", Type: input.Text(), Label: "doctor name", NotNull: true},
		{Name: "doctor_specialty_snapshot", Type: input.Text()},
		{Name: "updated_at", Type: model.NanoTime(), NotNull: true},
	},
}

var CreateVisitArgsModel = model.Definition{
	Name: "create_visit_args",
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

var GetVisitArgsModel = model.Definition{
	Name: "get_visit_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

var ListVisitsArgsModel = model.Definition{
	Name: "list_visits_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "patient_id", Type: model.Text()},
	},
}

var ListRecentPatientsArgsModel = model.Definition{
	Name: "list_recent_patients_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "limit", Type: model.Int()},
	},
}

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

func statusSelectInput() input.Input {
	inp := input.Select()
	if s, ok := inp.(interface{ SetOptions(...fmt.KeyValue) }); ok {
		s.SetOptions(
			fmt.KeyValue{Key: StatusCreated, Value: "Agendado"},
			fmt.KeyValue{Key: StatusArrived, Value: "En recepción"},
			fmt.KeyValue{Key: StatusTriaged, Value: "En triaje"},
			fmt.KeyValue{Key: StatusInProgress, Value: "En atención"},
			fmt.KeyValue{Key: StatusCompleted, Value: "Completado"},
			fmt.KeyValue{Key: StatusCancelled, Value: "Cancelado"},
		)
	}
	return inp
}

const ErrCannotRead model.Invalid = "clinical_encounter: cannot read the request"

const (
	ErrNotFound    model.NotFound = "visit not found"
	ErrMissingArgs model.Invalid  = "missing required arguments"
	ErrImportDate  model.Invalid  = "an imported visit needs the date it happened"
	ErrIdTaken     model.Conflict = "visit id already used by another patient"
)

const (
	TopicVisitCreated = "clinical_encounter.visit.created"
)

var EncounterSupplyModel = model.Definition{
	Name: "encounter_supply",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "encounter_id", Type: model.Text(), Ref: &MedicalHistoryModel, DB: &model.FieldDB{RefColumn: "id"}, NotNull: true},
		{Name: "item_id", Type: model.Text(), NotNull: true},
		{Name: "quantity", Type: input.Decimal(), Label: "Quantity", NotNull: true},
		{Name: "billable", Type: input.Checkbox(), Label: "Billable"},
		{Name: "from_recipe", Type: model.Bool()},
		{Name: "status", Type: model.Text(), NotNull: true},
		{Name: "room_id", Type: model.Text()},
		{Name: "cost", Type: model.Int()},
		{Name: "failure", Type: model.Text()},
		{Name: "deducted_at", Type: model.NanoTime(), OmitEmpty: true},
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true},
	},
}

var CompleteVisitArgsModel = model.Definition{
	Name: "complete_visit_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

var UpsertSupplyUseArgsModel = model.Definition{
	Name: "upsert_supply_use_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "encounter_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
		{Name: "item_id", Type: model.Text()},
		{Name: "quantity", Type: model.Float()},
		{Name: "billable", Type: model.Bool()},
	},
}

var SupplyUseRefArgsModel = model.Definition{
	Name: "supply_use_ref_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

var ListSupplyUseArgsModel = model.Definition{
	Name: "list_supply_use_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "encounter_id", Type: model.Text()},
	},
}

var ListPendingSupplyArgsModel = model.Definition{
	Name: "list_pending_supply_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
	},
}
