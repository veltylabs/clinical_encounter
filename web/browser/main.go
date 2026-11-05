package main

import (
	. "webtyp.com/dom"
	"webtyp.com/events/mock"
	"webtyp.com/layout/platformd"
	"webtyp.com/orm"
	"webtyp.com/router/loopback"
	"webtyp.com/storage/mem"
	"webtyp.com/time"
	"webtyp.com/unixid"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
	ceseed "github.com/veltylabs/clinical_encounter/seed"
	"github.com/veltylabs/clinical_encounter/ui"
)

const demoTenantID = "demo"

type demoUser struct{}

func (demoUser) UserName() string    { return "Demo" }
func (demoUser) UserAvatar() string  { return "" }
func (demoUser) UserRoles() []string { return []string{"Doctor"} }

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

func main() {
	ids, err := unixid.NewUnixID()
	if err != nil {
		panic(err)
	}
	db := orm.New(mem.New())
	broker := &mock.Broker{}

	zone, _ := time.LoadZone("America/Santiago")

	ceMod, err := clinicalencounter.New(db, clinicalencounter.Deps{
		Publisher: broker,
		IDs:       ids,
		TenantID:  demoTenantID,
		Zone:      zone,
		Doctors:   fakeDoctorReader{},
		Agenda:    fakeAgendaReader{},
		Patients:  fakePatientReader{},
	})
	if err != nil {
		panic(err)
	}

	if _, err := ceseed.Load(ceMod, demoTenantID); err != nil {
		panic(err)
	}

	caller := loopback.WithTenant(demoTenantID, ceMod)
	screens := ui.Screens(caller, zone, ids, demoTenantID, "demo-user-ana")

	p := &platformd.Platform{
		AppName:   ui.Label + " — demo",
		User:      demoUser{},
		DefaultID: ui.ID,
	}
	_ = screens
	Append("body", p)
	select {}
}
