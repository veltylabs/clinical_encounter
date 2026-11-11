← [Index](PLAN.md) | Next → [Stage 2](PLAN_STAGE_2_module_go.md)

> Read AGENTS.md and the index `PLAN.md` (rules, tables, external calls) first. Commit when this stage compiles.

## Stage 1 — `definition.go`

- Set `Allow` on each table exactly as in «Tables, stores and rules»; a table not listed there gets no `Allow` (internal).
- Mark `ServerOnly: true`: `tenant_id`, `created_at`, `updated_at`, and every field the server fills (`status` on create, `cost`, `failure`, `deducted_at`, `room_id`).
- Delete every `*Args` / `*Result` definition that only carried a removed operation (`grep -n "ArgsModel\|ResultModel" model.go` → only the ones this plan keeps).
- Declare `ExternalCalls: []model.ExternalCallDefinition{...}` on:
  - `MedicalHistoryModel`: `{Name: "DeductSupplies", Mode: model.RetryInBackground, MaxAttempts: 3}`.
  - `EncounterSupplyModel`: `{Name: "Deduct", Mode: model.RetryInBackground, MaxAttempts: 3}`.
- `upsert_supply_use` used to need only Update; creating a row now needs `encounter_supply`/Create: list it in the README permissions table.
- **Keep every `*Args` model that an exported reader's signature still uses** (`grep -n "^func (m \*Module) [A-Z][A-Za-z]*([^)]*Args" *.go` lists them): mark each `Wire: true`; delete only the `*Args` no remaining signature uses.
- D52 (the entity is its own filter): a filter model (`*Filter`, `*FilterModel`, an `*ArgsModel` that only carried filter values) is not kept: its fields become fields of the read model or table (`OmitEmpty`), and the rule reads them from its receiver; delete the filter type and its `Wire`/`NewFilter` plumbing.
- Add `var Definition = model.ModuleDefinition{Name: "clinical_encounter", Entities: []model.Definition{…every table and read model of this plan…}}` and, on every table and read model, `Rules` as «Tables, stores and rules» says (`AddMyBusinessRule` where a rule is named, `HaveNoBusinessRule` for the other allowed actions; a read model: `Read: model.AddMyBusinessRule`). Delete the hand topics (`Topic*`) and `Err*`/resources no code uses any more.
