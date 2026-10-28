package clinical_encounter

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

const titleMedicalHistory = "Historial clínico"

// NewView construye el Presenter del historial médico — el motor agnóstico de
// tecnología que envuelve un renderer (rightpanel, o cualquier otro). Este
// módulo lo construye (view + model + router); la app decide qué renderer lo
// dibuja.
func NewView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListVisitsByPatient, Save: OpCreateVisit},
		func() model.ModelSlice { return &MedicalHistoryList{} })
	return view.New(b, &MedicalHistory{}, view.WithTitle(titleMedicalHistory))
}
