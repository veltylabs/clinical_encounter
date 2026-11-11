---
PLAN: "test: clinical_encounter rule tests, part 2a (queue double and four tests)"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 18023957337203884021
PR: https://github.com/veltylabs/clinical_encounter/pull/19
REVIEWED_COMMIT: 6ba9c1c54f9a8df5a701c533e676d719e8b7591b
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work. Commit after every stage.
> Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.
> **Publishing:** you publish only by OPENING THE PULL REQUEST (a `git push` from your sandbox does not reach GitHub). Submit after EACH test file that compiles and keep committing.
> **Never ask a yes/no question about progress**; ask only when a signature in this plan is wrong, and name the symbol.
> **Before you submit:** `git status` must be clean and `git diff --stat main` must list the files of the plan. A pull request that carries only go.mod, go.sum or docs/PLAN.md is a failed delivery: `git add -A`, commit and submit again.

# Plan: `clinical_encounter` — the first four tests (part 2a)

The module is migrated and published (v0.2.79). Its old tests of removed code are gone; this plan writes the first four tests, small and exact. Nothing else: no translations, no docs (part 2b). Never edit `definition_gen.go`.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```
No `go get`, no `replace`. Read files by ranges (`sed -n 1,80p file`).

## Stage 1 — `tests/queue_double_test.go` (new, copy as is)

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

## Stage 2 — `tests/rules_test.go` (new)

The fixtures `fakeRecipes`, `fakeRooms`, `fakeStock`, `fakeIDs` already exist in `tests/supply_fakes_test.go`; `setup(t)` in `tests/setup_test.go`. Write exactly (adapt only a name the compiler refuses, and say which in the PR):

```go
package tests

import (
	"testing"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/router"
	"webtyp.com/router/mock"
	"webtyp.com/storage/mem"
)

// moduleWithSupplies builds the module with the recipe, the room and the stock fakes the visit needs.
func moduleWithSupplies(t *testing.T, rooms *fakeRooms, stock *fakeStock) *clinicalencounter.Module {
	t.Helper()
	m, err := clinicalencounter.New(orm.New(mem.New()), clinicalencounter.Deps{
		TenantID:     "t1",
		Recipes:      &fakeRecipes{items: []string{"i1"}, qtys: []float64{1}},
		Rooms:        rooms,
		Supplies:     stock,
		IDs:          &fakeIDs{},
		SuppliesArea: "supplies_area",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return m
}

// newVisit creates visit v1 of doctor d1 with one planned supply line.
func newVisit(t *testing.T, m *clinicalencounter.Module) {
	t.Helper()
	_, err := m.CreateVisit(clinicalencounter.CreateVisitArgs{
		TenantId: "t1", Id: "v1", PatientId: "p1", DoctorId: "d1", Reason: "control", AttentionAt: 100,
		PatientNameSnapshot: "P", PatientRutSnapshot: "1-9", DoctorNameSnapshot: "D", ServiceId: "s1",
	})
	if err != nil {
		t.Fatalf("CreateVisit: %v", err)
	}
}

// complete asks, as `user`, to complete visit v1 and answers the callback error.
func complete(m *clinicalencounter.Module, q *callQueue, user string) error {
	var callErr error
	crud.LoopbackActingAsWithQueue(nil, user, q, m).Call(
		router.Operation{Module: clinicalencounter.ModelName, Entity: "medical_history", Action: model.Update},
		&crud.UpdateArgs{IDs: []string{"v1"}, Fields: []string{"status"}, Record: &clinicalencounter.MedicalHistory{Status: clinicalencounter.StatusCompleted}},
		nil, func(err error) { callErr = err })
	return callErr
}

func TestMount_MatchesDefinition(t *testing.T) {
	m := moduleWithSupplies(t, &fakeRooms{}, &fakeStock{})
	if err := crud.MountWithQueue(&mock.Router{}, nil, nil, &callQueue{}, m); err != nil {
		t.Fatalf("the generated mount does not match the definition: %v", err)
	}
}

func TestComplete_OnlyOwnProfessional(t *testing.T) {
	m := moduleWithSupplies(t, &fakeRooms{roomId: "r1", found: true}, &fakeStock{cost: 500})
	newVisit(t, m)
	q := &callQueue{}
	if err := complete(m, q, "u2"); err == nil {
		t.Fatal("another user completed the visit")
	}
	if len(q.added) != 0 {
		t.Fatalf("a refused completion must not queue a call, got %d", len(q.added))
	}
}

func TestComplete_DeductsViaExternalCall(t *testing.T) {
	m := moduleWithSupplies(t, &fakeRooms{roomId: "r1", found: true}, &fakeStock{cost: 500})
	newVisit(t, m)
	q := &callQueue{}
	if err := complete(m, q, "d1"); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if len(q.added) != 1 || q.added[0].Call != "DeductSupplies" {
		t.Fatalf("expected one DeductSupplies call, got %+v", q.added)
	}
	visit, err := m.GetVisit("t1", "v1")
	if err != nil {
		t.Fatalf("GetVisit: %v", err)
	}
	if _, err := visit.DeductSupplies(crud.CurrentUser{}, m, "k"); err != nil {
		t.Fatalf("DeductSupplies: %v", err)
	}
	lines, err := m.ListSupplyUse("t1", "v1")
	if err != nil || len(lines) != 1 || lines[0].Status != clinicalencounter.SupplyDeducted {
		t.Fatalf("expected one deducted line, got %+v, %v", lines, err)
	}
}

func TestComplete_FailureStaysPending(t *testing.T) {
	m := moduleWithSupplies(t, &fakeRooms{found: false}, &fakeStock{})
	newVisit(t, m)
	q := &callQueue{}
	if err := complete(m, q, "d1"); err != nil {
		t.Fatalf("complete: %v", err)
	}
	visit, _ := m.GetVisit("t1", "v1")
	if _, err := visit.DeductSupplies(crud.CurrentUser{}, m, "k"); err != nil {
		t.Fatalf("DeductSupplies: %v", err)
	}
	lines, _ := m.ListSupplyUse("t1", "v1")
	if len(lines) != 1 || lines[0].Status != clinicalencounter.SupplyPendingDeduction || lines[0].Failure == "" {
		t.Fatalf("expected one pending line with the failure text, got %+v", lines)
	}
}
```
(`GetVisit` answers a `*MedicalHistory` or a value: use `&visit` if it is a value. `mock.Router` is `webtyp.com/router/mock`.) Each test has its row in `docs/TESTS.md` (`| [Name](../tests/rules_test.go) | rule |`, header `| Test | Rule |`).

## Delivery check

`gotest` passes on every destination; the four tests exist with their rows. Submit as soon as Stage 1 and the first test pass, then keep committing. A tree that does not compile or whose tests did not run is a failed delivery.

## Executor notes
No changes were required to compiler refused names.
