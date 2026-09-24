package tests

import (
	"testing"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
	ceseed "github.com/veltylabs/clinical_encounter/seed"
	patientseed "github.com/veltylabs/patient_directory/seed"
	patientdirectory "github.com/veltylabs/patient_directory"
	staffseed "github.com/veltylabs/staff_manager/seed"
	staffmanager "github.com/veltylabs/staff_manager"
	"webtyp.com/events/mock"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
)

func TestSeedLoad(t *testing.T) {
	db := orm.New(mem.New())
	broker := &mock.Broker{}
	ids := &testIDGen{}

	m, err := clinicalencounter.New(db, clinicalencounter.Deps{IDs: ids, Publisher: broker})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	up := ceseed.Upstream{
		Patients: patientseed.Data{
			Patients: []patientdirectory.Patient{
				{Id: "p1", Name: "Patient One", Rut: "11111111-1"},
				{Id: "p2", Name: "Patient Two", Rut: "22222222-2"},
			},
		},
		Staff: staffseed.Data{
			Staff: []staffmanager.StaffMember{
				{Id: "s1", Name: "Doctor One", Specialty: "Medicina General"},
			},
		},
	}

	data, err := ceseed.Load(m, up)
	if err != nil {
		t.Fatalf("seed.Load: %v", err)
	}

	if len(data.Visits) != 2 {
		t.Errorf("expected 2 visits sembradas, got %d", len(data.Visits))
	}
}
