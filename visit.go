package clinical_encounter

import (
	"webtyp.com/crud"
	"webtyp.com/storage"
	"webtyp.com/events"
	"webtyp.com/model"
	"webtyp.com/time"
)

func (m *Module) CreateVisit(args CreateVisitArgs) (*MedicalHistory, error) {
	if args.TenantId == "" {
		return nil, &model.FieldError{Field: "tenant_id", Err: ErrTenantRequired}
	}
	if args.Id == "" || args.PatientId == "" || args.DoctorId == "" || args.Reason == "" ||
		args.PatientNameSnapshot == "" || args.DoctorNameSnapshot == "" {
		return nil, ErrMissingArgs
	}

	if existing, err := m.GetVisit(args.TenantId, args.Id); err == nil {
		if existing.PatientId == args.PatientId {
			return existing, nil
		}
		return nil, ErrIdTaken
	} else if model.KindOf(err) != model.ErrorNotFound {
		return nil, err
	}

	// attention_at es model.Int() en MedicalHistoryModel, no un widget input.*
	// (ver el comentario del campo en model.go: el estado cambia a través del FSM,
	// el tiempo de atención es proporcionado por la máquina, no escrito por el doctor) — así que
	// webtyp/form nunca lo renderiza y todo envío de "nueva ficha" desde un crudview
	// llega aquí con AttentionAt == 0. Un encuentro clínico se crea
	// porque está ocurriendo ahora, por lo que toma el valor por defecto en lugar de
	// rechazar — rechazar haría que el flujo genérico de creación fallara permanentemente.
	if args.AttentionAt == 0 {
		args.AttentionAt = int64(time.NowUnix())
	}

	record := &MedicalHistory{
		Id:                      args.Id,
		TenantId:                args.TenantId,
		PatientId:               args.PatientId,
		DoctorId:                args.DoctorId,
		ReservationId:           args.ReservationId,
		ServiceId:               args.ServiceId,
		Status:                  StatusCreated,
		AttentionAt:             time.Unix(args.AttentionAt),
		Reason:                  args.Reason,
		Diagnostic:              args.Diagnostic,
		Prescription:            args.Prescription,
		PatientNameSnapshot:     args.PatientNameSnapshot,
		PatientRutSnapshot:      args.PatientRutSnapshot,
		DoctorNameSnapshot:      args.DoctorNameSnapshot,
		DoctorSpecialtySnapshot: args.DoctorSpecialtySnapshot,
		UpdatedAt:               time.NowNano(),
	}

	if err := m.db.Create(record); err != nil {
		return nil, err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: "clinical_encounter.medical_history.created", Payload: record})
	}
	if m.supplyOn && record.ServiceId != "" {
		itemIds, quantities, err := m.recipes.ServiceRecipe(record.TenantId, record.ServiceId)
		if err != nil {
			return record, err
		}
		for i, itemId := range itemIds {
			line := &EncounterSupply{
				Id:          m.ids.NewID(),
				TenantId:    record.TenantId,
				EncounterId: record.Id,
				ItemId:      itemId,
				Quantity:    quantities[i],
				Billable:    false,
				FromRecipe:  true,
				Status:      SupplyPlanned,
				UpdatedAt:   time.NowNano(),
			}
			if err := m.db.Create(line); err != nil {
				return record, err
			}
		}
	}
	return record, nil
}



func (m *Module) GetVisit(tenantId, id string) (*MedicalHistory, error) {
	if tenantId == "" {
		return nil, model.Invalid("tenant_id is required")
	}
	var rec MedicalHistory
	err := m.store.ReadOne(&rec, id)
	if err != nil {
		if model.KindOf(err) == model.ErrorNotFound {
			return nil, model.NotFound("visit not found")
		}
		return nil, err
	}
    if tenantId != "" && rec.TenantId != tenantId {
        return nil, model.NotFound("visit not found")
    }
	return &rec, nil
}

func (m *Module) ListVisitsByPatient(tenantId, patientId string) ([]*MedicalHistory, error) {
	if tenantId == "" {
		return nil, model.Invalid("tenant_id is required")
	}
	var results []*MedicalHistory
	err := m.store.ReadAll(func() model.Model { return &MedicalHistory{} }, crud.Query{Conditions: []storage.Condition{storage.Eq("patient_id", patientId)}, Order: []storage.Order{storage.Desc("attention_at")}}, func(rec model.Model) {
		results = append(results, rec.(*MedicalHistory))
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (m *Module) ListVisitsPaged(tenantId, patientId, after string, limit int64) ([]*MedicalHistory, error) {
	if tenantId == "" {
		return nil, model.Invalid("tenant_id is required")
	}
    q := crud.Query{Limit: limit}
    if patientId != "" {
        q.Conditions = append(q.Conditions, storage.Eq("patient_id", patientId))
    }
    if after != "" {
        q.After = after
    }
	var results []*MedicalHistory
	err := m.store.ReadAll(func() model.Model { return &MedicalHistory{} }, q, func(rec model.Model) {
		results = append(results, rec.(*MedicalHistory))
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}
