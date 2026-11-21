package clinical_encounter

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

const titleMedicalHistory = "Medical history"

// NewView construye el Presenter de las fichas; opts agrega opciones de view (p. ej. view.WithFilter por paciente).
//
// Ejemplo: p := NewView(caller, view.WithFilter(func() model.Model { return &MedicalHistory{PatientId: id} }))
func NewView(caller router.Caller, opts ...view.Option) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, Page: view.Page{First: view.All}},
		func() model.TableSlice { return &MedicalHistoryList{} })
	return view.New(b, &MedicalHistory{}, append([]view.Option{view.WithTitle(titleMedicalHistory)}, opts...)...)
}
