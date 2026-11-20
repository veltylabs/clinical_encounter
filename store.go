package clinical_encounter

import "webtyp.com/crud"

// StoreMedicalHistory es el almacenamiento de medical_history.
// Ejemplo: var store StoreMedicalHistory
func (m *Module) StoreMedicalHistory() crud.Store { return m.store }

// StoreEncounterSupply es el almacenamiento de encounter_supply.
// Ejemplo: var store StoreEncounterSupply
func (m *Module) StoreEncounterSupply() crud.Store { return m.store }
