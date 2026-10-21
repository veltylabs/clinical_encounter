# 7. supply.go, complete.go, alerts.go, errors.go, ops.go
cat << 'GO_EOF' > supply.go
package clinical_encounter

import (
	"webtyp.com/orm"
	"webtyp.com/time"
)

type UpsertSupplyUseArgs struct {
	TenantId    string
	EncounterId string
	Id          string
	ItemId      string
	Quantity    float64
	Billable    bool
}

type SupplyUseRefArgs struct {
	TenantId string
	Id       string
}

type ListSupplyUseArgs struct {
	TenantId    string
	EncounterId string
}

type ListPendingSupplyArgs struct {
	TenantId string
}

func (m *Module) ListSupplyUse(tenantId, encounterId string) ([]*EncounterSupply, error) {
	if !m.supplyOn {
		return nil, ErrSupplyUseOff
	}
	if tenantId == "" {
		return nil, ErrTenantRequired
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
		return nil, ErrTenantRequired
	}
	if args.Quantity <= 0 {
		return nil, ErrQuantity
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
		return ErrTenantRequired
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
		return nil, ErrTenantRequired
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
		return nil, ErrTenantRequired
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
GO_EOF

cat << 'GO_EOF' > complete.go
package clinical_encounter

import (
	"webtyp.com/events"
	"webtyp.com/msgtype"
	"webtyp.com/notify"
	"webtyp.com/orm"
	"webtyp.com/time"
)

const SupplyDocType = "medical_history"

func (m *Module) CompleteVisit(tenantId, id, by string) (*MedicalHistory, []*EncounterSupply, error) {
	if tenantId == "" {
		return nil, nil, ErrTenantRequired
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
GO_EOF

cat << 'GO_EOF' > alerts.go
package clinical_encounter

import "webtyp.com/notify"

const (
	AlertSupplyPending = "clinical_encounter.supply.pending_deduction"
)

const (
	alertSubject = "Supplies pending deduction"
	alertBody    = "An attention was completed with supplies that could not be taken from stock. Open Supplies to retry them."
)

func AlertKinds(areaCode string) []notify.Kind {
	return []notify.Kind{
		{
			Code:        AlertSupplyPending,
			Area:        areaCode,
			Name: "Atenciones con insumos sin rebajar",
		},
	}
}
GO_EOF

cat << 'GO_EOF' > errors.go
package clinical_encounter

import (
	"strings"
	"webtyp.com/router"
)

const (
	ErrTenantRequired = domainError("tenant_id is required")
	ErrQuantity       = domainError("quantity must be greater than zero")
	ErrSupplyNotFound = domainError("supply use line not found")
	ErrVisitClosed    = domainError("cannot modify supply use of a closed visit")
	ErrSupplyUseOff   = domainError("supply use is not enabled")
	ErrLineDeducted   = domainError("cannot modify a deducted line")
	ErrLineNotPending = domainError("line is not pending deduction")
)

func writeError(ctx router.Context, err error) {
	status := statusOf(err)
	ctx.WriteStatus(status)
	// We don't expose internal errors to the client; router framework handles encoding
}

func statusOf(err error) int {
	if isDecodeError(err) {
		return 400
	}
	if e, ok := err.(domainError); ok {
		switch e {
		case ErrMissingArgs, ErrTenantRequired, ErrQuantity:
			return 400
		case ErrNotFound, ErrSupplyNotFound:
			return 404
		case ErrIdTaken, ErrVisitClosed, ErrLineDeducted, ErrLineNotPending:
			return 409
		case ErrSupplyUseOff:
			return 501
		}
	}
	return 500
}

func isDecodeError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "json:") || strings.Contains(msg, "decode") || strings.Contains(msg, "EOF")
}
GO_EOF

cat << 'PY_EOF' > patch_ops.py
import re

with open("ops.go", "r") as f:
    text = f.read()

text = text.replace('OpListRecentPatients  = "list_recent_patients"', """OpListRecentPatients  = "list_recent_patients"

	OpCompleteVisit       = "complete_visit"
	OpListSupplyUse       = "list_supply_use"
	OpUpsertSupplyUse     = "upsert_supply_use"
	OpDeleteSupplyUse     = "delete_supply_use"
	OpListPendingSupply   = "list_pending_supply"
	OpRetrySupplyUse      = "retry_supply_use"
)

const ResourceSupplyUse = "encounter_supply\"""")

text = text.replace('reg.Operation(OpListRecentPatients, m.opListRecentPatients).Requires("medical_history", model.Read).Accepts(&ListRecentPatientsArgs{})', """reg.Operation(OpListRecentPatients, m.opListRecentPatients).Requires("medical_history", model.Read).Accepts(&ListRecentPatientsArgs{})

	reg.Operation(OpCompleteVisit, m.opCompleteVisit).Requires("medical_history", model.Update).Accepts(&CompleteVisitArgs{})
	reg.Operation(OpListSupplyUse, m.opListSupplyUse).Requires(ResourceSupplyUse, model.Read).Accepts(&ListSupplyUseArgs{})
	reg.Operation(OpUpsertSupplyUse, m.opUpsertSupplyUse).Requires(ResourceSupplyUse, model.Update).Accepts(&UpsertSupplyUseArgs{})
	reg.Operation(OpDeleteSupplyUse, m.opDeleteSupplyUse).Requires(ResourceSupplyUse, model.Delete).Accepts(&SupplyUseRefArgs{})
	reg.Operation(OpListPendingSupply, m.opListPendingSupply).Requires(ResourceSupplyUse, model.Read).Accepts(&ListPendingSupplyArgs{})
	reg.Operation(OpRetrySupplyUse, m.opRetrySupplyUse).Requires(ResourceSupplyUse, model.Update).Accepts(&SupplyUseRefArgs{})""")

text = re.sub(r'if e, ok := err\.\(domainError\); ok \{[\s\S]*?ctx\.WriteStatus\(500\)\n\t\treturn\n\t\}', """writeError(ctx, err)
		return
	}""", text)

text = re.sub(r'if e, ok := err\.\(domainError\); ok && e == ErrNotFound \{[\s\S]*?ctx\.WriteStatus\(500\)\n\t\treturn\n\t\}', """writeError(ctx, err)
		return
	}""", text)

text = text.replace("""ctx.WriteStatus(500)
		return
	}""", """writeError(ctx, err)
		return
	}""")

text = text.replace('record, err := m.GetVisit(args.Id)', 'record, err := m.GetVisit(args.TenantId, args.Id)')
text = text.replace('records, err := m.ListVisitsByPatient(args.PatientId)', 'records, err := m.ListVisitsByPatient(args.TenantId, args.PatientId)')
text = text.replace('records, err := m.ListRecentPatients(int(args.Limit))', 'records, err := m.ListRecentPatients(args.TenantId, int(args.Limit))')

text += """
type CompleteVisitArgs struct {
	TenantId string
	Id       string
}

func (m *Module) opCompleteVisit(ctx router.Context) {
	var args CompleteVisitArgs
	if err := ctx.Decode(&args); err != nil {
		writeError(ctx, err)
		return
	}
	record, _, err := m.CompleteVisit(args.TenantId, args.Id, ctx.UserID())
	if err != nil {
		writeError(ctx, err)
		return
	}
	if err := ctx.Encode(record); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opListSupplyUse(ctx router.Context) {
	var args ListSupplyUseArgs
	if err := ctx.Decode(&args); err != nil {
		writeError(ctx, err)
		return
	}
	records, err := m.ListSupplyUse(args.TenantId, args.EncounterId)
	if err != nil {
		writeError(ctx, err)
		return
	}
	list := make(EncounterSupplyList, len(records))
	for i, r := range records {
		list[i] = r
	}
	if err := ctx.Encode(&list); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opUpsertSupplyUse(ctx router.Context) {
	var args UpsertSupplyUseArgs
	if err := ctx.Decode(&args); err != nil {
		writeError(ctx, err)
		return
	}
	record, err := m.UpsertSupplyUse(args)
	if err != nil {
		writeError(ctx, err)
		return
	}
	if err := ctx.Encode(record); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opDeleteSupplyUse(ctx router.Context) {
	var args SupplyUseRefArgs
	if err := ctx.Decode(&args); err != nil {
		writeError(ctx, err)
		return
	}
	err := m.DeleteSupplyUse(args.TenantId, args.Id)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.WriteStatus(200)
}

func (m *Module) opListPendingSupply(ctx router.Context) {
	var args ListPendingSupplyArgs
	if err := ctx.Decode(&args); err != nil {
		writeError(ctx, err)
		return
	}
	records, err := m.ListPendingSupply(args.TenantId)
	if err != nil {
		writeError(ctx, err)
		return
	}
	list := make(EncounterSupplyList, len(records))
	for i, r := range records {
		list[i] = r
	}
	if err := ctx.Encode(&list); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opRetrySupplyUse(ctx router.Context) {
	var args SupplyUseRefArgs
	if err := ctx.Decode(&args); err != nil {
		writeError(ctx, err)
		return
	}
	record, err := m.RetrySupplyUse(args.TenantId, args.Id, ctx.UserID())
	if err != nil {
		writeError(ctx, err)
		return
	}
	if err := ctx.Encode(record); err != nil {
		ctx.WriteStatus(500)
	}
}
"""

with open("ops.go", "w") as f:
    f.write(text)
PY_EOF
python3 patch_ops.py
