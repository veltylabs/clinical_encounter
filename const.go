package clinical_encounter

// Estado de la visita — ciclo de vida alineado con FHIR
const (
	StatusCreated    = "created"     // agendado, aún no en clínica
	StatusArrived    = "arrived"     // paciente registrado en recepción
	StatusTriaged    = "triaged"     // enfermera tomó signos vitales, esperando al doctor
	StatusInProgress = "in_progress" // doctor comenzó la consulta
	StatusCompleted  = "completed"
	StatusCancelled  = "cancelled"
)

// Estado de la línea de insumos del encuentro
const (
	SupplyPlanned          = "planned"
	SupplyDeducted         = "deducted"
	SupplyPendingDeduction = "pending_deduction"
)

const TopicVisitCompleted = "clinical_encounter.visit.completed"
