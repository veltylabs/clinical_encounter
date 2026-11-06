package clinical_encounter

import (
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/time"
)

// ImportVisit stores a visit from a previous system already closed, with its own id and dates: it creates it, or
// updates it when that id already exists, so importing twice leaves the same row. It requires what CreateVisit
// requires (id, tenant, patient, doctor, reason, patient name snapshot, doctor name snapshot) and a
// real moment (AttentionAt > 0). It never uses "now" for a date, never expands supply recipes, publishes no event and
// opens no transaction (the importer already runs inside one).
//
// Example: v, err := enc.ImportVisit(MedicalHistory{Id: "8000000000000000005", TenantId: t, PatientId: p,
//
//	DoctorId: d, AttentionAt: at, Reason: "control", PatientNameSnapshot: "Ana", PatientRutSnapshot: "12345678-5",
//	DoctorNameSnapshot: "Dra. Pérez"})
func (m *Module) ImportVisit(v MedicalHistory) (*MedicalHistory, error) {
	if v.TenantId == "" {
		return nil, &model.FieldError{Field: "tenant_id", Err: ErrTenantRequired}
	}
	if v.Id == "" || v.PatientId == "" || v.DoctorId == "" || v.Reason == "" ||
		v.PatientNameSnapshot == "" || v.DoctorNameSnapshot == "" {
		return nil, ErrMissingArgs
	}
	if v.AttentionAt <= 0 {
		return nil, ErrImportDate
	}

	v.Status = StatusCompleted
	if v.StartedAt == 0 {
		v.StartedAt = time.Nano(int64(v.AttentionAt) * 1e9)
	}
	if v.FinishedAt == 0 {
		v.FinishedAt = v.StartedAt
	}
	v.UpdatedAt = time.NowNano()

	existing, err := m.GetVisit(v.TenantId, v.Id)
	if err == nil {
		if existing.PatientId != v.PatientId {
			return nil, ErrIdTaken
		}
		if err := m.db.Update(&v, orm.Eq(MedicalHistory_.Id, v.Id), orm.Eq(MedicalHistory_.TenantId, v.TenantId)); err != nil {
			return nil, err
		}
		return &v, nil
	}

	if err := m.db.Create(&v); err != nil {
		return nil, err
	}
	return &v, nil
}
