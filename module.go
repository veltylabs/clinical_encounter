package clinical_encounter

import (
	"webtyp.com/events"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/time"
)

// ModelName es la identidad de este módulo: mcp.HarvestOps califica cada op
// debajo como "clinical_encounter.<name>" en el cable, y NewView de view.go
// pasa esta misma constante como view.Ops.Module para que el cliente componga el
// nombre calificado idéntico.
// Ejemplo: ModelName == "clinical_encounter"

type Deps struct {
	Publisher    events.Publisher  // opcional: nil deshabilita eventos y la alerta
	IDs          model.IDGenerator // requerido cuando el uso de insumos está configurado (ids de las líneas)
	Recipes      RecipeReader      // uso de insumos: los tres se configuran juntos o en lo absoluto
	Rooms        RoomFinder
	Supplies     SupplyStock
	SuppliesArea string // el código notify.Area de la aplicación del área de Insumos; requerido con uso de insumos
	TenantID     string
	Zone         time.Zone
	Doctors      DoctorReader
	Agenda       AgendaReader
	Patients     PatientReader
}

type Module struct {
	store        *orm.Store
	db           *orm.DB
	pub          events.Publisher
	supplyOn     bool
	ids          model.IDGenerator
	recipes      RecipeReader
	rooms        RoomFinder
	supplies     SupplyStock
	suppliesArea string
	tenantID     string
	zone         time.Zone
	doctors      DoctorReader
	agenda       AgendaReader
	patients     PatientReader
}

func New(db *orm.DB, deps Deps) (*Module, error) {
	m := &Module{
		store:        orm.TenantStore(db, deps.TenantID, deps.IDs),
		db:           db,
		pub:          deps.Publisher,
		ids:          deps.IDs,
		recipes:      deps.Recipes,
		rooms:        deps.Rooms,
		supplies:     deps.Supplies,
		suppliesArea: deps.SuppliesArea,
		tenantID:     deps.TenantID,
		zone:         deps.Zone,
		doctors:      deps.Doctors,
		agenda:       deps.Agenda,
		patients:     deps.Patients,
	}

	hasRecipes := deps.Recipes != nil
	hasRooms := deps.Rooms != nil
	hasSupplies := deps.Supplies != nil

	if hasRecipes && hasRooms && hasSupplies {
		if deps.IDs == nil {
			return nil, fmt.Errf("clinical_encounter: supply use is configured but IDs generator is missing")
		}
		if deps.SuppliesArea == "" {
			return nil, fmt.Errf("clinical_encounter: supply use is configured but SuppliesArea is empty")
		}
		m.supplyOn = true
	} else if hasRecipes || hasRooms || hasSupplies {
		return nil, fmt.Errf("clinical_encounter: supply use requires Recipes, Rooms and Supplies together")
	}

	return m, nil
}
