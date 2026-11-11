package clinical_encounter
import (
	"webtyp.com/crud"
	"webtyp.com/msgtype"
	"webtyp.com/notify"
)
func (p *MedicalHistory) DeductSupplies(a crud.CurrentUser, m *Module, key crud.IdempotencyKey) (crud.FieldsToSave, error) {
	lines, err := m.ListSupplyUse(p.TenantId, p.Id)
	if err != nil {
		return nil, err
	}
	pending := 0
	for _, l := range lines {
		if l.Status != SupplyPlanned {
			continue
		}
		m.deductLine(p, l, p.DoctorId)
		if err := m.store.UpdateFields(l, l.Id, supplyDeductFields); err != nil {
			return nil, err
		}
		if l.Status == SupplyPendingDeduction {
			pending++
		}
	}
	if pending > 0 && m.pub != nil {
		notify.Raise(m.pub, notify.Alert{
			Code:    AlertSupplyPending,
			Scope:   m.suppliesArea,
			Level:   msgtype.Warning,
			Subject: alertSubject,
			Body:    alertBody,
			Source:  notify.Ref{Type: SupplyDocType, ID: p.Id},
		})
	}
	return nil, nil
}
