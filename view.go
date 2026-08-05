package clinical_encounter

import (
	"github.com/tinywasm/model"
	"github.com/tinywasm/router"
	"github.com/tinywasm/view"
)

// Item implementa view.Itemizer — el ÚNICO código específico de view que carga este registro. El
// Presenter indexa las filas por ID a partir de esto durante Reload; no hay lookup manual byID/WithFill.
func (it *MedicalHistory) Item() view.Item {
	return view.Item{ID: it.Id, Label: it.Reason, Description: it.Status}
}

// NewView construye el Presenter del historial médico — el motor agnóstico de tecnología que envuelve
// un renderer (tinywasm/layout/rightpanel, or any other). Este módulo lo construye (solo
// view+model+router); la app decide qué renderer lo dibuja.
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
