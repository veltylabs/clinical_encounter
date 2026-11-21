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

// AgendaChanged es el aviso que recibe el navegador del profesional cuando cambia una reserva de su agenda: la
// pantalla vuelve a leer my_agenda_result.
const AgendaChanged = "clinical_encounter.agenda.changed"

// Estados de reserva que entrega AgendaReader (contrato con appointment_booking, valores planos).
const (
	ReservationPending    = "PENDING"
	ReservationConfirmed  = "CONFIRMED"
	ReservationCompleted  = "COMPLETED"
	ReservationNoShow     = "NO_SHOW"
	ReservationConflicted = "CONFLICTED"
)

// Attendable dice si una reserva con ese estado abre la ficha del paciente: solo la confirmada y la ya atendida.
func Attendable(reservationStatus string) bool {
	return reservationStatus == ReservationConfirmed || reservationStatus == ReservationCompleted
}
