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

// TestCreateVisit_DefaultsAttentionAtToNow: attention_at is model.Int() in
// MedicalHistoryModel (not an input.* widget — see model.go's own field
// comment), so webtyp/form never renders it and every crudview-driven "new
// ficha" submission reaches here with AttentionAt == 0. Rejecting that as
// ErrMissingArgs (the pre-fix behavior) made the generic "+" create flow
// permanently fail in production — a clinical encounter is created because
// it is happening now, so a zero AttentionAt must default to time.Now(),
// not error.
func TestCreateVisit_DefaultsAttentionAtToNow(t *testing.T) {
	m := setup(t)
	rec, err := m.CreateVisit(clinicalencounter.CreateVisitArgs{
		PatientId:           "pat_1",
		DoctorId:            "doc_1",
		Reason:              "Control rutinario",
		PatientNameSnapshot: "Juan Pérez",
		PatientRutSnapshot:  "1-9",
		DoctorNameSnapshot:  "Dr. Soto",
	})
	if err != nil {
		t.Fatalf("CreateVisit: %v", err)
	}
	if rec.AttentionAt <= 0 {
		t.Fatalf("expected AttentionAt to default to now, got %d", rec.AttentionAt)
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
