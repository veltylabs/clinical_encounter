← [Stage 2](PLAN_STAGE_2_module_go.md) | Next → [Stage 4](PLAN_STAGE_4_readers_used_by_other_module.md)

> Read AGENTS.md and the index `PLAN.md` (rules, tables, external calls) first. Commit when this stage compiles.

## Stage 3 — generate, then write the rules & external calls

1. Delete `ops.go`, `crud.go`, `queries.go`, `model_crud.go`, `model_orm.go` and `definition_crud.go` (those that exist). Run `definitionc`: it writes one `definition_gen.go` (the types, schemas, external call types `MedicalHistoryDeductSupplies`, `EncounterSupplyDeduct`, and on `Module`, `ModelName()`, `Definition()`, `MountOperations()`) and the skeletons:
   - `rule_<entity>_<verb>.go` for every action declared `AddMyBusinessRule`.
   - `call_<entity>_<call>.go` for every call in `ExternalCalls`.
2. Add `store.go`: for each table of «Tables, stores and rules» that has a Store, `// Store<Entity> is the store of <entity>.` `func (m *Module) Store<Entity>() crud.Store { return m.store }`.
3. Fill each rule skeleton:
   - `MedicalHistory.Update`: checks professional = `a.UserID` and the other checks of `CompleteVisit` (lines 1–40 of `complete.go`); when the sent `fields` include `status` = `StatusCompleted` and `m.supplyOn`: `return crud.ApplyAndPublish(MedicalHistoryDeductSupplies{}), nil`, otherwise `crud.ApplyAndPublish(), nil`.
   - `EncounterSupply.Update`: if `old.Status != SupplyPendingDeduction` → `return crud.RuleAnswer{}, ErrLineNotPending`; otherwise `return crud.ApplyAndPublish(EncounterSupplyDeduct{}), nil`.
   - Rejection: `return crud.RuleAnswer{}, err`.
4. Calls. First change `deductLine(v, l, by)` in `complete.go` so it only sets fields on `l` (delete its last line, the `m.db.Update(...)` and `l.UpdatedAt`; it returns nothing) and add `var supplyDeductFields = []string{"status", "cost", "failure", "room_id", "deducted_at"}`. Both calls run in the background: `a` is empty, never read it; the professional is `visit.DoctorId` (completion is allowed only to the professional).

```go
// call_medical_history_deduct_supplies.go
func (p *MedicalHistory) DeductSupplies(a crud.CurrentUser, m *Module, key crud.IdempotencyKey) (crud.FieldsToSave, error) {
	lines, err := m.ListSupplyUse(p.TenantId, p.Id)
	if err != nil {
		return nil, err
	}
	pending := 0
	for _, l := range lines {
		if l.Status != SupplyPlanned {
			continue
		}
		m.deductLine(p, l, p.DoctorId)
		if err := m.store.UpdateFields(l, l.Id, supplyDeductFields); err != nil {
			return nil, err
		}
		if l.Status == SupplyPendingDeduction {
			pending++
		}
	}
	if pending > 0 && m.pub != nil {
		// the same notify.Raise(...) block that CompleteVisit has today (AlertSupplyPending), moved here unchanged
	}
	return nil, nil
}

// call_encounter_supply_deduct.go
func (p *EncounterSupply) Deduct(a crud.CurrentUser, m *Module, key crud.IdempotencyKey) (crud.FieldsToSave, error) {
	visit, err := m.GetVisit(p.TenantId, p.EncounterId)
	if err != nil {
		return nil, err
	}
	m.deductLine(visit, p, visit.DoctorId)
	return crud.Fields(supplyDeductFields...), nil
}
```

   The comment line inside the `if pending > 0` block is replaced by the real `notify.Raise(m.pub, notify.Alert{...})` call copied from `CompleteVisit` (delete it there). `deductLine` records a failure in the row (`pending_deduction` + `failure` text) and returns nothing, so these calls always save the outcome; `RetryInBackground` is only the way the work leaves the request.
Delete every `Op*` constant, `res*`/`Resource*` variable, every `op*` handler, every `view.MountSave`/`view.MountDelete`/`reg.Operation` call and the `webtyp.com/router`, `webtyp.com/rpc` imports that only they used (`grep -rn "reg.Operation\|view.MountSave\|view.MountDelete\|router.Context\|rpc.WriteError" --include=*.go . | grep -v _test` → empty).
