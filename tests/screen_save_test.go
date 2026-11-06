package tests

import (
	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"testing"
	"webtyp.com/events/mock"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/router/loopback"
	"webtyp.com/storage/mem"
	"webtyp.com/view"
)

func TestScreenSave(t *testing.T) {
	db := orm.New(mem.New())
	m, err := clinicalencounter.New(db, clinicalencounter.Deps{TenantID: "t1", Publisher: &mock.Broker{}})
	if err != nil {
		t.Fatal(err)
	}

	caller := loopback.ActingAs("t1", "u1", m)
	b := view.NewCallerLister(view.NewRecords().Caller(caller),
		view.Ops{Module: clinicalencounter.ModelName, List: clinicalencounter.OpListVisitsByPatient, Save: clinicalencounter.OpCreateVisit},
		func() model.ModelSlice { return &clinicalencounter.MedicalHistoryList{} })

	rec := &clinicalencounter.MedicalHistory{
		Id:                  "vis1",
		TenantId:            "t1",
		PatientId:           "p1",
		DoctorId:            "d1",
		Reason:              "reason",
		PatientNameSnapshot: "name",
		PatientRutSnapshot:  "rut",
		DoctorNameSnapshot:  "dname",
		Status:              clinicalencounter.StatusCreated,
	}

	done := make(chan error, 1)
	if saver, ok := b.(view.Saver); ok {
		saver.Save([]model.Model{rec}, func(err error) {
			done <- err
		})
	} else {
		t.Fatal("expected view.Saver")
	}

	err = <-done
	if err != nil {
		t.Fatalf("expected nil error on save, got %v", err)
	}

	// Try to save a duplicate to test conflict
	done = make(chan error, 1)
	rec2 := &clinicalencounter.MedicalHistory{
		Id:                  "vis1",
		TenantId:            "t1",
		PatientId:           "p2",
		DoctorId:            "d1",
		Reason:              "reason",
		PatientNameSnapshot: "name",
		PatientRutSnapshot:  "rut",
		DoctorNameSnapshot:  "dname",
		Status:              clinicalencounter.StatusCreated,
	}
	if saver, ok := b.(view.Saver); ok {
		saver.Save([]model.Model{rec2}, func(err error) {
			done <- err
		})
	}

	err = <-done
	if model.KindOf(err) != model.ErrorConflict {
		t.Fatalf("expected error conflict, got %v", err)
	}
}
