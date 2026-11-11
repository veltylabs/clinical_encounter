package clinical_encounter

import "webtyp.com/crud"

func (p *EncounterSupply) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	if !m.supplyOn {
		return crud.RuleAnswer{}, ErrSupplyUseOff
	}
	if p.Quantity <= 0 {
		return crud.RuleAnswer{}, ErrQuantity
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
	p.Status = SupplyPlanned
	return crud.ApplyAndPublish(), nil
}
