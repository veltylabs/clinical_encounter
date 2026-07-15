//go:build wasm

package mockdb

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/time"

	"github.com/veltylabs/clinical_encounter"
)

// MemExecutor implements orm.Executor for in-memory storage.
type MemExecutor struct {
	data []*clinical_encounter.MedicalHistory
}

// MemCompiler implements orm.Compiler for in-memory storage.
type MemCompiler struct{}

// New creates a DB with mock data for clinical_encounter.
func New() *orm.DB {
	exec := &MemExecutor{
		data: mockData(),
	}
	return orm.New(exec, &MemCompiler{})
}

// Exec implements orm.Executor.
func (e *MemExecutor) Exec(query string, args ...any) error {
	return nil
}

// QueryRow implements orm.Executor.
func (e *MemExecutor) QueryRow(query string, args ...any) orm.Scanner {
	return &MemScanner{}
}

// Query implements orm.Executor.
func (e *MemExecutor) Query(query string, args ...any) (orm.Rows, error) {
	return &MemRows{}, nil
}

// Close implements orm.Executor.
func (e *MemExecutor) Close() error {
	return nil
}

// Compile implements orm.Compiler.
func (c *MemCompiler) Compile(q orm.Query, m fmt.Model) (orm.Plan, error) {
	return orm.Plan{Query: "MOCK"}, nil
}

// MemScanner implements orm.Scanner.
type MemScanner struct{}

func (s *MemScanner) Scan(dest ...any) error {
	return nil
}

// MemRows implements orm.Rows.
type MemRows struct {
	idx  int
	data []*clinical_encounter.MedicalHistory
}

func (r *MemRows) Next() bool {
	return r.idx < len(r.data)
}

func (r *MemRows) Scan(dest ...any) error {
	return nil
}

func (r *MemRows) Close() error {
	return nil
}

func (r *MemRows) Err() error {
	return nil
}

// mockData returns 4 hardcoded MedicalHistory records.
func mockData() []*clinical_encounter.MedicalHistory {
	now := time.Now()
	return []*clinical_encounter.MedicalHistory{
		{
			ID:                      "mh_001",
			PatientID:               "pat_001",
			DoctorID:                "doc_001",
			Status:                  clinical_encounter.StatusCompleted,
			AttentionAt:             now - 42*24*3600,
			Reason:                  "Control rutinario",
			Diagnostic:              "Paciente estable",
			Prescription:            "Continuar medicación",
			PatientNameSnapshot:     "Juan Pérez González",
			PatientRutSnapshot:      "12.345.678-9",
			DoctorNameSnapshot:      "Dr. Rodrigo Soto",
			DoctorSpecialtySnapshot: "Medicina General",
			UpdatedAt:               now,
		},
		{
			ID:                      "mh_002",
			PatientID:               "pat_001",
			DoctorID:                "doc_001",
			Status:                  clinical_encounter.StatusCompleted,
			AttentionAt:             now - 67*24*3600,
			Reason:                  "Dolor de cabeza",
			Diagnostic:              "Migraña tensional",
			Prescription:            "Ibuporofeno 400mg",
			PatientNameSnapshot:     "Juan Pérez González",
			PatientRutSnapshot:      "12.345.678-9",
			DoctorNameSnapshot:      "Dr. Rodrigo Soto",
			DoctorSpecialtySnapshot: "Medicina General",
			UpdatedAt:               now,
		},
		{
			ID:                      "mh_003",
			PatientID:               "pat_001",
			DoctorID:                "doc_001",
			Status:                  clinical_encounter.StatusInProgress,
			AttentionAt:             now - 21*24*3600,
			Reason:                  "Revisión exámenes",
			Diagnostic:              "Pendiente análisis",
			Prescription:            "",
			PatientNameSnapshot:     "Juan Pérez González",
			PatientRutSnapshot:      "12.345.678-9",
			DoctorNameSnapshot:      "Dr. Rodrigo Soto",
			DoctorSpecialtySnapshot: "Medicina General",
			UpdatedAt:               now,
		},
		{
			ID:                      "mh_004",
			PatientID:               "pat_001",
			DoctorID:                "doc_001",
			Status:                  clinical_encounter.StatusArrived,
			AttentionAt:             now,
			Reason:                  "Consulta general",
			Diagnostic:              "",
			Prescription:            "",
			PatientNameSnapshot:     "Juan Pérez González",
			PatientRutSnapshot:      "12.345.678-9",
			DoctorNameSnapshot:      "Dr. Rodrigo Soto",
			DoctorSpecialtySnapshot: "Medicina General",
			UpdatedAt:               now,
		},
	}
}
