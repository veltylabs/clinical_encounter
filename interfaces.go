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

// DoctorReader resuelve qué profesional clínico corresponde a una cuenta de usuario, y al revés.
// Ejemplo: staffID, name, spec, ok, err := d.StaffByUser("t1", "u1")
// Ejemplo: userID, ok, err := d.UserByStaff("t1", "doc1")
type DoctorReader interface {
	StaffByUser(tenantID, userID string) (staffID, name, specialty string, found bool, err error)
	UserByStaff(tenantID, staffID string) (userID string, found bool, err error)
}

// AgendaReader es la parte de appointment_booking que este módulo usa: las reservas activas de un profesional en
// [from, to) con su estado (los valores Reservation* de const.go), y el aviso de cada cambio de su agenda.
// Ejemplo: resIDs, patIDs, svcIDs, starts, statuses, err := a.ActiveByStaff("t1", "doc1", from, to)
// Ejemplo: a.WatchStaffAgenda(func(staffID string) { … })
type AgendaReader interface {
	ActiveByStaff(tenantID, staffID string, from, to time.Unix) (reservationIDs, patientIDs, serviceIDs []string, startTimes []time.Unix, statuses []string, err error)
	WatchStaffAgenda(fn func(staffID string))
}

// PatientReader es la parte de client_directory que este módulo usa.
// Ejemplo: name, doc, _, _, ok, err := p.GetPatient("t1", "p1")
// Ejemplo: birth, ok, err := p.PatientBirthdate("t1", "p1") // "YYYY-MM-DD", "" si no la tiene
type PatientReader interface {
	GetPatient(tenantID, id string) (name, documentNumber, phone, email string, found bool, err error)
	PatientBirthdate(tenantID, id string) (birthdate string, found bool, err error)
}

// ServiceReader es la parte de service_catalog que este módulo usa: el nombre del servicio de una reserva.
// Ejemplo: name, _, _, _, err := s.ServiceCard("t1", "svc1")
type ServiceReader interface {
	ServiceCard(tenantId, serviceId string) (name string, price float64, currency string, active bool, err error)
}

// Notifier avisa, sin datos, a los navegadores donde un usuario tiene sesión (la app pasa backend.Deps.Notices).
// Ejemplo: n.Notify(userID, AgendaChanged)
type Notifier interface {
	Notify(userID, topic string)
}
