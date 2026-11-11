package tests

import (
	"testing"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"webtyp.com/events/mock"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	"webtyp.com/time"
)

type fakeDoctorReader struct {
	staffID   string
	name      string
	specialty string
	found     bool
	err       error
}

func (f *fakeDoctorReader) StaffByUser(tenantID, userID string) (string, string, string, bool, error) {
	return f.staffID, f.name, f.specialty, f.found, f.err
}

type fakeAgendaReader struct {
	resIDs []string
	patIDs []string
	svcIDs []string
	starts []time.Unix
	err    error
}

func (f *fakeAgendaReader) ConfirmedByStaff(tenantID, staffID string, fromUtc, toUtc time.Unix) ([]string, []string, []string, []time.Unix, error) {
	return f.resIDs, f.patIDs, f.svcIDs, f.starts, f.err
}

type fakePatientReader struct {
	patients map[string]struct {
		name string
		doc  string
	}
}

func (f *fakePatientReader) GetPatient(tenantID, patientID string) (string, string, string, string, bool, error) {
	if p, ok := f.patients[patientID]; ok {
		return p.name, p.doc, "", "", true, nil
	}
	return "", "", "", "", false, nil
}

func TestMyAgenda(t *testing.T) {
	db := orm.New(mem.New())
	broker := &mock.Broker{}

	zone, err := time.LoadZone("America/Santiago")
	if err != nil {
		t.Fatalf("LoadZone: %v", err)
	}

	docFake := &fakeDoctorReader{
		staffID:   "staff-1",
		name:      "Dra. Soto",
		specialty: "Pediatría",
		found:     true,
	}

	agendaFake := &fakeAgendaReader{
		resIDs: []string{"res-1"},
		patIDs: []string{"pat-1"},
		svcIDs: []string{"svc-1"},
		starts: []time.Unix{1700000000},
	}

	patFake := &fakePatientReader{
		patients: map[string]struct {
			name string
			doc  string
		}{
			"pat-1": {name: "Juan Pérez", doc: "12.345.678-9"},
		},
	}

	m, err := clinicalencounter.New(db, clinicalencounter.Deps{TenantID: "t1",
		Publisher: broker,
		Zone:      zone,
		Doctors:   docFake,
		Agenda:    agendaFake,
		Patients:  patFake,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := m.MyAgenda(clinicalencounter.MyAgendaArgs{
		TenantId: "t1",
		UserId:   "user-doc-1",
		Date:     "2026-11-04",
	})
	if err != nil {
		t.Fatalf("MyAgenda: %v", err)
	}

	if res.StaffId != "staff-1" {
		t.Errorf("StaffId: got %q, want staff-1", res.StaffId)
	}
	if len(res.Entries) != 1 {
		t.Fatalf("Entries count: got %d, want 1", len(res.Entries))
	}
	e := res.Entries[0]
	if e.PatientId != "pat-1" || e.PatientName != "Juan Pérez" || e.PatientDocument != "12.345.678-9" {
		t.Errorf("unexpected entry data: %+v", e)
	}
}
