package tests

import (
	"testing"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"webtyp.com/events/inprocess"
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
	userOf    map[string]string // staffID → userID
}

func (f *fakeDoctorReader) StaffByUser(tenantID, userID string) (string, string, string, bool, error) {
	return f.staffID, f.name, f.specialty, f.found, f.err
}

func (f *fakeDoctorReader) UserByStaff(tenantID, staffID string) (string, bool, error) {
	u, ok := f.userOf[staffID]
	return u, ok, nil
}

type fakeAgendaReader struct {
	resIDs   []string
	patIDs   []string
	svcIDs   []string
	starts   []time.Unix
	statuses []string
	err      error
	watchers []func(string)
}

func (f *fakeAgendaReader) ActiveByStaff(tenantID, staffID string, fromUtc, toUtc time.Unix) ([]string, []string, []string, []time.Unix, []string, error) {
	return f.resIDs, f.patIDs, f.svcIDs, f.starts, f.statuses, f.err
}

func (f *fakeAgendaReader) WatchStaffAgenda(fn func(string)) { f.watchers = append(f.watchers, fn) }

type fakePatient struct{ name, doc, birth string }

type fakePatientReader struct{ patients map[string]fakePatient }

func (f *fakePatientReader) GetPatient(tenantID, patientID string) (string, string, string, string, bool, error) {
	if p, ok := f.patients[patientID]; ok {
		return p.name, p.doc, "", "", true, nil
	}
	return "", "", "", "", false, nil
}

func (f *fakePatientReader) PatientBirthdate(tenantID, patientID string) (string, bool, error) {
	p, ok := f.patients[patientID]
	return p.birth, ok, nil
}

type fakeServiceReader struct{}

func (fakeServiceReader) ServiceCard(tenantId, serviceId string) (string, float64, string, bool, error) {
	return "Consulta General", 0, "", true, nil
}

// fakeNotifier records the notices the module sends.
type fakeNotifier struct{ sent []string }

func (n *fakeNotifier) Notify(userID, topic string) { n.sent = append(n.sent, userID+" "+topic) }

func newAgendaModule(t *testing.T, notices clinicalencounter.Notifier) (*clinicalencounter.Module, *fakeAgendaReader) {
	t.Helper()
	zone, err := time.LoadZone("America/Santiago")
	if err != nil {
		t.Fatalf("LoadZone: %v", err)
	}
	agenda := &fakeAgendaReader{
		resIDs:   []string{"res-1", "res-2"},
		patIDs:   []string{"pat-1", "pat-2"},
		svcIDs:   []string{"svc-1", "svc-1"},
		starts:   []time.Unix{1700000000, 1700001800},
		statuses: []string{clinicalencounter.ReservationConfirmed, clinicalencounter.ReservationPending},
	}
	m, err := clinicalencounter.New(orm.New(mem.New()), clinicalencounter.Deps{TenantID: "t1",
		Publisher: &inprocess.Broker{},
		Zone:      zone,
		Doctors:   &fakeDoctorReader{staffID: "staff-1", name: "Dra. Soto", specialty: "Pediatría", found: true, userOf: map[string]string{"staff-1": "user-doc-1"}},
		Agenda:    agenda,
		Patients: &fakePatientReader{patients: map[string]fakePatient{
			"pat-1": {name: "Juan Pérez", doc: "12.345.678-9", birth: "1960-11-05"},
			"pat-2": {name: "Ana Rojas", doc: "9.876.543-2"},
		}},
		Services: fakeServiceReader{},
		Notices:  notices,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return m, agenda
}

// La agenda del día trae cada reserva activa con su estado, la edad del paciente y el nombre del servicio.
func TestMyAgenda(t *testing.T) {
	m, _ := newAgendaModule(t, nil)
	res, err := m.MyAgenda(clinicalencounter.MyAgendaArgs{TenantId: "t1", UserId: "user-doc-1", Date: "2026-11-04"})
	if err != nil {
		t.Fatalf("MyAgenda: %v", err)
	}
	if res.StaffId != "staff-1" || len(res.Entries) != 2 {
		t.Fatalf("staff=%q entries=%d, want staff-1 and 2", res.StaffId, len(res.Entries))
	}
	first, second := res.Entries[0], res.Entries[1]
	if first.PatientName != "Juan Pérez" || first.PatientDocument != "12.345.678-9" || first.ServiceName != "Consulta General" {
		t.Errorf("unexpected entry data: %+v", first)
	}
	if first.PatientAge != 65 { // cumple 66 el 5 de noviembre, la agenda es del 4
		t.Errorf("age = %d, want 65", first.PatientAge)
	}
	if first.ReservationStatus != clinicalencounter.ReservationConfirmed || second.ReservationStatus != clinicalencounter.ReservationPending {
		t.Errorf("statuses = %q, %q", first.ReservationStatus, second.ReservationStatus)
	}
	if second.PatientAge != 0 {
		t.Errorf("a patient without birth date has age 0, got %d", second.PatientAge)
	}
}

// Solo la reserva confirmada o ya atendida abre la ficha.
func TestAttendable(t *testing.T) {
	for status, want := range map[string]bool{
		clinicalencounter.ReservationConfirmed: true, clinicalencounter.ReservationCompleted: true,
		clinicalencounter.ReservationPending: false, clinicalencounter.ReservationNoShow: false, clinicalencounter.ReservationConflicted: false,
	} {
		if got := clinicalencounter.Attendable(status); got != want {
			t.Errorf("Attendable(%s) = %v, want %v", status, got, want)
		}
	}
}

// Cuando cambia una reserva del profesional, su usuario recibe el aviso AgendaChanged.
func TestAgendaChange_NotifiesTheDoctor(t *testing.T) {
	notices := &fakeNotifier{}
	_, agenda := newAgendaModule(t, notices)
	if len(agenda.watchers) != 1 {
		t.Fatalf("watchers = %d, want 1", len(agenda.watchers))
	}
	agenda.watchers[0]("staff-1")
	agenda.watchers[0]("staff-without-user")
	if len(notices.sent) != 1 || notices.sent[0] != "user-doc-1 "+clinicalencounter.AgendaChanged {
		t.Fatalf("sent = %v, want one notice to user-doc-1", notices.sent)
	}
}
