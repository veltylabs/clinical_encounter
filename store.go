package clinical_encounter

import "webtyp.com/crud"

// StoreMedicalHistory is the store of medical_history.
func (m *Module) StoreMedicalHistory() crud.Store { return m.store }

// StoreEncounterSupply is the store of encounter_supply.
func (m *Module) StoreEncounterSupply() crud.Store { return m.store }
