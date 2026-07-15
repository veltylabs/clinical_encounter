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
