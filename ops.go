package clinical_encounter

import (
	"webtyp.com/view"
	"webtyp.com/rpc"
	"webtyp.com/model"
	"webtyp.com/router"
)

const (
	OpCreateVisit         = "create_visit"
	OpGetVisit            = "get_medical_history"
	OpListVisitsByPatient = "list_medical_history"
	OpListRecentPatients  = "list_recent_patients"

	OpCompleteVisit       = "complete_visit"
	OpListSupplyUse       = "list_supply_use"
	OpUpsertSupplyUse     = "upsert_supply_use"
	OpDeleteSupplyUse     = "delete_supply_use"
	OpListPendingSupply   = "list_pending_supply"
	OpRetrySupplyUse      = "retry_supply_use"
)



func (m *Module) MountOperations(reg router.OperationRegistry) {
	view.MountSave(reg, OpCreateVisit, func(rec *MedicalHistory) error {
		out, err := m.CreateVisit(CreateVisitArgs{
			Id:                      rec.Id,
			TenantId:                rec.TenantId,
			PatientId:               rec.PatientId,
			DoctorId:                rec.DoctorId,
			AttentionAt:             int64(rec.AttentionAt),
			Reason:                  rec.Reason,
			PatientNameSnapshot:     rec.PatientNameSnapshot,
			PatientRutSnapshot:      rec.PatientRutSnapshot,
			DoctorNameSnapshot:      rec.DoctorNameSnapshot,
			DoctorSpecialtySnapshot: rec.DoctorSpecialtySnapshot,
			ReservationId:           rec.ReservationId,
			ServiceId:               rec.ServiceId,
			Diagnostic:              rec.Diagnostic,
			Prescription:            rec.Prescription,
		})
		if err != nil {
			return err
		}
		*rec = *out
		return nil
	}).Requires(model.ResourceOf(&MedicalHistoryModel), model.Create)
	reg.Operation(OpGetVisit, m.opGetVisit).Requires(model.ResourceOf(&MedicalHistoryModel), model.Read).Accepts(&GetVisitArgs{})
	reg.Operation(OpListVisitsByPatient, m.opListVisits).Requires(model.ResourceOf(&MedicalHistoryModel), model.Read).Accepts(&ListVisitsArgs{})
	reg.Operation(OpListRecentPatients, m.opListRecentPatients).Requires(model.ResourceOf(&MedicalHistoryModel), model.Read).Accepts(&ListRecentPatientsArgs{})

	reg.Operation(OpCompleteVisit, m.opCompleteVisit).Requires(model.ResourceOf(&MedicalHistoryModel), model.Update).Accepts(&CompleteVisitArgs{})
	reg.Operation(OpListSupplyUse, m.opListSupplyUse).Requires(model.ResourceOf(&EncounterSupplyModel), model.Read).Accepts(&ListSupplyUseArgs{})
	reg.Operation(OpUpsertSupplyUse, m.opUpsertSupplyUse).Requires(model.ResourceOf(&EncounterSupplyModel), model.Update).Accepts(&UpsertSupplyUseArgs{})
	reg.Operation(OpDeleteSupplyUse, m.opDeleteSupplyUse).Requires(model.ResourceOf(&EncounterSupplyModel), model.Delete).Accepts(&SupplyUseRefArgs{})
	reg.Operation(OpListPendingSupply, m.opListPendingSupply).Requires(model.ResourceOf(&EncounterSupplyModel), model.Read).Accepts(&ListPendingSupplyArgs{})
	reg.Operation(OpRetrySupplyUse, m.opRetrySupplyUse).Requires(model.ResourceOf(&EncounterSupplyModel), model.Update).Accepts(&SupplyUseRefArgs{})
}

var _ router.OperationModule = (*Module)(nil)


func (m *Module) opGetVisit(ctx router.Context) {
	var args GetVisitArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, ErrCannotRead)
		return
	}
	record, err := m.GetVisit(args.TenantId, args.Id)
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	if err := ctx.Encode(record); err != nil {
		rpc.WriteError(ctx, err)
	}
}

func (m *Module) opListVisits(ctx router.Context) {
	var args ListVisitsArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, ErrCannotRead)
		return
	}
	records, err := m.ListVisitsByPatient(args.TenantId, args.PatientId)
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	list := make(MedicalHistoryList, len(records))
	for i, r := range records {
		list[i] = r
	}
	if err := ctx.Encode(&list); err != nil {
		rpc.WriteError(ctx, err)
	}
}

func (m *Module) opListRecentPatients(ctx router.Context) {
	var args ListRecentPatientsArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, ErrCannotRead)
		return
	}
	records, err := m.ListRecentPatients(args.TenantId, int(args.Limit))
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	list := make(MedicalHistoryList, len(records))
	for i, r := range records {
		list[i] = r
	}
	if err := ctx.Encode(&list); err != nil {
		rpc.WriteError(ctx, err)
	}
}



func (m *Module) opCompleteVisit(ctx router.Context) {
	var args CompleteVisitArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	record, _, err := m.CompleteVisit(args.TenantId, args.Id, ctx.UserID())
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	if err := ctx.Encode(record); err != nil {
		rpc.WriteError(ctx, err)
	}
}

func (m *Module) opListSupplyUse(ctx router.Context) {
	var args ListSupplyUseArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	records, err := m.ListSupplyUse(args.TenantId, args.EncounterId)
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	list := make(EncounterSupplyList, len(records))
	for i, r := range records {
		list[i] = r
	}
	if err := ctx.Encode(&list); err != nil {
		rpc.WriteError(ctx, err)
	}
}

func (m *Module) opUpsertSupplyUse(ctx router.Context) {
	var args UpsertSupplyUseArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	record, err := m.UpsertSupplyUse(args)
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	if err := ctx.Encode(record); err != nil {
		rpc.WriteError(ctx, err)
	}
}

func (m *Module) opDeleteSupplyUse(ctx router.Context) {
	var args SupplyUseRefArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	err := m.DeleteSupplyUse(args.TenantId, args.Id)
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	ctx.WriteStatus(200)
}

func (m *Module) opListPendingSupply(ctx router.Context) {
	var args ListPendingSupplyArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	records, err := m.ListPendingSupply(args.TenantId)
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	list := make(EncounterSupplyList, len(records))
	for i, r := range records {
		list[i] = r
	}
	if err := ctx.Encode(&list); err != nil {
		rpc.WriteError(ctx, err)
	}
}

func (m *Module) opRetrySupplyUse(ctx router.Context) {
	var args SupplyUseRefArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	record, err := m.RetrySupplyUse(args.TenantId, args.Id, ctx.UserID())
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	if err := ctx.Encode(record); err != nil {
		rpc.WriteError(ctx, err)
	}
}
