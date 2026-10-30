//go:build browser

package ui

import (
	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"webtyp.com/components/selectsearch"
	"webtyp.com/components/targethour"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/layout/crudview"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/svg"
	tintime "webtyp.com/time"
	"webtyp.com/view"

	ab "github.com/veltylabs/appointment_booking"
	patientdirectory "github.com/veltylabs/patient_directory"
	staffmanager "github.com/veltylabs/staff_manager"
)

// blockingCall awaits an async router.Caller.Call synchronously via a
// buffered channel. This is legal ONLY because Browser (below) runs inside
// BuildClient, in main()'s own call stack, BEFORE the browser event loop
// starts spinning — the exact same justification config/client.go already
// documents for auth.OpMe. It would be a silent WASM deadlock anywhere a
// callback already in flight calls this (the event loop has nothing left to
// service the response with) — see README.md's "Por qué el picker no puede
// ser síncrono" for the case that IS such a callback (Filter) and therefore
// pre-loads instead of calling this.
func blockingCall(caller router.Caller, op string, args model.Encodable, into model.Decodable) error {
	done := make(chan error, 1)
	caller.Call(op, args, into, func(err error) { done <- err })
	return <-done
}

// todayRange returns today's calendar day as UTC midnight seconds, twice —
// appointment_booking.ListReservationsByStaff filters ReservationDate
// inclusively between From and To, and a reservation's date is stored as
// that convention (see business_calendar.specificDateLabel's identical
// note): a single day is From == To.
func todayRange() (from, to int64) {
	todayISO := tintime.FormatISO8601(tintime.NowUnix())[:10]
	day, err := tintime.ParseDate(todayISO)
	if err != nil {
		return 0, 0
	}
	return int64(day), int64(day)
}

// doctorIdentity es lo que el usuario logueado necesita para actuar como un doctor: su
// staff_id (para buscar las reservas confirmadas de hoy) y los dos valores snapshot
// que requiere CreateVisitArgs. StaffID == "" significa que la cuenta logueada
// no es en sí misma una fila de personal (ej. un administrador navegando en este módulo) —
// no es un error, solo "sin agenda por la cual filtrar".
type doctorIdentity struct {
	StaffID   string
	Name      string
	Specialty string
}

// loadDoctorIdentity encuentra qué fila de personal pertenece al usuario logueado.
// staff_manager no tiene operación "obtener personal por user_id" — la lista
// de personal de un tenant es pequeña (un puñado de profesionales, no miles), así que
// listar todos y hacer match del UserId localmente reutiliza la operación existente en lugar de agregar una.
func loadDoctorIdentity(caller router.Caller, tenantID, userID string) (doctorIdentity, error) {
	var staff staffmanager.StaffMemberList
	err := blockingCall(caller, staffmanager.ModelName+"."+staffmanager.OpListStaff,
		&staffmanager.ListStaffArgs{TenantId: tenantID}, &staff)
	if err != nil {
		return doctorIdentity{}, err
	}
	for _, s := range staff {
		if s.UserId == userID {
			return doctorIdentity{StaffID: s.Id, Name: s.Name, Specialty: s.Specialty}, nil
		}
	}
	return doctorIdentity{}, nil
}

// agendaPatient es una fila del picker de la agenda del doctor de hoy — ver
// docs/UI.md. El historial se pre-busca (no se busca cuando el picker se dispara):
// view.Presenter.Filter(term) es síncrono, y estos datos solo existen
// detrás de una llamada de red real, por lo que ya tiene que estar en memoria para
// el momento que Filter se ejecute.
type agendaPatient struct {
	ID            string
	Label         string
	Sublabel      string
	Description   string
	ReservationID string
	ServiceID     string
	History       []model.Item
}

// loadTodayAgenda lista las reservas CONFIRMED de este miembro del personal para
// hoy y resuelve cada una a una fila del picker — el doctor ve exactamente a
// quién tiene programado ver hoy, nada más, y nunca escribe un nombre o
// un RUT (ver docs/UI.md "Error de diseño descartado").
func loadTodayAgenda(caller router.Caller, tenantID, staffID string) ([]agendaPatient, error) {
	from, to := todayRange()
	var reservations ab.ReservationList
	err := blockingCall(caller, ab.ModelName+"."+ab.OpListReservationsByStaff,
		&ab.ListReservationsByStaffArgs{TenantId: tenantID, StaffId: staffID, From: from, To: to},
		&reservations)
	if err != nil {
		return nil, err
	}

	agenda := make([]agendaPatient, 0, len(reservations))
	for _, r := range reservations {
		if r.Status != ab.StatusConfirmed {
			continue
		}

		var patient patientdirectory.Patient
		if err := blockingCall(caller, patientdirectory.ModelName+"."+patientdirectory.OpGetPatient,
			&patientdirectory.GetPatientArgs{TenantId: tenantID, Id: r.ClientId}, &patient); err != nil {
			// Un paciente irresoluble no debe botar toda la agenda —
			// simplemente no obtiene una fila.
			continue
		}

		var history clinicalencounter.MedicalHistoryList
		_ = blockingCall(caller, clinicalencounter.ModelName+"."+clinicalencounter.OpListVisitsByPatient,
			&clinicalencounter.ListVisitsArgs{TenantId: tenantID, PatientId: r.ClientId}, &history)
		items := make([]model.Item, len(history))
		for i, h := range history {
			items[i] = h.Item()
		}

		agenda = append(agenda, agendaPatient{
			ID:            r.ClientId,
			Label:         patient.Name,
			Sublabel:      patient.Rut,
			Description:   r.LocalStringTime,
			ReservationID: r.Id,
			ServiceID:     r.EmployeeServiceConfigId,
			History:       items,
		})
	}
	return agenda, nil
}

