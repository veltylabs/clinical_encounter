package clinical_encounter

import "webtyp.com/crud"

func (next *EncounterSupply) Update(a crud.CurrentUser, tx crud.Tx, old *EncounterSupply, fields []string, m *Module) (crud.RuleAnswer, error) {
    hasStatus := false
	for _, f := range fields {
		if f == "status" {
			hasStatus = true
			break
		}
	}

    if hasStatus && next.Status == SupplyPendingDeduction {
        if old.Status != SupplyPendingDeduction {
            return crud.RuleAnswer{}, ErrLineNotPending
        }
        return crud.ApplyAndPublish(EncounterSupplyDeduct{}), nil
    }

	if next.Quantity <= 0 {
		return crud.RuleAnswer{}, ErrQuantity
	}
	if old.Status != SupplyPlanned {
		return crud.RuleAnswer{}, ErrLineDeducted
	}
	if old.EncounterId != "" {
		visit, err := m.GetVisit(old.TenantId, old.EncounterId)
		if err != nil {
			return crud.RuleAnswer{}, err
		}
		if visit.Status == StatusCompleted || visit.Status == StatusCancelled {
			return crud.RuleAnswer{}, ErrVisitClosed
		}
	}

	return crud.ApplyAndPublish(), nil
}
