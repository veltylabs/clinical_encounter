//go:build wasm

package main

import (
	"github.com/tinywasm/dom"
	"github.com/tinywasm/layout/rightpanel"
	"github.com/tinywasm/unixid"

	"github.com/veltylabs/clinical_encounter"
	"github.com/veltylabs/clinical_encounter/web/mockdb"
)

func main() {
	// Initialize DB
	db := mockdb.New()
	uid, _ := unixid.NewUnixID()
	mod := &clinical_encounter.Module{DB: db, UID: uid, Pub: nil}

	// Mock data
	items := mockHistoryData()
	patientName := "Juan Pérez González"
	patientRUT := "12.345.678-9"

	// Initialize controller
	ctrl := clinical_encounter.NewViewCtrl(items)

	// Build views
	patientHead := &clinical_encounter.PatientHeadView{
		Name: patientName,
		RUT:  patientRUT,
	}

	// Build layout
	panel := &rightpanel.RightPanel{
		Module:        mod,
		Title:         "Ficha Paciente",
		Head:          patientHead,
		HeadControls:  nil,
		Article:       ctrl.GetVisitFormView(),
		AsideControls: &clinical_encounter.HistorySearchView{},
		Aside:         ctrl.GetHistoryListView(),
	}

	// Render
	dom.Render("body", panel.Render())

	// Setup event handlers
	ctrl.SetupEventHandlers()

	select {}
}

// mockHistoryData returns mock medical history records.
func mockHistoryData() []*clinical_encounter.MedicalHistory {
	return []*clinical_encounter.MedicalHistory{
		{
			ID:                  "mh_001",
			Status:              clinical_encounter.StatusCompleted,
			Reason:              "Control rutinario",
			PatientNameSnapshot: "Juan Pérez González",
			DoctorNameSnapshot:  "Dr. Rodrigo Soto",
		},
		{
			ID:                  "mh_002",
			Status:              clinical_encounter.StatusCompleted,
			Reason:              "Dolor de cabeza",
			PatientNameSnapshot: "Juan Pérez González",
			DoctorNameSnapshot:  "Dr. Rodrigo Soto",
		},
		{
			ID:                  "mh_003",
			Status:              clinical_encounter.StatusInProgress,
			Reason:              "Revisión exámenes",
			PatientNameSnapshot: "Juan Pérez González",
			DoctorNameSnapshot:  "Dr. Rodrigo Soto",
		},
		{
			ID:                  "mh_004",
			Status:              clinical_encounter.StatusArrived,
			Reason:              "Consulta general",
			PatientNameSnapshot: "Juan Pérez González",
			DoctorNameSnapshot:  "Dr. Rodrigo Soto",
		},
	}
}