// requirePatient envuelve al Presenter genérico que view.New construye para
// que Filter(term) signifique "qué paciente", no una búsqueda de texto libre —
// mismo rol que la referencia tinywasm/app-demo/modules/medicalhistory que
// esto lleva a producción (ver docs/UI.md), excepto que el historial proviene del
// caché pre-buscado de agenda en lugar de una consulta síncrona en-memoria, porque aquí
// son datos de servidor reales, por paciente.
type requirePatient struct {
	view.Presenter
	agenda []agendaPatient
	// chosen recuerda al paciente que el picker acaba de elegir, para que
	// Config.NewRecord siembre la ficha nueva con él. Es un PUNTERO porque
	// Filter tiene receptor por valor (lo exige view.Presenter): escribir en
	// un campo del valor se perdería al volver.
	chosen *agendaPatient
}

func (p requirePatient) Filter(term string) []model.Item {
	if term == "" {
		if p.chosen != nil {
			*p.chosen = agendaPatient{}
		}
		return nil
	}
	for _, a := range p.agenda {
		if a.ID == term {
			if p.chosen != nil {
				*p.chosen = a
			}
			return a.History
		}
	}
	return nil
}

// Save redirige explícitamente: embeber una INTERFAZ solo promueve los
// métodos que view.Presenter mismo declara, nunca los de view.Saver — ver la
// nota idéntica en la referencia app-demo que esto sigue. Update/Delete
// deliberadamente NO son redirigidos aquí (a diferencia de esa referencia): el
// presentador clinicalencounter.NewView subyacente soporta create solamente (Ops no tiene
// op de Update/Delete — un encuentro clínico es solo para anexar), y declarar
// esos métodos en este envoltorio lo haría empezar a satisfacer
// view.Updater/view.Deleter estructuralmente, mostrando los botones footer de edit/delete
// del crudview para una acción que nunca podría funcionar.
func (p requirePatient) Save(recs []model.Model, done func(error)) {
	if s, ok := p.Presenter.(view.Saver); ok {
		s.Save(recs, done)
		return
	}
	done(fmt.Errf("requirePatient: underlying presenter cannot save"))
}

var _ view.Presenter = requirePatient{}
var _ view.Saver = requirePatient{}

// Screens construye las pantallas de este módulo para el menú de la aplicación.
// userID es la cuenta logueada (auth.ProfileDTO.Id, resuelta una vez al
// inicio mediante la llamada auth.OpMe de config/client.go) — usada para encontrar qué
// fila de personal es "el doctor" para que la agenda de hoy pueda ser construida. Ver docs/UI.md para
// el diseño completo.
// Ejemplo: ui.Screens(caller, ids, "tenant1", "user1")
func Screens(caller router.Caller, ids model.IDGenerator, tenantID, userID string) []view.Screen {
	cfg := crudview.Config{
		ParentID:  ID,
		Presenter: clinicalencounter.NewView(caller),
		IDs:       ids,
		List: func(selected *dom.SignalString, onSelect func(model.Item)) crudview.ListView {
			return &targethour.TargetHour{Selected: selected, OnSelect: onSelect}
		},
	}

	doctor, err := loadDoctorIdentity(caller, tenantID, userID)
	if err != nil {
		return nil
	}
	if doctor.StaffID != "" {
		agenda, err := loadTodayAgenda(caller, tenantID, doctor.StaffID)
		if err != nil {
			return nil
		}
		picker := &selectsearch.SelectSearch{Placeholder: "Select a patient from today.s agenda..."}
		opts := make([]selectsearch.SsOption, len(agenda))
		for i, a := range agenda {
			opts[i] = selectsearch.SsOption{ID: a.ID, Label: a.Label, Sublabel: a.Sublabel, Description: a.Description}
		}
		picker.Options = opts

		// El paciente elegido viaja del picker al borrador nuevo por acá:
		// crudview llama a Filter(term) cuando el picker cambia, y a NewRecord
		// cuando se pulsa "+". Sin esto el doctor reescribía a mano el id, el
		// nombre y el RUT de alguien que acababa de seleccionar en pantalla.
		chosen := &agendaPatient{}
		cfg.Presenter = requirePatient{Presenter: cfg.Presenter, agenda: agenda, chosen: chosen}
		cfg.Filter = picker

		// Una instancia NUEVA en cada llamada, nunca una compartida: crudview
		// documenta que un registro reutilizado arrastraría las ediciones del
		// borrador anterior al siguiente.
		//
		// Sin paciente elegido devuelve un registro vacío — no un error ni un
		// panic: pulsar "+" antes de elegir es una acción legítima, y el
		// formulario simplemente sale en blanco como antes.
		cfg.NewRecord = func() model.Model {
			rec := &clinicalencounter.MedicalHistory{
				TenantId:                tenantID,
				DoctorId:                doctor.StaffID,
				DoctorNameSnapshot:      doctor.Name,
				DoctorSpecialtySnapshot: doctor.Specialty,
				Status:                  clinicalencounter.StatusCreated,
			}
			if chosen.ID != "" {
				rec.PatientId = chosen.ID
				rec.PatientNameSnapshot = chosen.Label
				rec.PatientRutSnapshot = chosen.Sublabel
				rec.ReservationId = chosen.ReservationID
				rec.ServiceId = chosen.ServiceID
			}
			return rec
		}
	}

	v, err := crudview.New(cfg)
	if err != nil {
		return nil
	}
	return []view.Screen{{
		ID:         ID,
		Label:      Label,
		Icon:       string(svg.Icon(ID)),
		Component:  v,
		Permission: model.ResourceOf(&clinicalencounter.MedicalHistoryModel),
		Help:       Label,
	}}
}
