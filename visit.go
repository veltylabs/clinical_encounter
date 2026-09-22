package clinical_encounter

import (
	"webtyp.com/events"
	"webtyp.com/orm"
	"webtyp.com/time"
)

func (m *Module) CreateVisit(args CreateVisitArgs) (*MedicalHistory, error) {
	if args.PatientId == "" || args.DoctorId == "" || args.Reason == "" ||
		args.PatientNameSnapshot == "" || args.PatientRutSnapshot == "" || args.DoctorNameSnapshot == "" {
		return nil, ErrMissingArgs
	}
	// attention_at is model.Int() in MedicalHistoryModel, not an input.*
	// widget (see model.go's field comment: status changes through the FSM,
	// attention time is machine-supplied, not typed by the doctor) — so
	// webtyp/form never renders it and every crudview-driven "new ficha"
	// submission reaches here with AttentionAt == 0. A clinical encounter is
	// created because it is happening now, so this defaults instead of
	// rejecting — rejecting made the generic create flow permanently fail.
	if args.AttentionAt == 0 {
		args.AttentionAt = time.Now()
	}

	record := &MedicalHistory{
		Id:                      m.ids.NewID(),
		PatientId:               args.PatientId,
		DoctorId:                args.DoctorId,
		ReservationId:           args.ReservationId,
		Status:                  StatusCreated,
		AttentionAt:             args.AttentionAt,
		Reason:                  args.Reason,
		Diagnostic:              args.Diagnostic,
		Prescription:            args.Prescription,
		PatientNameSnapshot:     args.PatientNameSnapshot,
		PatientRutSnapshot:      args.PatientRutSnapshot,
		DoctorNameSnapshot:      args.DoctorNameSnapshot,
		DoctorSpecialtySnapshot: args.DoctorSpecialtySnapshot,
		UpdatedAt:               time.Now(),
	}

	if err := m.db.Create(record); err != nil {
		return nil, err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicVisitCreated, Payload: record})
	}
	return record, nil
}

func (m *Module) GetVisit(id string) (*MedicalHistory, error) {
	var rec MedicalHistory
	qb := m.db.Query(&rec).Where(MedicalHistory_.Id).Eq(id)
	_, err := ReadOneMedicalHistory(qb, &rec)
	if err != nil {
		if err == orm.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &rec, nil
}

func (m *Module) ListVisitsByPatient(patientId string) ([]*MedicalHistory, error) {
	var rec MedicalHistory
	qb := m.db.Query(&rec).
		Where(MedicalHistory_.PatientId).Eq(patientId).
		OrderBy(MedicalHistory_.AttentionAt).Desc()
	results, err := ReadAllMedicalHistory(qb)
	if err != nil {
		return nil, err
	}
	return results, nil // ReadAllX returns []*MedicalHistory directly (same as item_catalog) — never dereference it
}
