package clinical_encounter

import (
	"fmt"
	"webtyp.com/events"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/time"
)

// ModelName is this module's identity: mcp.HarvestOps qualifies every op
// below as "clinical_encounter.<name>" on the wire, and view.go's NewView
// passes this same constant as view.Ops.Module so the client composes the
// identical qualified name.
const ModelName = "clinical_encounter"

// Deps son los puertos de infraestructura del módulo — nunca una implementación concreta.

// RecipeReader is the part of service_catalog this module uses.
type RecipeReader interface {
	ServiceRecipe(tenantId, serviceId string) (itemIds []string, quantities []float64, err error)
}

// RoomFinder is the part of room_layout this module uses.
type RoomFinder interface {
	RoomOfOccupant(tenantID, occupantID string, at time.Unix) (roomID string, found bool, err error)
}

// SupplyStock is the part of supply_manager this module uses.
type SupplyStock interface {
	UseInRoom(tenantId, roomId, itemId string, quantity float64, docType, docId, by string) (int64, error)
}
type Deps struct {
	Publisher    events.Publisher  // optional: nil disables events and the alert
	IDs          model.IDGenerator // required when supply use is configured (ids of the lines)
	Recipes      RecipeReader      // supply use: the three are set together or not at all
	Rooms        RoomFinder
	Supplies     SupplyStock
	SuppliesArea string            // the application's notify.Area code of the Supplies area; required with supply use
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
