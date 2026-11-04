package tests

import (
	"testing"
	"webtyp.com/time"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
)

func TestImportVisitKeepsDates(t *testing.T) {
	m := setup(t)
	attentionAt := time.Unix(1520848800) // 2018-03-12 10:00 UTC
	v := clinicalencounter.MedicalHistory{
		Id:                  "v1",
		TenantId:            "t1",
		PatientId:           "p1",
		DoctorId:            "d1",
		Reason:              "control",
		AttentionAt:         attentionAt,
		PatientNameSnapshot: "Ana",
		PatientRutSnapshot:  "1-9",
		DoctorNameSnapshot:  "Dra. Pérez",
	}

	imported, err := m.ImportVisit(v)
	if err != nil {
		t.Fatalf("ImportVisit: %v", err)
	}

	if imported.Status != clinicalencounter.StatusCompleted {
		t.Errorf("expected status %s, got %s", clinicalencounter.StatusCompleted, imported.Status)
	}

	expectedNano := time.Nano(int64(attentionAt) * 1e9)
	if imported.StartedAt != expectedNano {
		t.Errorf("expected StartedAt %v, got %v", expectedNano, imported.StartedAt)
	}
	if imported.FinishedAt != expectedNano {
		t.Errorf("expected FinishedAt %v, got %v", expectedNano, imported.FinishedAt)
	}
}

func TestImportVisitTwice(t *testing.T) {
	m := setup(t)
	v := clinicalencounter.MedicalHistory{
		Id:                  "v1",
		TenantId:            "t1",
		PatientId:           "p1",
		DoctorId:            "d1",
		Reason:              "control",
		AttentionAt:         100,
		PatientNameSnapshot: "Ana",
		PatientRutSnapshot:  "1-9",
		DoctorNameSnapshot:  "Dra. Pérez",
	}

	_, err := m.ImportVisit(v)
	if err != nil {
		t.Fatalf("first ImportVisit: %v", err)
	}

	v.Diagnostic = "new diagnostic"
	_, err = m.ImportVisit(v)
	if err != nil {
		t.Fatalf("second ImportVisit: %v", err)
	}

	list, err := m.ListVisitsByPatient("t1", "p1")
	if err != nil {
		t.Fatalf("ListVisitsByPatient: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 row, got %d", len(list))
	}
	if list[0].Diagnostic != "new diagnostic" {
		t.Errorf("expected updated diagnostic 'new diagnostic', got %q", list[0].Diagnostic)
	}
}

func TestImportVisitOtherPatient(t *testing.T) {
	m := setup(t)
	v1 := clinicalencounter.MedicalHistory{
		Id:                  "v1",
		TenantId:            "t1",
		PatientId:           "p1",
		DoctorId:            "d1",
		Reason:              "control",
		AttentionAt:         100,
		PatientNameSnapshot: "Ana",
		PatientRutSnapshot:  "1-9",
		DoctorNameSnapshot:  "Dra. Pérez",
	}
	_, err := m.ImportVisit(v1)
	if err != nil {
		t.Fatalf("first ImportVisit: %v", err)
	}

	v2 := v1
	v2.PatientId = "p2"
	_, err = m.ImportVisit(v2)
	if err == nil || err.Error() != clinicalencounter.ErrIdTaken.Error() {
		t.Fatalf("expected ErrIdTaken, got %v", err)
	}
}

func TestImportVisitNoDate(t *testing.T) {
	m := setup(t)
	v := clinicalencounter.MedicalHistory{
		Id:                  "v1",
		TenantId:            "t1",
		PatientId:           "p1",
		DoctorId:            "d1",
		Reason:              "control",
		AttentionAt:         0,
		PatientNameSnapshot: "Ana",
		PatientRutSnapshot:  "1-9",
		DoctorNameSnapshot:  "Dra. Pérez",
	}
	_, err := m.ImportVisit(v)
	if err == nil || err.Error() != clinicalencounter.ErrImportDate.Error() {
		t.Fatalf("expected ErrImportDate, got %v", err)
	}
}

func TestImportVisitMissingReason(t *testing.T) {
	m := setup(t)
	v := clinicalencounter.MedicalHistory{
		Id:                  "v1",
		TenantId:            "t1",
		PatientId:           "p1",
		DoctorId:            "d1",
		Reason:              "",
		AttentionAt:         100,
		PatientNameSnapshot: "Ana",
		PatientRutSnapshot:  "1-9",
		DoctorNameSnapshot:  "Dra. Pérez",
	}
	_, err := m.ImportVisit(v)
	if err == nil || err.Error() != clinicalencounter.ErrMissingArgs.Error() {
		t.Fatalf("expected ErrMissingArgs, got %v", err)
	}
}

func TestImportVisitNoSupplies(t *testing.T) {
	recipes := &fakeRecipes{
		items: []string{"item1"},
		qtys:  []float64{1.0},
	}
	m := setupSupplyModule(t, recipes)

	v := clinicalencounter.MedicalHistory{
		Id:                  "v1",
		TenantId:            "t1",
		PatientId:           "p1",
		DoctorId:            "d1",
		ServiceId:           "srv1",
		Reason:              "control",
		AttentionAt:         100,
		PatientNameSnapshot: "Ana",
		PatientRutSnapshot:  "1-9",
		DoctorNameSnapshot:  "Dra. Pérez",
	}
	_, err := m.ImportVisit(v)
	if err != nil {
		t.Fatalf("ImportVisit: %v", err)
	}

	lines, err := m.ListSupplyUse("t1", "v1")
	if err != nil {
		t.Fatalf("ListSupplyUse: %v", err)
	}
	if len(lines) != 0 {
		t.Fatalf("expected 0 lines, got %d", len(lines))
	}
}

func TestImportVisitNoEvent(t *testing.T) {
	pub := &mockPublisher{}
	db := orm.New(mem.New())
	m, err := clinicalencounter.New(db, clinicalencounter.Deps{Publisher: pub})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	v := clinicalencounter.MedicalHistory{
		Id:                  "v1",
		TenantId:            "t1",
		PatientId:           "p1",
		DoctorId:            "d1",
		Reason:              "control",
		AttentionAt:         100,
		PatientNameSnapshot: "Ana",
		PatientRutSnapshot:  "1-9",
		DoctorNameSnapshot:  "Dra. Pérez",
	}

	_, err = m.ImportVisit(v)
	if err != nil {
		t.Fatalf("ImportVisit: %v", err)
	}

	if len(pub.Events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(pub.Events))
	}
}
