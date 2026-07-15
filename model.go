package clinical_encounter

type MedicalHistory struct {
	ID                      string
	PatientID               string `db:"not_null"`
	DoctorID                string `db:"not_null"`
	ReservationID           string // soft ref — optional
	Status                  string `db:"not_null"` // see Status* constants
	AttentionAt             int64  `db:"not_null"` // Unix timestamp (date + time unified)
	Reason                  string `db:"not_null"`
	Diagnostic              string
	Prescription            string
	Cie10Code               string // Código de Clasificación Internacional de Enfermedades (diagnóstico)
	StartedAt               int64
	FinishedAt              int64
	PatientNameSnapshot     string `db:"not_null"`
	PatientRutSnapshot      string `db:"not_null"`
	DoctorNameSnapshot      string `db:"not_null"`
	DoctorSpecialtySnapshot string
	UpdatedAt               int64 `db:"not_null"`
}

type CreateVisitArgs struct {
	PatientID               string
	DoctorID                string
	AttentionAt             int64
	Reason                  string
	PatientNameSnapshot     string
	PatientRutSnapshot      string
	DoctorNameSnapshot      string
	ReservationID           string
	Diagnostic              string
	Prescription            string
	DoctorSpecialtySnapshot string
}
