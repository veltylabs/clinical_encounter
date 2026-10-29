package clinical_encounter

import (
	"webtyp.com/model"
	"webtyp.com/events"
	"webtyp.com/orm"
	"webtyp.com/time"
)

func (m *Module) CreateVisit(args CreateVisitArgs) (*MedicalHistory, error) {
	if args.TenantId == "" {
		return nil, &model.FieldError{Field: "tenant_id", Err: ErrTenantRequired}
	}
	if args.Id == "" || args.PatientId == "" || args.DoctorId == "" || args.Reason == "" ||
		args.PatientNameSnapshot == "" || args.PatientRutSnapshot == "" || args.DoctorNameSnapshot == "" {
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

	// attention_at is model.Int() in MedicalHistoryModel, not an input.*
	// widget (see model.go's field comment: status changes through the FSM,
	// attention time is machine-supplied, not typed by the doctor) — so
	// webtyp/form never renders it and every crudview-driven "new ficha"
	// submission reaches here with AttentionAt == 0. A clinical encounter is
	// created because it is happening now, so this defaults instead of
	// rejecting — rejecting made the generic create flow permanently fail.
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
		m.pub.Publish(events.Event{Topic: TopicVisitCreated, Payload: record})
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
		return nil, &model.FieldError{Field: "tenant_id", Err: ErrTenantRequired}
	}
	var rec MedicalHistory
	qb := m.db.Query(&rec).Where(MedicalHistory_.Id).Eq(id).Where(MedicalHistory_.TenantId).Eq(tenantId)
	_, err := ReadOneMedicalHistory(qb, &rec)
	if err != nil {
		if orm.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &rec, nil
}

func (m *Module) ListVisitsByPatient(tenantId, patientId string) ([]*MedicalHistory, error) {
	if tenantId == "" {
		return nil, &model.FieldError{Field: "tenant_id", Err: ErrTenantRequired}
	}
	var rec MedicalHistory
	qb := m.db.Query(&rec).
		Where(MedicalHistory_.PatientId).Eq(patientId).
		Where(MedicalHistory_.TenantId).Eq(tenantId).
		OrderBy(MedicalHistory_.AttentionAt).Desc()
	results, err := ReadAllMedicalHistory(qb)
	if err != nil {
		return nil, err
	}
	return results, nil // ReadAllX returns []*MedicalHistory directly (same as service_catalog) — never dereference it
}
