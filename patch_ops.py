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
