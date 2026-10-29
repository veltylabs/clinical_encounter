package clinical_encounter

import (
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/time"
)









func (m *Module) ListSupplyUse(tenantId, encounterId string) ([]*EncounterSupply, error) {
	if !m.supplyOn {
		return nil, ErrSupplyUseOff
	}
	if tenantId == "" {
		return nil, &model.FieldError{Field: "tenant_id", Err: ErrTenantRequired}
	}
	var rec EncounterSupply
	qb := m.db.Query(&rec).
		Where(EncounterSupply_.TenantId).Eq(tenantId).
		Where(EncounterSupply_.EncounterId).Eq(encounterId)
	return ReadAllEncounterSupply(qb)
}

func (m *Module) UpsertSupplyUse(args UpsertSupplyUseArgs) (*EncounterSupply, error) {
	if !m.supplyOn {
		return nil, ErrSupplyUseOff
	}
	if args.TenantId == "" {
		return nil, &model.FieldError{Field: "tenant_id", Err: ErrTenantRequired}
	}
	if args.Quantity <= 0 {
		return nil, &model.FieldError{Field: "quantity", Err: ErrQuantity}
	}

	visit, err := m.GetVisit(args.TenantId, args.EncounterId)
	if err != nil {
		return nil, err
	}
	if visit.Status == StatusCompleted || visit.Status == StatusCancelled {
		return nil, ErrVisitClosed
	}

	if args.Id == "" {
		line := &EncounterSupply{
			Id:          m.ids.NewID(),
			TenantId:    args.TenantId,
			EncounterId: args.EncounterId,
			ItemId:      args.ItemId,
			Quantity:    args.Quantity,
			Billable:    args.Billable,
			FromRecipe:  false,
			Status:      SupplyPlanned,
			UpdatedAt:   time.NowNano(),
		}
		if err := m.db.Create(line); err != nil {
			return nil, err
		}
		return line, nil
	}

	var existing EncounterSupply
	qb := m.db.Query(&existing).
		Where(EncounterSupply_.Id).Eq(args.Id).
		Where(EncounterSupply_.TenantId).Eq(args.TenantId)
	_, err = ReadOneEncounterSupply(qb, &existing)
	if err != nil {
		if orm.IsNotFound(err) {
			return nil, ErrSupplyNotFound
		}
		return nil, err
	}

	if existing.Status == SupplyDeducted {
		return nil, ErrLineDeducted
	}

	existing.ItemId = args.ItemId
	existing.Quantity = args.Quantity
	existing.Billable = args.Billable
	existing.UpdatedAt = time.NowNano()

	if err := m.db.Update(&existing, orm.Eq(EncounterSupply_.Id, existing.Id), orm.Eq(EncounterSupply_.TenantId, existing.TenantId)); err != nil {
		return nil, err
	}
	return &existing, nil
}

func (m *Module) DeleteSupplyUse(tenantId, id string) error {
	if !m.supplyOn {
		return ErrSupplyUseOff
	}
	if tenantId == "" {
		return &model.FieldError{Field: "tenant_id", Err: ErrTenantRequired}
	}

	var existing EncounterSupply
	qb := m.db.Query(&existing).
		Where(EncounterSupply_.Id).Eq(id).
		Where(EncounterSupply_.TenantId).Eq(tenantId)
	_, err := ReadOneEncounterSupply(qb, &existing)
	if err != nil {
		if orm.IsNotFound(err) {
			return ErrSupplyNotFound
		}
		return err
	}

	visit, err := m.GetVisit(tenantId, existing.EncounterId)
	if err != nil {
		return err
	}
	if visit.Status == StatusCompleted || visit.Status == StatusCancelled {
		return ErrVisitClosed
	}

	return m.db.Delete(&EncounterSupply{}, orm.Eq(EncounterSupply_.Id, id), orm.Eq(EncounterSupply_.TenantId, tenantId))
}

func (m *Module) ListPendingSupply(tenantId string) ([]*EncounterSupply, error) {
	if !m.supplyOn {
		return nil, ErrSupplyUseOff
	}
	if tenantId == "" {
		return nil, &model.FieldError{Field: "tenant_id", Err: ErrTenantRequired}
	}
	var rec EncounterSupply
	qb := m.db.Query(&rec).
		Where(EncounterSupply_.TenantId).Eq(tenantId).
		Where(EncounterSupply_.Status).Eq(SupplyPendingDeduction)
	return ReadAllEncounterSupply(qb)
}

func (m *Module) RetrySupplyUse(tenantId, id, by string) (*EncounterSupply, error) {
	if !m.supplyOn {
		return nil, ErrSupplyUseOff
	}
	if tenantId == "" {
		return nil, &model.FieldError{Field: "tenant_id", Err: ErrTenantRequired}
	}

	var existing EncounterSupply
	qb := m.db.Query(&existing).
		Where(EncounterSupply_.Id).Eq(id).
		Where(EncounterSupply_.TenantId).Eq(tenantId)
	_, err := ReadOneEncounterSupply(qb, &existing)
	if err != nil {
		if orm.IsNotFound(err) {
			return nil, ErrSupplyNotFound
		}
		return nil, err
	}

	if existing.Status != SupplyPendingDeduction {
		return nil, ErrLineNotPending
	}

	visit, err := m.GetVisit(tenantId, existing.EncounterId)
	if err != nil {
		return nil, err
	}

	if err := m.deductLine(visit, &existing, by); err != nil {
		return nil, err
	}

	return &existing, nil
}
