package clinical_encounter

import (
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

const ResourceSupplyUse = "encounter_supply"

func (m *Module) MountOperations(reg router.OperationRegistry) {
	reg.Operation(OpCreateVisit, m.opCreateVisit).Requires("medical_history", model.Create).Accepts(&CreateVisitArgs{})
	reg.Operation(OpGetVisit, m.opGetVisit).Requires("medical_history", model.Read).Accepts(&GetVisitArgs{})
	reg.Operation(OpListVisitsByPatient, m.opListVisits).Requires("medical_history", model.Read).Accepts(&ListVisitsArgs{})
	reg.Operation(OpListRecentPatients, m.opListRecentPatients).Requires("medical_history", model.Read).Accepts(&ListRecentPatientsArgs{})

	reg.Operation(OpCompleteVisit, m.opCompleteVisit).Requires("medical_history", model.Update).Accepts(&CompleteVisitArgs{})
	reg.Operation(OpListSupplyUse, m.opListSupplyUse).Requires(ResourceSupplyUse, model.Read).Accepts(&ListSupplyUseArgs{})
	reg.Operation(OpUpsertSupplyUse, m.opUpsertSupplyUse).Requires(ResourceSupplyUse, model.Update).Accepts(&UpsertSupplyUseArgs{})
	reg.Operation(OpDeleteSupplyUse, m.opDeleteSupplyUse).Requires(ResourceSupplyUse, model.Delete).Accepts(&SupplyUseRefArgs{})
	reg.Operation(OpListPendingSupply, m.opListPendingSupply).Requires(ResourceSupplyUse, model.Read).Accepts(&ListPendingSupplyArgs{})
	reg.Operation(OpRetrySupplyUse, m.opRetrySupplyUse).Requires(ResourceSupplyUse, model.Update).Accepts(&SupplyUseRefArgs{})
}

var _ router.OperationModule = (*Module)(nil)

func (m *Module) opCreateVisit(ctx router.Context) {
	var args CreateVisitArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	record, err := m.CreateVisit(args)
	if err != nil {
		// Convención de estado (en todo el ecosistema): 400 = entrada inválida, 404 = no encontrado,
		// 500 = solo error interno genuino — nunca colapsar errores de cliente en 500.
		writeError(ctx, err)
		return
	}
	if err := ctx.Encode(record); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opGetVisit(ctx router.Context) {
	var args GetVisitArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	record, err := m.GetVisit(args.TenantId, args.Id)
	if err != nil {
		writeError(ctx, err)
		return
	}
	if err := ctx.Encode(record); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opListVisits(ctx router.Context) {
	var args ListVisitsArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	records, err := m.ListVisitsByPatient(args.TenantId, args.PatientId)
	if err != nil {
		writeError(ctx, err)
		return
	}
	list := make(MedicalHistoryList, len(records))
	for i, r := range records {
		list[i] = r
	}
	if err := ctx.Encode(&list); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opListRecentPatients(ctx router.Context) {
	var args ListRecentPatientsArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	records, err := m.ListRecentPatients(args.TenantId, int(args.Limit))
	if err != nil {
		writeError(ctx, err)
		return
	}
	list := make(MedicalHistoryList, len(records))
	for i, r := range records {
		list[i] = r
	}
	if err := ctx.Encode(&list); err != nil {
		ctx.WriteStatus(500)
	}
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
