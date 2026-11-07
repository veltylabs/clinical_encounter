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
	"webtyp.com/time"
	"webtyp.com/view"
)

// blockingCall awaits an async router.Caller.Call synchronously via a buffered channel.
func blockingCall(caller router.Caller, op string, args model.Encodable, into model.Decodable) error {
	done := make(chan error, 1)
	caller.Call(op, args, into, func(err error) { done <- err })
	return <-done
}

type agendaPatient struct {
	ID            string
	Label         string
	Sublabel      string
	Description   string
	ReservationID string
	ServiceID     string
	History       []model.Item
}

type requirePatient struct {
	view.Presenter
	agenda []agendaPatient
	chosen *agendaPatient
}

func (p requirePatient) Search(term string, done func(error)) {
	if term == "" {
		if p.chosen != nil {
			*p.chosen = agendaPatient{}
		}
		if done != nil {
			done(nil)
		}
		return
	}
	for _, a := range p.agenda {
		if a.ID == term {
			if p.chosen != nil {
				*p.chosen = a
			}
			if done != nil {
				done(nil)
			}
			return
		}
	}
	if p.chosen != nil {
		*p.chosen = agendaPatient{}
	}
	if done != nil {
		done(nil)
	}
}

func (p requirePatient) Items() []model.Item {
	if p.chosen == nil || p.chosen.ID == "" {
		return nil
	}
	return p.chosen.History
}

func (p requirePatient) HasMore() bool { return false }
func (p requirePatient) More(done func(error)) {
	if done != nil {
		done(nil)
	}
}

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
func Screens(caller router.Caller, zone time.Zone, ids model.IDGenerator, tenantID, userID string) []view.Screen {
	cfg := crudview.Config{
		ParentID:  ID,
		Presenter: clinicalencounter.NewView(caller),
		IDs:       ids,
		List: func(selected *dom.SignalString, onSelect func(model.Item)) crudview.ListView {
			return &targethour.TargetHour{Selected: selected, OnSelect: onSelect}
		},
	}

	dateStr := ""
	if zone.Name() != "" {
		dateStr = zone.Date(time.NowUnix())
	}

	var myAgenda clinicalencounter.MyAgendaResult
	err := blockingCall(caller, clinicalencounter.ModelName+"."+clinicalencounter.OpMyAgenda,
		&clinicalencounter.MyAgendaArgs{TenantId: tenantID, UserId: userID, Date: dateStr}, &myAgenda)
	if err != nil {
		return nil
	}

	if myAgenda.StaffId != "" {
		agenda := make([]agendaPatient, 0, len(myAgenda.Entries))
		for _, e := range myAgenda.Entries {
			var history clinicalencounter.MedicalHistoryList
			_ = blockingCall(caller, clinicalencounter.ModelName+"."+clinicalencounter.OpListVisitsByPatient,
				&view.ListArgs{Search: e.PatientId}, &history)
			items := make([]model.Item, len(history))
			for i, h := range history {
				items[i] = h.Item()
			}

			agenda = append(agenda, agendaPatient{
				ID:            e.PatientId,
				Label:         e.PatientName,
				Sublabel:      e.PatientDocument,
				Description:   e.TimeLabel,
				ReservationID: e.ReservationId,
				ServiceID:     e.ServiceId,
				History:       items,
			})
		}

		picker := &selectsearch.SelectSearch{Placeholder: "Select a patient from today's agenda..."}
		opts := make([]selectsearch.SsOption, len(agenda))
		for i, a := range agenda {
			opts[i] = selectsearch.SsOption{ID: a.ID, Label: a.Label, Sublabel: a.Sublabel, Description: a.Description}
		}
		picker.Options = opts

		chosen := &agendaPatient{}
		cfg.Presenter = requirePatient{Presenter: cfg.Presenter, agenda: agenda, chosen: chosen}
		cfg.Filter = picker

		cfg.NewRecord = func() model.Model {
			rec := &clinicalencounter.MedicalHistory{
				TenantId:                tenantID,
				DoctorId:                myAgenda.StaffId,
				DoctorNameSnapshot:      myAgenda.DoctorName,
				DoctorSpecialtySnapshot: myAgenda.DoctorSpecialty,
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
