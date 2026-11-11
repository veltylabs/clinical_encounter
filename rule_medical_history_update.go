package clinical_encounter

import (
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/time"
)

func (next *MedicalHistory) Update(a crud.CurrentUser, tx crud.Tx, old *MedicalHistory, fields []string, m *Module) (crud.RuleAnswer, error) {
	if old.DoctorId != a.UserID {
		return crud.RuleAnswer{}, &model.FieldError{Field: "doctor_id", Err: ErrNotProfessional}
	}
	if old.Status == StatusCompleted || old.Status == StatusCancelled {
		return crud.RuleAnswer{}, ErrVisitClosed
	}

	hasStatus := false
	for _, f := range fields {
		if f == "status" {
			hasStatus = true
			break
		}
	}

	if hasStatus && next.Status == StatusCompleted {
		next.FinishedAt = time.NowNano()
		if m.supplyOn {
			return crud.ApplyAndPublish(MedicalHistoryDeductSupplies{}), nil
		}
	}

	return crud.ApplyAndPublish(), nil
}
