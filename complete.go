package clinical_encounter

import (
	"webtyp.com/model"
	"webtyp.com/events"
	"webtyp.com/msgtype"
	"webtyp.com/notify"
	"webtyp.com/orm"
	"webtyp.com/time"
)

const SupplyDocType = "medical_history"

func (m *Module) CompleteVisit(tenantId, id, by string) (*MedicalHistory, []*EncounterSupply, error) {
	if tenantId == "" {
		return nil, nil, &model.FieldError{Field: "tenant_id", Err: ErrTenantRequired}
	}

	visit, err := m.GetVisit(tenantId, id)
	if err != nil {
		return nil, nil, err
	}

	if visit.Status == StatusCompleted || visit.Status == StatusCancelled {
		return nil, nil, ErrVisitClosed
	}

	visit.Status = StatusCompleted
	visit.FinishedAt = time.NowNano()
	visit.UpdatedAt = time.NowNano()

	if err := m.db.Update(visit, orm.Eq(MedicalHistory_.Id, visit.Id), orm.Eq(MedicalHistory_.TenantId, visit.TenantId)); err != nil {
		return nil, nil, err
	}

	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicVisitCompleted, Payload: visit})
	}

	if !m.supplyOn {
		return visit, nil, nil
	}

	lines, err := m.ListSupplyUse(tenantId, id)
	if err != nil {
		return nil, nil, err
	}

	pendingCount := 0

	for _, line := range lines {
		if line.Status == SupplyPlanned {
			if err := m.deductLine(visit, line, by); err != nil {
				return nil, nil, err
			}
			if line.Status == SupplyPendingDeduction {
				pendingCount++
			}
		}
	}

	if pendingCount > 0 && m.pub != nil {
		notify.Raise(m.pub, notify.Alert{
			Code:    AlertSupplyPending,
			Scope:   m.suppliesArea,
			Level:   msgtype.Warning,
			Subject: alertSubject,
			Body:    alertBody,
			Source:  notify.Ref{Type: SupplyDocType, ID: visit.Id},
		})
	}

	return visit, lines, nil
}

const msgNoShift = "the professional has no shift in a room at the attention time"

func (m *Module) deductLine(v *MedicalHistory, l *EncounterSupply, by string) error {
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
	l.UpdatedAt = time.NowNano()
	return m.db.Update(l, orm.Eq(EncounterSupply_.Id, l.Id), orm.Eq(EncounterSupply_.TenantId, l.TenantId))
}
