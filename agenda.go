package clinical_encounter

import (
	"webtyp.com/fmt"
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

	resIDs, patIDs, svcIDs, starts, statuses, err := m.agenda.ActiveByStaff(tenantID, staffID, from, to)
	if err != nil {
		return nil, err
	}

	today := date
	for i := range resIDs {
		e := AgendaEntry{PatientId: patIDs[i], ReservationId: resIDs[i]}
		if m.patients != nil {
			name, doc, _, _, ok, err := m.patients.GetPatient(tenantID, e.PatientId)
			if err != nil || !ok {
				continue
			}
			e.PatientName, e.PatientDocument = name, doc
			if birth, ok, err := m.patients.PatientBirthdate(tenantID, e.PatientId); err == nil && ok {
				e.PatientAge = ageOn(birth, today)
			}
		}
		if m.zone.Name() != "" && i < len(starts) {
			e.TimeLabel = m.zone.Clock(starts[i])
		}
		if i < len(svcIDs) {
			e.ServiceId = svcIDs[i]
			if m.services != nil && e.ServiceId != "" {
				if name, _, _, _, err := m.services.ServiceCard(tenantID, e.ServiceId); err == nil {
					e.ServiceName = name
				}
			}
		}
		if i < len(statuses) {
			e.ReservationStatus = statuses[i]
		}
		out.Entries = append(out.Entries, e)
	}

	return out, nil
}

// ageOn son los años cumplidos el día today por quien nació el día birth (ambos "YYYY-MM-DD"); 0 si falta alguno o
// no se pueden leer.
func ageOn(birth, today string) int64 {
	if len(birth) < 10 || len(today) < 10 {
		return 0
	}
	by, err1 := fmt.Convert(birth[0:4]).Int64()
	ty, err2 := fmt.Convert(today[0:4]).Int64()
	if err1 != nil || err2 != nil || ty < by {
		return 0
	}
	age := ty - by
	if today[5:10] < birth[5:10] { // aún no llega su cumpleaños este año
		age--
	}
	if age < 0 {
		return 0
	}
	return age
}
