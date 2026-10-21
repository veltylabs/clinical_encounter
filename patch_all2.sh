# 3. const.go
cat << 'C_EOF' >> const.go

// Encounter supply line status
const (
	SupplyPlanned          = "planned"
	SupplyDeducted         = "deducted"
	SupplyPendingDeduction = "pending_deduction"
)

const TopicVisitCompleted = "clinical_encounter.visit.completed"
C_EOF

# 4. module.go
cat << 'PY_EOF' > patch_module.py
with open("module.go", "r") as f:
    text = f.read()

text = text.replace('import (\n\t"webtyp.com/events"\n\t"webtyp.com/orm"\n)', 'import (\n\t"fmt"\n\t"webtyp.com/events"\n\t"webtyp.com/model"\n\t"webtyp.com/orm"\n\t"webtyp.com/time"\n)')

interfaces = """
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
"""

deps = """type Deps struct {
	Publisher    events.Publisher  // optional: nil disables events and the alert
	IDs          model.IDGenerator // required when supply use is configured (ids of the lines)
	Recipes      RecipeReader      // supply use: the three are set together or not at all
	Rooms        RoomFinder
	Supplies     SupplyStock
	SuppliesArea string            // the application's notify.Area code of the Supplies area; required with supply use
}"""

text = text.replace("""type Deps struct {
	Publisher events.Publisher  // opcional — nil deshabilita la publicación silenciosamente
}""", interfaces + deps)

modStruct = """type Module struct {
	db           *orm.DB
	pub          events.Publisher
	supplyOn     bool
	ids          model.IDGenerator
	recipes      RecipeReader
	rooms        RoomFinder
	supplies     SupplyStock
	suppliesArea string
}"""

text = text.replace("""type Module struct {
	db  *orm.DB
	pub events.Publisher
}""", modStruct)

newFunc = """func New(db *orm.DB, deps Deps) (*Module, error) {
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
}"""

text = text.replace("""func New(db *orm.DB, deps Deps) (*Module, error) {
	return &Module{db: db, pub: deps.Publisher}, nil
}""", newFunc)

with open("module.go", "w") as f:
    f.write(text)
PY_EOF
python3 patch_module.py
