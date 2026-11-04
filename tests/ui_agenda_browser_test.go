//go:build browser

package tests

import (
	"testing"

	"webtyp.com/json"
	"webtyp.com/model"
	"webtyp.com/time"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"github.com/veltylabs/clinical_encounter/ui"

	"webtyp.com/components/searchbar"
	"webtyp.com/components/selectsearch"
	"webtyp.com/layout/crudview"
	"webtyp.com/view"
)

func encodeInto(v model.Encodable, into model.Decodable) error {
	var out []byte
	if err := json.Encode(v, &out); err != nil {
		return err
	}
	return json.Decode(string(out), into)
}

// TestWASM_ClinicalEncounter_TodayAgendaFromConfirmedReservations pins the
// design: the agenda picker is built via OpMyAgenda, showing confirmed entries
// and allowing filtering by patient without external dependencies.
func TestWASM_ClinicalEncounter_TodayAgendaFromConfirmedReservations(t *testing.T) {
	agendaResult := clinicalencounter.MyAgendaResult{
		StaffId:         "staff1",
		DoctorName:      "Dra. Soto",
		DoctorSpecialty: "Pediatría",
		Entries: []clinicalencounter.AgendaEntry{
			{
				PatientId:       "pat1",
				PatientName:     "Juan Pérez",
				PatientDocument: "12.345.678-9",
				TimeLabel:       "09:00",
				ReservationId:   "r1",
				ServiceId:       "s1",
			},
		},
	}

	var calledOps []string
	mock := &mockCaller{
		onCall: func(op string, args model.Encodable, into model.Decodable) error {
			calledOps = append(calledOps, op)
			switch op {
			case clinicalencounter.ModelName + "." + clinicalencounter.OpMyAgenda:
				return encodeInto(&agendaResult, into)
			case clinicalencounter.ModelName + "." + clinicalencounter.OpListVisitsByPatient:
				empty := clinicalencounter.MedicalHistoryList{}
				return encodeInto(&empty, into)
			}
			return nil
		},
	}

	screens := ui.Screens(view.NewRecords().Caller(mock), time.MachineZone(), &testIDGen{}, "t1", "user1")
	if len(screens) == 0 {
		t.Fatalf("expected at least 1 screen, got 0")
	}
	cv, ok := screens[0].Component.(*crudview.CrudView)
	if !ok {
		t.Fatalf("expected *crudview.CrudView, got %T", screens[0].Component)
	}
	if _, ok := cv.Filter.(*selectsearch.SelectSearch); !ok {
		t.Fatalf("expected the agenda picker (*selectsearch.SelectSearch) once the logged-in user resolves to a staff row, got %T", cv.Filter)
	}

	want := clinicalencounter.ModelName + "." + clinicalencounter.OpMyAgenda
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

	// The confirmed reservation's patient is selectable; unknown patient is not.
	if items := cv.Presenter.Filter("pat1"); items == nil {
		t.Error("expected pat1 to be selectable via Filter")
	}
	if items := cv.Presenter.Filter("pat2"); items != nil {
		t.Errorf("expected pat2 to be absent from the agenda, got %v", items)
	}
}

// TestWASM_ClinicalEncounter_NoStaffRowFallsBackToPlainList covers the
// administrator case: a logged-in account with no matching staff_member row
// gets the plain free-text list instead of a failed construction.
func TestWASM_ClinicalEncounter_NoStaffRowFallsBackToPlainList(t *testing.T) {
	mock := &mockCaller{
		onCall: func(op string, args model.Encodable, into model.Decodable) error {
			if op == clinicalencounter.ModelName+"."+clinicalencounter.OpMyAgenda {
				empty := clinicalencounter.MyAgendaResult{}
				return encodeInto(&empty, into)
			}
			return nil
		},
	}

	screens := ui.Screens(view.NewRecords().Caller(mock), time.MachineZone(), &testIDGen{}, "t1", "admin-with-no-staff-row")
	if len(screens) == 0 {
		t.Fatalf("expected at least 1 screen, got 0")
	}
	cv, ok := screens[0].Component.(*crudview.CrudView)
	if !ok {
		t.Fatalf("expected *crudview.CrudView, got %T", screens[0].Component)
	}
	if _, ok := cv.Filter.(*searchbar.SearchBar); !ok {
		t.Errorf("expected crudview's default SearchBar when the logged-in account has no staff row, got %T", cv.Filter)
	}
}

// TestWASM_ClinicalEncounter_NewDraftIsSeededFromThePicker:
// Elegir un paciente en la agenda y pulsar "+" debe dejar la ficha nueva ya
// con su id, nombre y RUT — y con el doctor de la sesión.
func TestWASM_ClinicalEncounter_NewDraftIsSeededFromThePicker(t *testing.T) {
	agendaResult := clinicalencounter.MyAgendaResult{
		StaffId:         "staff1",
		DoctorName:      "Dra. Soto",
		DoctorSpecialty: "Pediatría",
		Entries: []clinicalencounter.AgendaEntry{
			{
				PatientId:       "pat1",
				PatientName:     "Juan Pérez",
				PatientDocument: "12.345.678-9",
				TimeLabel:       "09:00",
				ReservationId:   "r1",
				ServiceId:       "s1",
			},
		},
	}

	mock := &mockCaller{
		onCall: func(op string, args model.Encodable, into model.Decodable) error {
			switch op {
			case clinicalencounter.ModelName + "." + clinicalencounter.OpMyAgenda:
				return encodeInto(&agendaResult, into)
			case clinicalencounter.ModelName + "." + clinicalencounter.OpListVisitsByPatient:
				empty := clinicalencounter.MedicalHistoryList{}
				return encodeInto(&empty, into)
			}
			return nil
		},
	}

	screens := ui.Screens(view.NewRecords().Caller(mock), time.MachineZone(), &testIDGen{}, "t1", "user1")
	if len(screens) == 0 {
		t.Fatalf("expected at least 1 screen, got 0")
	}
	cv, ok := screens[0].Component.(*crudview.CrudView)
	if !ok {
		t.Fatalf("expected *crudview.CrudView, got %T", screens[0].Component)
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
