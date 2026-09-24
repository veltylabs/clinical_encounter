//go:build wasm

package tests

import (
	"testing"

	"webtyp.com/json"
	"webtyp.com/model"

	ab "github.com/veltylabs/appointment_booking"
	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"github.com/veltylabs/clinical_encounter/ui"
	patientdirectory "github.com/veltylabs/patient_directory"
	staffmanager "github.com/veltylabs/staff_manager"

	"webtyp.com/components/searchbar"
	"webtyp.com/components/selectsearch"
	"webtyp.com/layout/crudview"
)

func encodeInto(v model.Encodable, into model.Decodable) error {
	var out []byte
	if err := json.Encode(v, &out); err != nil {
		return err
	}
	return json.Decode(string(out), into)
}

// TestWASM_ClinicalEncounter_TodayAgendaFromConfirmedReservations pins the
// design in docs/UI.md: the doctor's picker is
// built from staff_manager + appointment_booking + patient_directory, shows
// only CONFIRMED reservations, and Filter(patientID) resolves to that
// patient's pre-fetched history without any further network call.
func TestWASM_ClinicalEncounter_TodayAgendaFromConfirmedReservations(t *testing.T) {
	staff := staffmanager.StaffMemberList{
		{Id: "staff1", TenantId: "t1", UserId: "user1", Name: "Dra. Soto", Specialty: "Pediatría"},
	}
	reservations := ab.ReservationList{
		{Id: "r1", TenantId: "t1", ClientId: "pat1", StaffIdsnapshot: "staff1", Status: ab.StatusConfirmed, LocalStringTime: "09:00"},
		{Id: "r2", TenantId: "t1", ClientId: "pat2", StaffIdsnapshot: "staff1", Status: ab.StatusPending, LocalStringTime: "09:30"},
	}
	patient := patientdirectory.Patient{Id: "pat1", TenantId: "t1", Name: "Juan Pérez", Rut: "12.345.678-9"}

	var calledOps []string
	mock := &mockCaller{
		onCall: func(op string, args model.Encodable, into model.Decodable) error {
			calledOps = append(calledOps, op)
			switch op {
			case staffmanager.ModelName + "." + staffmanager.OpListStaff:
				return encodeInto(&staff, into)
			case ab.ModelName + "." + ab.OpListReservationsByStaff:
				return encodeInto(&reservations, into)
			case patientdirectory.ModelName + "." + patientdirectory.OpGetPatient:
				return encodeInto(&patient, into)
			case clinicalencounter.ModelName + "." + clinicalencounter.OpListVisitsByPatient:
				empty := clinicalencounter.MedicalHistoryList{}
				return encodeInto(&empty, into)
			}
			return nil
		},
	}

	m, err := ui.Browser(mock, &testIDGen{}, "t1", "user1")
	if err != nil {
		t.Fatalf("Browser: %v", err)
	}

	cv, ok := m.View().(*crudview.CrudView)
	if !ok {
		t.Fatalf("expected *crudview.CrudView, got %T", m.View())
	}
	if _, ok := cv.Filter.(*selectsearch.SelectSearch); !ok {
		t.Fatalf("expected the agenda picker (*selectsearch.SelectSearch) once the logged-in user resolves to a staff row, got %T", cv.Filter)
	}

	for _, want := range []string{
		staffmanager.ModelName + "." + staffmanager.OpListStaff,
		ab.ModelName + "." + ab.OpListReservationsByStaff,
		patientdirectory.ModelName + "." + patientdirectory.OpGetPatient,
	} {
		found := false
		for _, op := range calledOps {
			if op == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected op %q to be called while building the agenda, calledOps=%v", want, calledOps)
		}
	}

	// The confirmed reservation's patient is selectable; the pending one is not.
	if items := cv.Presenter.Filter("pat1"); items == nil {
		t.Error("expected pat1 (CONFIRMED reservation) to be selectable via Filter")
	}
	if items := cv.Presenter.Filter("pat2"); items != nil {
		t.Errorf("expected pat2 (PENDING reservation) to be absent from the agenda, got %v", items)
	}
}

