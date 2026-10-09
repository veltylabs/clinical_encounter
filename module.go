package clinical_encounter

import (
	"webtyp.com/events"
	"webtyp.com/orm"
)

// ModelName is this module's identity: mcp.HarvestOps qualifies every op
// below as "clinical_encounter.<name>" on the wire, and view.go's NewView
// passes this same constant as view.Ops.Module so the client composes the
// identical qualified name.
const ModelName = "clinical_encounter"

// Deps son los puertos de infraestructura del módulo — nunca una implementación concreta.
type Deps struct {
	Publisher events.Publisher  // opcional — nil deshabilita la publicación silenciosamente
}

type Module struct {
	db  *orm.DB
	pub events.Publisher
}

func New(db *orm.DB, deps Deps) (*Module, error) {
	return &Module{db: db, pub: deps.Publisher}, nil
}

func (m *Module) ModelName() string {
	return ModelName
}
