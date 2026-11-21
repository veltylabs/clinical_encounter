package clinical_encounter

import (
	"webtyp.com/crud"
	"webtyp.com/storage"
)

// Search es la regla de negocio de leer medical_history: solo las fichas del paciente del filtro. Sin paciente no hay
// fichas: la lista se alcanza eligiendo un paciente, nunca trae todas (hay decenas de miles).
func (p *MedicalHistory) Search(a crud.CurrentUser, filter *MedicalHistory, m *Module) ([]storage.Condition, error) {
	return []storage.Condition{storage.Eq(MedicalHistory_.PatientId, filter.PatientId)}, nil
}
