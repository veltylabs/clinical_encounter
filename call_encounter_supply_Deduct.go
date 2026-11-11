package clinical_encounter

import "webtyp.com/crud"

func (p *EncounterSupply) Deduct(a crud.CurrentUser, m *Module, key crud.IdempotencyKey) (crud.FieldsToSave, error) {
	visit, err := m.GetVisit(p.TenantId, p.EncounterId)
	if err != nil {
		return nil, err
	}
	m.deductLine(visit, p, visit.DoctorId)
	return crud.Fields(supplyDeductFields...), nil
}
