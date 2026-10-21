package clinical_encounter

// Visit status — FHIR-aligned lifecycle
const (
	StatusCreated    = "created"     // scheduled, not yet at clinic
	StatusArrived    = "arrived"     // patient registered at reception
	StatusTriaged    = "triaged"     // nurse took vitals, waiting for doctor
	StatusInProgress = "in_progress" // doctor started the consultation
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