// TestWASM_ClinicalEncounter_NoStaffRowFallsBackToPlainList covers the
// administrator case: a logged-in account with no matching staff_member row
// gets the plain free-text list instead of a failed construction.
func TestWASM_ClinicalEncounter_NoStaffRowFallsBackToPlainList(t *testing.T) {
	mock := &mockCaller{
		onCall: func(op string, args model.Encodable, into model.Decodable) error {
			if op == staffmanager.ModelName+"."+staffmanager.OpListStaff {
				empty := staffmanager.StaffMemberList{}
				return encodeInto(&empty, into)
			}
			return nil
		},
	}

	m, err := ui.Browser(mock, &testIDGen{}, "t1", "admin-with-no-staff-row")
	if err != nil {
		t.Fatalf("Browser: %v", err)
	}
	cv, ok := m.View().(*crudview.CrudView)
	if !ok {
		t.Fatalf("expected *crudview.CrudView, got %T", m.View())
	}
	if _, ok := cv.Filter.(*searchbar.SearchBar); !ok {
		t.Errorf("expected crudview's default SearchBar when the logged-in account has no staff row, got %T", cv.Filter)
	}
}

// TestWASM_ClinicalEncounter_NewDraftIsSeededFromThePicker: el pago de P1.
// Elegir un paciente en la agenda y pulsar "+" debe dejar la ficha nueva ya
// con su id, nombre y RUT — y con el doctor de la sesión. Es lo único que
// separa "el doctor selecciona y escribe" de "el doctor selecciona": si esto
// se rompe, el formulario sale en blanco y nadie se entera hasta que alguien
// tipea un RUT a mano.
func TestWASM_ClinicalEncounter_NewDraftIsSeededFromThePicker(t *testing.T) {
	staff := staffmanager.StaffMemberList{
		{Id: "staff1", TenantId: "t1", UserId: "user1", Name: "Dra. Soto", Specialty: "Pediatría"},
	}
	reservations := ab.ReservationList{
		{Id: "r1", TenantId: "t1", ClientId: "pat1", StaffIdsnapshot: "staff1", Status: ab.StatusConfirmed, LocalStringTime: "09:00"},
	}
	patient := patientdirectory.Patient{Id: "pat1", TenantId: "t1", Name: "Juan Pérez", Rut: "12.345.678-9"}

	mock := &mockCaller{
		onCall: func(op string, args model.Encodable, into model.Decodable) error {
			switch op {
			case staffmanager.ModelName + "." + staffmanager.OpListStaff:
				return encodeInto(&staff, into)
			case ab.ModelName + "." + ab.OpListReservationsByStaff:
				return encodeInto(&reservations, into)
			case patientdirectory.ModelName + "." + patientdirectory.OpGetPatient:
				return encodeInto(&patient, into)
			case clinicalencounter.ModelName + "." + clinicalencounter.OpListVisitsByPatient:
				empty := clinicalencounter.MedicalHistoryList{}
				return encodeInto(&empty, into)
			}
			return nil
		},
	}

	m, err := ui.Browser(mock, &testIDGen{}, "t1", "user1")
	if err != nil {
		t.Fatalf("Browser: %v", err)
	}
	cv, ok := m.View().(*crudview.CrudView)
	if !ok {
		t.Fatalf("expected *crudview.CrudView, got %T", m.View())
	}
	if cv.NewRecord == nil {
		t.Fatal("la pantalla debe sembrar el borrador nuevo; NewRecord quedó sin cablear")
	}

	// Sin paciente elegido: el borrador trae al doctor, no al paciente.
	empty, _ := cv.NewRecord().(*clinicalencounter.MedicalHistory)
	if empty == nil || empty.PatientId != "" {
		t.Errorf("sin selección no debe inventarse un paciente, got %+v", empty)
	}
	if empty.DoctorId != "staff1" {
		t.Errorf("el doctor sale de la sesión, no del picker: got %q", empty.DoctorId)
	}

	// El picker elige a pat1 — crudview llama a Filter con ese término.
	cv.Presenter.Filter("pat1")

	seeded, _ := cv.NewRecord().(*clinicalencounter.MedicalHistory)
	if seeded == nil {
		t.Fatal("NewRecord debe devolver un *MedicalHistory")
	}
	if seeded.PatientId != "pat1" {
		t.Errorf("patient_id sembrado: got %q, want pat1", seeded.PatientId)
	}
	if seeded.PatientNameSnapshot != "Juan Pérez" {
		t.Errorf("nombre sembrado: got %q", seeded.PatientNameSnapshot)
	}
	if seeded.PatientRutSnapshot != "12.345.678-9" {
		t.Errorf("RUT sembrado: got %q", seeded.PatientRutSnapshot)
	}

	// Instancia fresca en cada "+": editar un borrador no debe filtrarse al siguiente.
	seeded.Reason = "editado"
	if again, _ := cv.NewRecord().(*clinicalencounter.MedicalHistory); again.Reason != "" {
		t.Errorf("cada borrador es una instancia nueva; got Reason=%q", again.Reason)
	}
}
