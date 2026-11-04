package clinical_encounter

import (
	"webtyp.com/time"
)

// MyAgenda resuelve la agenda del día para un profesional identificado por su usuario.
func (m *Module) MyAgenda(args MyAgendaArgs) (*MyAgendaResult, error) {
	tenantID := args.TenantId
	if tenantID == "" {
		tenantID = m.tenantID
	}
	if m.doctors == nil || args.UserId == "" {
		return &MyAgendaResult{}, nil
	}

	staffID, docName, docSpec, found, err := m.doctors.StaffByUser(tenantID, args.UserId)
	if err != nil {
		return nil, err
	}
	if !found || staffID == "" {
		return &MyAgendaResult{}, nil
	}

	date := args.Date
	if date == "" && m.zone.Name() != "" {
		date = m.zone.Date(time.NowUnix())
	}

	var from, to time.Unix
	if m.zone.Name() != "" {
		from, to, err = m.zone.Day(date)
		if err != nil {
			return nil, ErrMissingArgs
		}
	} else if date != "" {
		dayKey, err := time.ParseDateUTC(date)
		if err != nil {
			return nil, ErrMissingArgs
		}
		from, to = dayKey, dayKey+time.Day
	}

	out := &MyAgendaResult{
		StaffId:         staffID,
		DoctorName:      docName,
		DoctorSpecialty: docSpec,
		Entries:         make([]AgendaEntry, 0),
	}

	if m.agenda == nil {
		return out, nil
	}

	resIDs, patIDs, svcIDs, starts, err := m.agenda.ConfirmedByStaff(tenantID, staffID, from, to)
	if err != nil {
		return nil, err
	}

	for i := range resIDs {
		patID := patIDs[i]
		patName := ""
		patDoc := ""
		if m.patients != nil {
			name, doc, ok, err := m.patients.PatientIdentity(tenantID, patID)
			if err != nil || !ok {
				continue
			}
			patName = name
			patDoc = doc
		}

		timeLabel := ""
		if m.zone.Name() != "" && i < len(starts) {
			timeLabel = m.zone.Clock(starts[i])
		}

		svcID := ""
		if i < len(svcIDs) {
			svcID = svcIDs[i]
		}

		out.Entries = append(out.Entries, AgendaEntry{
			PatientId:       patID,
			PatientName:     patName,
			PatientDocument: patDoc,
			TimeLabel:       timeLabel,
			ReservationId:   resIDs[i],
			ServiceId:       svcID,
		})
	}

	return out, nil
}
