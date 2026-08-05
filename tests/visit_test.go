package tests

import (
	"testing"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
)

func TestCreateVisit_HappyPath(t *testing.T) {
	m := setup(t)
	rec, err := m.CreateVisit(clinicalencounter.CreateVisitArgs{
		PatientId:           "pat_1",
		DoctorId:            "doc_1",
		Reason:              "Control rutinario",
		AttentionAt:         1700000000,
		PatientNameSnapshot: "Juan Pérez",
		PatientRutSnapshot:  "1-9",
		DoctorNameSnapshot:  "Dr. Soto",
	})
	if err != nil {
		t.Fatalf("CreateVisit: %v", err)
	}
	if rec.Id == "" {
		t.Fatal("expected a non-empty Id")
	}
	if rec.Status != clinicalencounter.StatusCreated {
		t.Fatalf("expected Status %q, got %q", clinicalencounter.StatusCreated, rec.Status)
	}
}

func TestCreateVisit_MissingRequiredArgs(t *testing.T) {
	m := setup(t)
	_, err := m.CreateVisit(clinicalencounter.CreateVisitArgs{})
	if err != clinicalencounter.ErrMissingArgs {
		t.Fatalf("expected ErrMissingArgs, got %v", err)
	}
}

func TestGetVisit_NotFound(t *testing.T) {
	m := setup(t)
	_, err := m.GetVisit("does-not-exist")
	if err != clinicalencounter.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListVisitsByPatient(t *testing.T) {
	m := setup(t)
	_, err := m.CreateVisit(clinicalencounter.CreateVisitArgs{
		PatientId: "pat_1", DoctorId: "doc_1", Reason: "r", AttentionAt: 1,
		PatientNameSnapshot: "P", PatientRutSnapshot: "1-9", DoctorNameSnapshot: "D",
	})
	if err != nil {
		t.Fatalf("CreateVisit: %v", err)
	}
	list, err := m.ListVisitsByPatient("pat_1")
	if err != nil {
		t.Fatalf("ListVisitsByPatient: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 record, got %d", len(list))
	}
}
