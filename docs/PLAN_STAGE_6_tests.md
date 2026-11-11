← [Stage 5](PLAN_STAGE_5_browser_side.md) | Next → [Stage 7](PLAN_STAGE_7_translations_and_documentati.md)

> Read AGENTS.md and the index `PLAN.md` (rules, tables, external calls) first. Commit when this stage compiles.

## Stage 6 — tests (`tests/`, through `crud.LoopbackActingAsWithQueue(nil, "u1", &callQueue{}, mod)`; rows in `docs/TESTS.md`)

Delete tests of removed code.

`tests/queue_double_test.go` (copy it as is; it is the only queue the tests use):

```go
package tests

import (
	"webtyp.com/crud"
	"webtyp.com/fmt"
	"webtyp.com/model"
)

// callQueue is the test double of crud.CallQueue: it records what crud asks and stores nothing.
type callQueue struct {
	added  []crud.CallRequest
	done   []string
	failed []string // the reasons
}

func (q *callQueue) Add(tx crud.Tx, r crud.CallRequest) (string, error) {
	q.added = append(q.added, r)
	return fmt.Sprintf("call-%d", len(q.added)), nil
}
func (q *callQueue) MarkDone(id string) error { q.done = append(q.done, id); return nil }
func (q *callQueue) MarkFailedAttempt(id string, call model.ExternalCallDefinition, reason string) error {
	q.failed = append(q.failed, reason)
	return nil
}
func (q *callQueue) ConfirmCall(id string) error        { return nil }
func (q *callQueue) RejectCall(id, reason string) error { return nil }
```

The caller of every test that touches a call is `crud.LoopbackActingAsWithQueue(nil, "u1", &callQueue{}, mod)` (keep the pointer in a variable to assert `added`, `done`, `failed`).

Operations are `router.Operation{Module: ModelName, Entity: "<entity>", Action: model.<Action>}`, never a string; fixtures follow `crud` v0.0.14 (`reg.AddStoredEntity`, `crud.CurrentUser`). Add `TestMount_MatchesDefinition` (`crud.Mount(reg, nil, nil, mod)` returns nil) and one test that an allowed write publishes a `*crud.Change` on `crud.Topic(ModelName, "<entity>", model.<Action>)`. Required:

| Test | Setup → call → expected |
|---|---|
| `TestComplete_DeductsViaExternalCall` | visit with one planned line, fake rooms and supplies; update visit `["status"]` = completed → `queue.added` has `DeductSupplies`; read the visit row, call `row.DeductSupplies(crud.CurrentUser{}, mod, "k")` → the line reads `deducted` and the fake supplies saw one use |
| `TestComplete_FailureStaysPending` | rooms fake finds no shift → after `DeductSupplies` the line reads `pending_deduction` with the failure text; update its `["status"]` = `SupplyPendingDeduction` → `queue.added` has `Deduct`; give the fake a shift, call `line.Deduct(crud.CurrentUser{}, mod, "k")` → `deducted` |
| `TestSupply_RetryOnlyPending` | update `["status"]` = `SupplyPendingDeduction` on a `planned` line → error `ErrLineNotPending` |
| `TestComplete_OnlyOwnProfessional` | another user completes → field error |
| `TestSupply_DeleteDeductedRejected` | delete a deducted row → error |
| `TestAgenda_IsCallers` | visits of two professionals; agenda as u1 → only u1's |
| `TestDemo_Valid` (`//go:build wasm`) | `seed.Demo()` → no error, `Validate()` nil, at least one screen |
