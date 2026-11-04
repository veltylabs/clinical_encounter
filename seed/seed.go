package seed

import (
	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"webtyp.com/fmt"
	"webtyp.com/time"
)

// seedVisitAgeDays es la antigüedad de las fichas de demo.
const seedVisitAgeDays = 30

type PatientSeed struct {
	ID             string
	Name           string
	DocumentNumber string
}

type DoctorSeed struct {
	ID        string
	Name      string
	Specialty string
}

type LoadOptions struct {
	Patients []PatientSeed
	Doctors  []DoctorSeed
}

type Data struct {
	Visits []*clinicalencounter.MedicalHistory
}

func Load(m *clinicalencounter.Module, tenantID string, opts ...LoadOptions) (Data, error) {
	var opt LoadOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	patients := opt.Patients
	doctors := opt.Doctors

	if len(opts) > 0 && (len(patients) < 2 || len(doctors) < 1) {
		return Data{}, fmt.Errf("seed: need at least 2 patients and 1 doctor from upstream")
	}

	if len(patients) == 0 {
		patients = []PatientSeed{
			{ID: "demo-pat-1", Name: "Juan Pérez", DocumentNumber: "12.345.678-9"},
			{ID: "demo-pat-2", Name: "María López", DocumentNumber: "9.876.543-2"},
		}
	}
	if len(doctors) == 0 {
		doctors = []DoctorSeed{
			{ID: "demo-doc-1", Name: "Dra. Soto", Specialty: "Pediatría"},
		}
	}

	doc := doctors[0]
	visits := make([]*clinicalencounter.MedicalHistory, 0, len(patients))
	thirtyDaysAgo := time.NowUnix() - seedVisitAgeDays*time.Day

	for i, p := range patients {
		if i >= 2 {
			break
		}
		id := "demo-visit-" + fmt.Convert(i+1).String()
		v, err := m.CreateVisit(clinicalencounter.CreateVisitArgs{
			TenantId:                tenantID,
			Id:                      id,
			PatientId:               p.ID,
			PatientNameSnapshot:     p.Name,
			PatientRutSnapshot:      p.DocumentNumber,
			DoctorId:                doc.ID,
			DoctorNameSnapshot:      doc.Name,
			DoctorSpecialtySnapshot: doc.Specialty,
			AttentionAt:             int64(thirtyDaysAgo),
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
