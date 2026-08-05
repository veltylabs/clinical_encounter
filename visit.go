package clinical_encounter

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/time"
)

func (m *Module) CreateVisit(args CreateVisitArgs) (*MedicalHistory, error) {
	if args.PatientID == "" || args.DoctorID == "" || args.Reason == "" ||
		args.PatientNameSnapshot == "" || args.PatientRutSnapshot == "" || args.DoctorNameSnapshot == "" {
		return nil, fmt.Err("missing", "required", "arguments")
	}

	if args.AttentionAt == 0 {
		return nil, fmt.Err("missing", "attention_at")
	}

	record := &MedicalHistory{
		ID:                      m.UID.NewID(),
		PatientID:               args.PatientID,
		DoctorID:                args.DoctorID,
		ReservationID:           args.ReservationID,
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

	if err := m.DB.Create(record); err != nil {
		return nil, err
	}

	return record, nil
}
