package clinical_encounter

import (
	"github.com/tinywasm/model"
	"github.com/tinywasm/router"
	"github.com/tinywasm/view"
)

// Item implements view.Itemizer — the ONLY view-specific code this record carries. The
// Presenter indexes rows by ID from this during Reload; there is no manual byID/WithFill lookup.
func (it *MedicalHistory) Item() view.Item {
	return view.Item{ID: it.Id, Label: it.Reason, Description: it.Status}
}

// NewView builds the medical-history Presenter — the tech-agnostic engine a renderer
// (tinywasm/layout/rightpanel, or any other) wraps. This module builds it (view+model+router
// only); the app decides which renderer draws it.
func NewView(caller router.Caller) view.Presenter {
	record := &MedicalHistory{}

	return view.New(
		caller,
		record,
		OpListVisitsByPatient,
		func() model.ModelSlice { return &MedicalHistoryList{} },
		view.WithTitle("Historial clínico"),
		view.WithSaveOp(OpCreateVisit),
	)
}
