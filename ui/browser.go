package ui

import (
	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"webtyp.com/components/selectsearch"
	"webtyp.com/components/targetdate"
	"webtyp.com/crud"
	"webtyp.com/dom"
	"webtyp.com/events"
	"webtyp.com/fmt"
	"webtyp.com/html"
	"webtyp.com/lang"
	"webtyp.com/layout/crudview"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/svg"
	"webtyp.com/time"
	"webtyp.com/view"
)

// Textos fijos de la pantalla (claves en inglés, traducidas en lang.json).
const (
	textPickPatient lang.Text = "Select a patient from today's agenda..."
	textYears       lang.Text = "years"
	textPending     lang.Text = "Pending"
	textConfirmed   lang.Text = "Confirmed"
	textCompleted   lang.Text = "Completed"
	textNoShow      lang.Text = "No show"
	textConflicted  lang.Text = "Conflict"
)

// statusText es la etiqueta visible de un estado de reserva.
func statusText(status string) string {
	switch status {
	case clinicalencounter.ReservationPending:
		return lang.Translate(textPending).String()
	case clinicalencounter.ReservationConfirmed:
		return lang.Translate(textConfirmed).String()
	case clinicalencounter.ReservationCompleted:
		return lang.Translate(textCompleted).String()
	case clinicalencounter.ReservationNoShow:
		return lang.Translate(textNoShow).String()
	case clinicalencounter.ReservationConflicted:
		return lang.Translate(textConflicted).String()
	}
	return status
}

// optionOf es la fila del selector para una reserva del día: «Nombre, 66 años» / RUT (lo que el doctor usa en sus
// documentos) / «09:35 · Confirmada». Solo una reserva que se atiende se puede elegir.
func optionOf(e clinicalencounter.AgendaEntry) selectsearch.SsOption {
	label := e.PatientName
	if e.PatientAge > 0 {
		label += ", " + fmt.Convert(e.PatientAge).String() + " " + lang.Translate(textYears).String()
	}
	return selectsearch.SsOption{
		ID:          e.ReservationId,
		Label:       label,
		Sublabel:    e.PatientDocument,
		Description: e.TimeLabel + " · " + statusText(e.ReservationStatus),
		Disabled:    !clinicalencounter.Attendable(e.ReservationStatus),
	}
}

// doctorAgenda es la agenda del día del profesional con sesión y la reserva elegida. La lista de fichas es el
// presenter estándar de medical_history filtrado por el paciente de esa reserva (view.WithFilter).
type doctorAgenda struct {
	caller router.Caller
	zone   time.Zone
	picker *selectsearch.SelectSearch
	list   *crudview.CrudView

	opened bool
	agenda clinicalencounter.MyAgendaResult
	chosen clinicalencounter.AgendaEntry
}

// loadAgenda lee la agenda del día y rellena el selector. Si la reserva elegida dejó de poder atenderse, se suelta.
func (a *doctorAgenda) loadAgenda() {
	date := ""
	if a.zone.Name() != "" {
		date = a.zone.Date(time.NowUnix())
	}
	var rows clinicalencounter.MyAgendaResultList
	a.caller.Call(router.Operation{Module: clinicalencounter.ModelName, Entity: clinicalencounter.MyAgendaResultModel.Name, Action: model.Read},
		&crud.ReadArgs{Filter: &clinicalencounter.MyAgendaResult{Date: date}}, &rows, func(err error) {
			if err != nil || len(rows) == 0 || rows[0] == nil {
				return
			}
			a.agenda = *rows[0]
			opts := make([]selectsearch.SsOption, len(a.agenda.Entries))
			still := false
			for i, e := range a.agenda.Entries {
				opts[i] = optionOf(e)
				if e.ReservationId == a.chosen.ReservationId && clinicalencounter.Attendable(e.ReservationStatus) {
					a.chosen, still = e, true
				}
			}
			a.picker.SetOptions(opts)
			if a.chosen.ReservationId != "" && !still {
				a.choose("")
			}
		})
}

