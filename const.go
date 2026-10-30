package clinical_encounter

// Visit status — FHIR-aligned lifecycle
const (
	StatusCreated    = "created"     // agendado, aún no en clínica
	StatusArrived    = "arrived"     // paciente registrado en recepción
	StatusTriaged    = "triaged"     // enfermera tomó signos vitales, esperando al doctor
	StatusInProgress = "in_progress" // doctor comenzó la consulta
	StatusCompleted  = "completed"
	StatusCancelled  = "cancelled"
)

// Encounter supply line status
const (
	SupplyPlanned          = "planned"
	SupplyDeducted         = "deducted"
	SupplyPendingDeduction = "pending_deduction"
)

const TopicVisitCompleted = "clinical_encounter.visit.completed"
