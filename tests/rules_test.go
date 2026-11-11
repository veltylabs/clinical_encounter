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
