package clinical_encounter

import (
	ab "github.com/veltylabs/appointment_booking"
	"github.com/veltylabs/patient_directory"
	"github.com/veltylabs/service_catalog"
	"github.com/veltylabs/staff_manager"
	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/model"
)

var MedicalHistoryModel = model.Definition{
	Name: "medical_history",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "patient_id", Type: input.Text(), Ref: &patientdirectory.PatientModel, DB: &model.FieldDB{RefColumn: "id"}, NotNull: true},
		{Name: "doctor_id", Type: input.Text(), Ref: &staffmanager.StaffMemberModel, DB: &model.FieldDB{RefColumn: "id"}, NotNull: true},
		{Name: "reservation_id", Type: input.Text(), Ref: &ab.ReservationModel, DB: &model.FieldDB{RefColumn: "id"}},
		{Name: "service_id", Type: input.Text(), Ref: &servicecatalog.ServiceModel, DB: &model.FieldDB{RefColumn: "id"}},
		// status: los valores válidos son ÚNICAMENTE las constantes Status* en const.go — los literales
		// viven solo en esas constantes y en ningún otro lugar (regla anti magic-string; ver la revisión de service_catalog).
		{Name: "status", Type: statusSelectInput(), NotNull: true, Item: model.ItemDescription},
		{Name: "attention_at", Type: model.UnixTime(), NotNull: true, Item: model.ItemBadge},
		{Name: "reason", Type: input.Textarea(), NotNull: true, Item: model.ItemLabel},
		{Name: "diagnostic", Type: input.Textarea()},
		{Name: "prescription", Type: input.Textarea()},
		{Name: "cie10_code", Type: input.Text()},
		{Name: "started_at", Type: model.NanoTime()},
		{Name: "finished_at", Type: model.NanoTime()},
		{Name: "patient_name_snapshot", Type: input.Text(), Label: "patient name", NotNull: true},
		{Name: "patient_rut_snapshot", Type: input.Rut(), NotNull: true},
		{Name: "doctor_name_snapshot", Type: input.Text(), Label: "doctor name", NotNull: true},
		{Name: "doctor_specialty_snapshot", Type: input.Text()},
		{Name: "updated_at", Type: model.NanoTime(), NotNull: true},
	},
}

// NotNull refleja exactamente el conjunto de argumentos requeridos que CreateVisit exige (visit.go) —
// el Definition es el único lugar donde se declara el contrato; las validaciones manuales del servicio
// son un duplicado de defensa en profundidad del mismo conjunto, nunca uno distinto.
// Ejemplo: CreateVisitArgsModel.Name == "create_visit_args"
var CreateVisitArgsModel = model.Definition{
	Name: "create_visit_args",
	// id lo genera quien llama.
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, NotNull: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "patient_id", Type: input.Text(), Ref: &patientdirectory.PatientModel, DB: &model.FieldDB{RefColumn: "id"}, NotNull: true},
		{Name: "doctor_id", Type: input.Text(), Ref: &staffmanager.StaffMemberModel, DB: &model.FieldDB{RefColumn: "id"}, NotNull: true},
		{Name: "attention_at", Type: input.Number(), NotNull: true},
		{Name: "reason", Type: input.Text(), NotNull: true},
		{Name: "patient_name_snapshot", Type: input.Text(), Label: "patient name", NotNull: true},
		{Name: "patient_rut_snapshot", Type: input.Text(), NotNull: true},
		{Name: "doctor_name_snapshot", Type: input.Text(), Label: "doctor name", NotNull: true},
		{Name: "reservation_id", Type: input.Text(), Ref: &ab.ReservationModel, DB: &model.FieldDB{RefColumn: "id"}},
		{Name: "service_id", Type: input.Text(), Ref: &servicecatalog.ServiceModel, DB: &model.FieldDB{RefColumn: "id"}},
		{Name: "diagnostic", Type: input.Textarea()},
		{Name: "prescription", Type: input.Textarea()},
		{Name: "doctor_specialty_snapshot", Type: input.Text()},
	},
}

// GetVisitArgsModel y ListVisitsArgsModel son solo de transporte (Field.DB es nil en todos los campos) —
// ver ops.go para las operaciones que los consumen.
// Ejemplo: GetVisitArgsModel.Name == "get_visit_args"
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

// ListRecentPatientsArgsModel es solo de transporte, sin campos de base de datos — el picker
// que lista QUÉ pacientes ya tienen ficha (ver ListRecentPatients en
// patients.go) no toma filtro propio hoy en día; Limit existe para que un llamador
// pueda limitar el escaneo sin un cambio de esquema una vez que el volumen lo requiera.
// Ejemplo: ListRecentPatientsArgsModel.Name == "list_recent_patients_args"
var ListRecentPatientsArgsModel = model.Definition{
	Name: "list_recent_patients_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "limit", Type: model.Int()},
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

// 
const ErrCannotRead model.Invalid = "clinical_encounter: cannot read the request"







const (
	ErrNotFound model.NotFound = "visit not found"
	ErrMissingArgs model.Invalid = "missing required arguments"
	ErrIdTaken model.Conflict = "visit id already used by another patient"
)

// Topics de eventos de dominio — <módulo>.<entidad>.<verbo-en-pasado>, los datos de tenant/id van en
// el payload, nunca en el nombre del topic (no aplica aquí — ver la nota "no TenantId field" en ARCHITECTURE.md).
// Ejemplo: TopicVisitCreated == "clinical_encounter.visit.created"
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
