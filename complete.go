package clinical_encounter

import (
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/time"
)

const SupplyDocType = "medical_history"

const msgNoShift = "the professional has no shift in a room at the attention time"

var supplyDeductFields = []string{"status", "cost", "failure", "room_id", "deducted_at"}

func (m *Module) deductLine(v *MedicalHistory, l *EncounterSupply, by string) {
	roomID, found, err := m.rooms.RoomOfOccupant(v.TenantId, v.DoctorId, v.AttentionAt)
	switch {
	case err != nil:
		l.Status, l.Failure = SupplyPendingDeduction, err.Error()
	case !found:
		l.Status, l.Failure = SupplyPendingDeduction, msgNoShift
	default:
		cost, useErr := m.supplies.UseInRoom(v.TenantId, roomID, l.ItemId, l.Quantity, SupplyDocType, v.Id, by)
		l.RoomId = roomID
		if useErr != nil && cost == 0 {
			l.Status, l.Failure = SupplyPendingDeduction, useErr.Error()
		} else {
			l.Status, l.Cost, l.Failure, l.DeductedAt = SupplyDeducted, cost, "", time.NowNano()
		}
	}
}

func (m *Module) CompleteVisit(tenantId, id, by string) (*MedicalHistory, []*EncounterSupply, error) {
	if tenantId == "" {
		return nil, nil, model.Invalid("tenant_id is required")
	}

	var visit *MedicalHistory
	err := m.store.Atomic(func(tx crud.Tx) error {
		var existing MedicalHistory
		err := tx.ReadOne(&existing, id)
		if err != nil {
			return err
		}

		next := existing
		next.Status = StatusCompleted
		_, err = next.Update(crud.CurrentUser{UserID: by}, tx, &existing, []string{"status"}, m)
		if err != nil {
			return err
		}

		if err := tx.UpdateFields(&next, next.Id, []string{"status", "finished_at"}); err != nil {
			return err
		}
		visit = &next
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if !m.supplyOn {
		return visit, nil, nil
	}

	lines, err := m.ListSupplyUse(tenantId, id)
	return visit, lines, err
}