// choose fija la reserva elegida y recarga la lista: el filtro ahora es su paciente.
func (a *doctorAgenda) choose(reservationID string) {
	a.chosen = clinicalencounter.AgendaEntry{}
	for _, e := range a.agenda.Entries {
		if e.ReservationId == reservationID && clinicalencounter.Attendable(e.ReservationStatus) {
			a.chosen = e
		}
	}
	if a.list != nil {
		a.list.Reload(nil)
	}
}

// patientFilter es el filtro de cada lectura de fichas: el paciente elegido (ninguno = ninguna ficha).
func (a *doctorAgenda) patientFilter() model.Model {
	return &clinicalencounter.MedicalHistory{PatientId: a.chosen.PatientId}
}

// opening carga la agenda la primera vez que crudview recarga la lista, es decir, al abrir la pantalla.
func (a *doctorAgenda) opening(crudview.ListView) {
	if !a.opened {
		a.opened = true
		a.loadAgenda()
	}
}

// newRecord siembra la ficha nueva con el profesional y la reserva elegida: el doctor no escribe lo que el sistema
// ya sabe.
func (a *doctorAgenda) newRecord() model.Model {
	rec := &clinicalencounter.MedicalHistory{
		DoctorId:                a.agenda.StaffId,
		DoctorNameSnapshot:      a.agenda.DoctorName,
		DoctorSpecialtySnapshot: a.agenda.DoctorSpecialty,
		Status:                  clinicalencounter.StatusCreated,
	}
	if a.chosen.PatientId != "" {
		rec.PatientId = a.chosen.PatientId
		rec.PatientNameSnapshot = a.chosen.PatientName
		rec.PatientRutSnapshot = a.chosen.PatientDocument
		rec.ReservationId = a.chosen.ReservationId
		rec.ServiceId = a.chosen.ServiceId
	}
	return rec
}

// Screens construye la pantalla «Ficha» para el menú: el doctor elige en el selector una reserva de su día y la lista
// muestra las fichas de ese paciente (día, motivo y doctor). La agenda se carga al abrir la pantalla y se vuelve a
// leer sola cuando el servidor avisa que cambió (view.BrowserDeps.Notices).
func Screens(d view.BrowserDeps) []view.Screen {
	a := &doctorAgenda{
		caller: d.Caller,
		zone:   d.Zone,
		picker: &selectsearch.SelectSearch{Placeholder: textPickPatient, Search: selectsearch.SearchNever, StartOpen: true},
	}
	a.picker.OnSelect = func(id, _ string) { a.choose(id) }
	if d.Notices != nil {
		d.Notices.Subscribe(clinicalencounter.AgendaChanged, func(events.Event) {
			if a.opened {
				a.loadAgenda()
			}
		})
	}

	v, err := crudview.New(crudview.Config{
		ParentID:   ID,
		Presenter:  clinicalencounter.NewView(d.Caller, view.WithFilter(a.patientFilter)),
		IDs:        d.IDs,
		Scope:      a.picker,
		HideSearch: true,
		List: func(selected *dom.SignalString, onSelect func(model.Item)) crudview.ListView {
			return &targetdate.TargetDate{Zone: d.Zone, Selected: selected, OnSelect: onSelect}
		},
		OnAfterReload: a.opening,
		NewRecord:     a.newRecord,
	})
	if err != nil {
		// Error de configuración: se ve en la pantalla, nunca un menú al que le falta la ficha sin decir por qué.
		return []view.Screen{{ID: ID, Label: Label, Icon: string(svg.Icon(ID)), Component: html.P().Text(err.Error()),
			Permission: model.ResourceOf(&clinicalencounter.MedicalHistoryModel), Help: Label}}
	}
	a.list = v
	return []view.Screen{{
		ID:         ID,
		Label:      Label,
		Icon:       string(svg.Icon(ID)),
		Component:  v,
		Permission: model.ResourceOf(&clinicalencounter.MedicalHistoryModel),
		Help:       Label,
	}}
}
