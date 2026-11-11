package clinical_encounter

import "webtyp.com/crud"

func (p *EncounterSupply) Delete(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	if p.Status != SupplyPlanned {
		return crud.RuleAnswer{}, ErrLineDeducted
	}
	if p.EncounterId != "" {
		visit, err := m.GetVisit(p.TenantId, p.EncounterId)
		if err != nil {
			return crud.RuleAnswer{}, err
		}
		if visit.Status == StatusCompleted || visit.Status == StatusCancelled {
			return crud.RuleAnswer{}, ErrVisitClosed
		}
	}
	return crud.ApplyAndPublish(), nil
}
