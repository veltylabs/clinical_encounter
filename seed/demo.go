//go:build wasm

package seed

import (
	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"github.com/veltylabs/clinical_encounter/migrate"
	"github.com/veltylabs/clinical_encounter/ui"
	"webtyp.com/crud"
	"webtyp.com/ddl"
	"webtyp.com/events/mock"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	"webtyp.com/time"
	"webtyp.com/unixid"
	"webtyp.com/view"
)

const demoTenantID = "demo"

type fakeDoctorReader struct{}

func (fakeDoctorReader) StaffByUser(tenantID, userID string) (string, string, string, bool, error) {
	return "demo-doc-1", "Dra. Soto", "Pediatría", true, nil
}

type fakeAgendaReader struct{}

func (fakeAgendaReader) ConfirmedByStaff(tenantID, staffID string, from, to time.Unix) ([]string, []string, []string, []time.Unix, error) {
	now := time.NowUnix()
	return []string{"res-1", "res-2"}, []string{"demo-pat-1", "demo-pat-2"}, []string{"svc-1", "svc-1"}, []time.Unix{now, now + 1800}, nil
}

type fakePatientReader struct{}

func (fakePatientReader) GetPatient(tenantID, patientID string) (string, string, string, string, bool, error) {
	switch patientID {
	case "demo-pat-1":
		return "Juan Pérez", "12.345.678-9", "", "", true, nil
	case "demo-pat-2":
		return "María López", "9.876.543-2", "", "", true, nil
	default:
		return "", "", "", "", false, nil
	}
}

// Demo crea una demostración ejecutable del módulo clinical_encounter sobre una base en memoria con datos semilla.
//
// Ejemplo: demo, err := seed.Demo()
func Demo() (view.Demo, error) {
	ids, err := unixid.NewUnixID()
	if err != nil {
		return view.Demo{}, err
	}
	db := orm.New(mem.New())
	zone, err := time.LoadZone("America/Santiago")
	if err != nil {
		return view.Demo{}, err
	}
	mod, err := clinicalencounter.New(db, clinicalencounter.Deps{
		Publisher: &mock.Broker{},
		IDs:       ids,
		TenantID:  demoTenantID,
		Zone:      zone,
		Doctors:   fakeDoctorReader{},
		Agenda:    fakeAgendaReader{},
		Patients:  fakePatientReader{},
	})
	if err != nil {
		return view.Demo{}, err
	}
	if _, err := Load(mod, demoTenantID); err != nil {
		return view.Demo{}, err
	}
	return view.Demo{
		Label:      ui.Label,
		Timezone:   "America/Santiago",
		ActingAs:   "demo-user-ana",
		Modules:    []crud.Module{mod},
		Migrations: []ddl.Step{{Name: clinicalencounter.Definition.Name, Run: migrate.Migrate}},
		Screens: func(d view.BrowserDeps) []view.Screen {
			return ui.Screens(d)
		},
	}, nil
}
