package clinical_encounter

import (
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/time"
)

// ImportVisit guarda una visita de un sistema anterior ya cerrada, con su propio id y fechas: la crea, o
// la actualiza cuando ese id ya existe, de modo que importar dos veces deja la misma fila. Requiere lo que requiere
// CreateVisit (id, tenant, paciente, doctor, motivo, captura del nombre del paciente, captura del nombre del doctor) y un
// momento real (AttentionAt > 0). Nunca usa "ahora" para una fecha, nunca expande recetas de insumos, no publica ningún evento y
// no abre ninguna transacción (el importador ya se ejecuta dentro de una).
//
// Ejemplo: v, err := enc.ImportVisit(MedicalHistory{Id: "8000000000000000005", TenantId: t, PatientId: p,
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
