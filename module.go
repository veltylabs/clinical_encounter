package clinical_encounter

import (
	"fmt"
	"webtyp.com/events"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/time"
)

// ModelName es la identidad de este módulo: mcp.HarvestOps califica cada op
// debajo como "clinical_encounter.<name>" en el cable, y NewView de view.go
// pasa esta misma constante como view.Ops.Module para que el cliente componga el
// nombre calificado idéntico.
// Ejemplo: ModelName == "clinical_encounter"
const ModelName = "clinical_encounter"

// Deps son los puertos de infraestructura del módulo — nunca una implementación concreta.

// RecipeReader es la parte de service_catalog que este módulo utiliza.
// Ejemplo: type MyRecipeReader struct{}
type RecipeReader interface {
	ServiceRecipe(tenantId, serviceId string) (itemIds []string, quantities []float64, err error)
}

// RoomFinder es la parte de room_layout que este módulo utiliza.
// Ejemplo: type MyRoomFinder struct{}
type RoomFinder interface {
	RoomOfOccupant(tenantID, occupantID string, at time.Unix) (roomID string, found bool, err error)
}

// SupplyStock es la parte de supply_manager que este módulo utiliza.
// Ejemplo: type MySupplyStock struct{}
type SupplyStock interface {
	UseInRoom(tenantId, roomId, itemId string, quantity float64, docType, docId, by string) (int64, error)
}
type Deps struct {
	Publisher    events.Publisher  // opcional: nil deshabilita eventos y la alerta
	IDs          model.IDGenerator // requerido cuando el uso de insumos está configurado (ids de las líneas)
	Recipes      RecipeReader      // uso de insumos: los tres se configuran juntos o en lo absoluto
	Rooms        RoomFinder
	Supplies     SupplyStock
	SuppliesArea string            // el código notify.Area de la aplicación del área de Insumos; requerido con uso de insumos
}

type Module struct {
	db           *orm.DB
	pub          events.Publisher
	supplyOn     bool
	ids          model.IDGenerator
	recipes      RecipeReader
	rooms        RoomFinder
	supplies     SupplyStock
	suppliesArea string
}

func New(db *orm.DB, deps Deps) (*Module, error) {
	m := &Module{
		db:           db,
		pub:          deps.Publisher,
		ids:          deps.IDs,
		recipes:      deps.Recipes,
		rooms:        deps.Rooms,
		supplies:     deps.Supplies,
		suppliesArea: deps.SuppliesArea,
	}

	hasRecipes := deps.Recipes != nil
	hasRooms := deps.Rooms != nil
	hasSupplies := deps.Supplies != nil

	if hasRecipes && hasRooms && hasSupplies {
		if deps.IDs == nil {
			return nil, fmt.Errorf("clinical_encounter: supply use is configured but IDs generator is missing")
		}
		if deps.SuppliesArea == "" {
			return nil, fmt.Errorf("clinical_encounter: supply use is configured but SuppliesArea is empty")
		}
		m.supplyOn = true
	} else if hasRecipes || hasRooms || hasSupplies {
		return nil, fmt.Errorf("clinical_encounter: supply use requires Recipes, Rooms and Supplies together")
	}

	return m, nil
}

func (m *Module) ModelName() string {
	return ModelName
}
