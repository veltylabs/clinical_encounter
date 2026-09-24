package ui

import (
	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"webtyp.com/components/selectsearch"
	"webtyp.com/components/targethour"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/layout/crudview"
	"webtyp.com/layout/platformd"
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
	todayISO := tintime.FormatISO8601(tintime.Now())[:10]
	nano, err := tintime.ParseDate(todayISO)
	if err != nil {
		return 0, 0
	}
	day := nano / 1000000000
	return day, day
}

// doctorIdentity is what the logged-in user needs to act as a doctor: their
// staff_id (to look up today's confirmed reservations) and the two snapshot
// values CreateVisitArgs requires. StaffID == "" means the logged-in account
// is not itself a staff row (e.g. an administrator browsing this module) —
// not an error, just "no agenda to narrow by".
type doctorIdentity struct {
	StaffID   string
	Name      string
	Specialty string
}

// loadDoctorIdentity finds which staff row belongs to the logged-in user.
// staff_manager has no "get staff by user_id" op — a tenant's staff roster
// is small (a handful of professionals, not thousands), so listing all and
// matching UserId locally reuses the existing op instead of adding one.
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

// agendaPatient is one row of the doctor's today-agenda picker — see
// README.md. History is pre-fetched (not looked up when the picker fires):
// view.Presenter.Filter(term) is synchronous, and this data only exists
// behind a real network call, so it has to already be in memory by the time
// Filter runs.
type agendaPatient struct {
	ID          string
	Label       string
	Sublabel    string
	Description string
	History     []view.Item
}

// loadTodayAgenda lists this staff member's CONFIRMED reservations for
// today and resolves each one to a picker row — the doctor sees exactly who
// they are scheduled to see today, nothing more, and never types a name or
// a RUT (see README.md's "Error de diseño descartado").
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
			// One unresolvable patient must not take the whole agenda down —
			// it just doesn't get a row.
			continue
		}

		var history clinicalencounter.MedicalHistoryList
		_ = blockingCall(caller, clinicalencounter.ModelName+"."+clinicalencounter.OpListVisitsByPatient,
			&clinicalencounter.ListVisitsArgs{PatientId: r.ClientId}, &history)
		items := make([]view.Item, len(history))
		for i, h := range history {
			items[i] = h.Item()
		}

		agenda = append(agenda, agendaPatient{
			ID:          r.ClientId,
			Label:       patient.Name,
			Sublabel:    patient.Rut,
			Description: r.LocalStringTime,
			History:     items,
		})
	}
	return agenda, nil
}

// requirePatient wraps the generic Presenter view.New builds so
// Filter(term) means "which patient", not a free-text search — same role as
// the tinywasm/app-demo/modules/medicalhistory reference this
// productionizes (see README.md), except history comes from agenda's
// pre-fetched cache instead of a synchronous in-memory query, because here
// it is real, per-patient server data.
type requirePatient struct {
	view.Presenter
	agenda []agendaPatient
	// chosen recuerda al paciente que el picker acaba de elegir, para que
	// Config.NewRecord siembre la ficha nueva con él. Es un PUNTERO porque
	// Filter tiene receptor por valor (lo exige view.Presenter): escribir en
	// un campo del valor se perdería al volver.
	chosen *agendaPatient
}

func (p requirePatient) Filter(term string) []view.Item {
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

// Save forwards explicitly: embedding an INTERFACE only promotes the
// methods view.Presenter itself declares, never view.Saver's — see the
// identical note on the app-demo reference this follows. Update/Delete are
// deliberately NOT forwarded here (unlike that reference): the underlying
// clinicalencounter.NewView presenter supports create only (Ops has no
// Update/Delete op — a clinical encounter is append-only), and declaring
// those methods on this wrapper would make it start satisfying
// view.Updater/view.Deleter structurally, showing crudview's edit/delete
// footer buttons for an action that can never work.
func (p requirePatient) Save(recs []model.Model, done func(error)) {
	if s, ok := p.Presenter.(view.Saver); ok {
		s.Save(recs, done)
		return
	}
	done(fmt.Errf("requirePatient: underlying presenter cannot save"))
}

var _ view.Presenter = requirePatient{}
var _ view.Saver = requirePatient{}

// Browser builds this module's view for the registry in modules/browser.go.
// userID is the logged-in account (auth.ProfileDTO.Id, resolved once at
// startup by config/client.go's auth.OpMe call) — used to find which staff
// row is "the doctor" so today's agenda can be built. See README.md for the
// full design and the one still-open gap (crudview cannot pre-fill a new
// draft's fields from the picker's selection yet).
func Browser(caller router.Caller, ids model.IDGenerator, tenantID, userID string) (platformd.UIModule, error) {
	cfg := crudview.Config{
		ParentID:  ID,
		Presenter: clinicalencounter.NewView(caller),
		IDs:       ids,
		List: func(selected *dom.SignalString, onSelect func(view.Item)) crudview.ListView {
			return &targethour.TargetHour{Selected: selected, OnSelect: onSelect}
		},
	}

	doctor, err := loadDoctorIdentity(caller, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if doctor.StaffID != "" {
		agenda, err := loadTodayAgenda(caller, tenantID, doctor.StaffID)
		if err != nil {
			return nil, err
		}
		picker := &selectsearch.SelectSearch{Placeholder: "Seleccione un paciente de la agenda de hoy..."}
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
				DoctorId:                doctor.StaffID,
				DoctorNameSnapshot:      doctor.Name,
				DoctorSpecialtySnapshot: doctor.Specialty,
				Status:                  clinicalencounter.StatusCreated,
			}
			if chosen.ID != "" {
				rec.PatientId = chosen.ID
				rec.PatientNameSnapshot = chosen.Label
				rec.PatientRutSnapshot = chosen.Sublabel
			}
			return rec
		}
	}

	v, err := crudview.New(cfg)
	if err != nil {
		return nil, err
	}
	return platformd.NewUIModule(ID, Label, svg.Icon(ID), v), nil
}
