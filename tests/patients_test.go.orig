package tests

import (
	"testing"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
)

func TestListRecentPatients_DedupesByPatientMostRecentFirst(t *testing.T) {
	m := setup(t)

	// Two visits for pat_1 (attention_at 1 and 3), one for pat_2 (attention_at 2).
	// Expect: [pat_1, pat_2] — one row per patient, ordered by the LATEST
	// attention_at each has, not by creation order.
	for _, v := range []clinicalencounter.CreateVisitArgs{
		{PatientId: "pat_1", DoctorId: "doc_1", Reason: "r1", AttentionAt: 1, PatientNameSnapshot: "Juan", PatientRutSnapshot: "1-9", DoctorNameSnapshot: "D"},
		{PatientId: "pat_2", DoctorId: "doc_1", Reason: "r2", AttentionAt: 2, PatientNameSnapshot: "Ana", PatientRutSnapshot: "2-7", DoctorNameSnapshot: "D"},
		{PatientId: "pat_1", DoctorId: "doc_1", Reason: "r3", AttentionAt: 3, PatientNameSnapshot: "Juan", PatientRutSnapshot: "1-9", DoctorNameSnapshot: "D"},
	} {
		if _, err := m.CreateVisit(v); err != nil {
			t.Fatalf("CreateVisit: %v", err)
		}
	}

	patients, err := m.ListRecentPatients(0)
	if err != nil {
		t.Fatalf("ListRecentPatients: %v", err)
	}
	if len(patients) != 2 {
		t.Fatalf("expected 2 distinct patients, got %d: %+v", len(patients), patients)
	}
	if patients[0].PatientId != "pat_1" {
		t.Errorf("expected pat_1 first (latest attention_at=3), got %q", patients[0].PatientId)
	}
	if patients[1].PatientId != "pat_2" {
		t.Errorf("expected pat_2 second, got %q", patients[1].PatientId)
	}
}

func TestListRecentPatients_RespectsLimit(t *testing.T) {
	m := setup(t)
	for i, id := range []string{"pat_1", "pat_2", "pat_3"} {
		_, err := m.CreateVisit(clinicalencounter.CreateVisitArgs{
			PatientId: id, DoctorId: "doc_1", Reason: "r", AttentionAt: int64(i + 1),
			PatientNameSnapshot: id, PatientRutSnapshot: "1-9", DoctorNameSnapshot: "D",
		})
		if err != nil {
			t.Fatalf("CreateVisit: %v", err)
		}
	}

	patients, err := m.ListRecentPatients(2)
	if err != nil {
		t.Fatalf("ListRecentPatients: %v", err)
	}
	if len(patients) != 2 {
		t.Fatalf("expected limit=2 to cap the result, got %d", len(patients))
	}
}

func TestListRecentPatients_EmptyWhenNoVisits(t *testing.T) {
	m := setup(t)
	patients, err := m.ListRecentPatients(0)
	if err != nil {
		t.Fatalf("ListRecentPatients: %v", err)
	}
	if len(patients) != 0 {
		t.Fatalf("expected no patients, got %d", len(patients))
	}
}
