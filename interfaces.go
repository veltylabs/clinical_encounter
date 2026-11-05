package clinical_encounter

import "webtyp.com/time"

// RecipeReader es la parte de service_catalog que este módulo utiliza.
// Ejemplo: ids, qtys, err := r.ServiceRecipe("t1", "svc1")
type RecipeReader interface {
	ServiceRecipe(tenantId, serviceId string) (itemIds []string, quantities []float64, err error)
}

// RoomFinder es la parte de room_layout que este módulo utiliza.
// Ejemplo: roomID, ok, err := f.RoomOfOccupant("t1", "doc1", at)
type RoomFinder interface {
	RoomOfOccupant(tenantID, occupantID string, at time.Unix) (roomID string, found bool, err error)
}

// SupplyStock es la parte de supply_manager que este módulo utiliza.
// Ejemplo: cost, err := s.UseInRoom("t1", "room1", "item1", 2, "encounter", "enc1", "u1")
type SupplyStock interface {
	UseInRoom(tenantId, roomId, itemId string, quantity float64, docType, docId, by string) (int64, error)
}

// DoctorReader resuelve qué profesional clínico corresponde a una cuenta de usuario.
// Ejemplo: staffID, name, spec, ok, err := d.StaffByUser("t1", "u1")
type DoctorReader interface {
	StaffByUser(tenantID, userID string) (staffID, name, specialty string, found bool, err error)
}

// AgendaReader lista las reservas confirmadas de un profesional en un rango de tiempo [from, to).
// Ejemplo: resIDs, patIDs, svcIDs, starts, err := a.ConfirmedByStaff("t1", "doc1", from, to)
type AgendaReader interface {
	ConfirmedByStaff(tenantID, staffID string, from, to time.Unix) (reservationIDs, patientIDs, serviceIDs []string, startTimes []time.Unix, err error)
}

// PatientReader es la parte de patient_directory que este módulo usa.
// Ejemplo: name, doc, _, _, ok, err := p.GetPatient("t1", "p1")
type PatientReader interface {
	GetPatient(tenantID, id string) (name, documentNumber, phone, email string, found bool, err error)
}
