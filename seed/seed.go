package seed

import (
	clinicalencounter "github.com/veltylabs/clinical_encounter"
	patientseed "github.com/veltylabs/patient_directory/seed"
	staffseed "github.com/veltylabs/staff_manager/seed"
	"webtyp.com/fmt"
	"webtyp.com/time"
)

// seedVisitAgeDays es la antigüedad de las fichas de demo.
const seedVisitAgeDays = 30

type Upstream struct {
	Patients patientseed.Data
	Staff    staffseed.Data
}

type Data struct {
	Visits []*clinicalencounter.MedicalHistory
}

func Load(m *clinicalencounter.Module, up Upstream) (Data, error) {
	if len(up.Patients.Patients) < 2 {
		return Data{}, fmt.Errf("seed: need at least 2 patients from upstream")
	}
	if len(up.Staff.Staff) < 1 {
		return Data{}, fmt.Errf("seed: need at least 1 staff member from upstream")
	}

	doc := up.Staff.Staff[0]
	patients := up.Patients.Patients[:2]
	visits := make([]*clinicalencounter.MedicalHistory, 0, len(patients))

	// time.Now() está en NANOsegundos (igual que AttentionAt, que visit.go
	// rellena con time.Now()): 30 días = 30*86400 segundos * 1e9.
	thirtyDaysAgo := time.Now() - seedVisitAgeDays*86400*1_000_000_000

	for _, p := range patients {
		v, err := m.CreateVisit(clinicalencounter.CreateVisitArgs{
			PatientId:               p.Id,
			PatientNameSnapshot:     p.Name,
			PatientRutSnapshot:      p.Rut,
			DoctorId:                doc.Id,
			DoctorNameSnapshot:      doc.Name,
			DoctorSpecialtySnapshot: doc.Specialty,
			AttentionAt:             thirtyDaysAgo,
			Reason:                  "Control",
			Diagnostic:              "Sin hallazgos",
		})
		if err != nil {
			return Data{}, fmt.Errf("seed: CreateVisit: %w", err)
		}
		visits = append(visits, v)
	}

	return Data{Visits: visits}, nil
}
