package clinical_encounter

import (
	"webtyp.com/events"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/orm"
)

// Deps son los puertos de infraestructura del módulo — nunca una implementación concreta.
type Deps struct {
	IDs       model.IDGenerator // requerido — el módulo nunca lo construye por sí mismo
	Publisher events.Publisher  // opcional — nil deshabilita la publicación silenciosamente
}

type Module struct {
	db  *orm.DB
	ids model.IDGenerator
	pub events.Publisher
}

func New(db *orm.DB, deps Deps) (*Module, error) {
	if deps.IDs == nil {
		return nil, fmt.Err("clinical_encounter: Deps.IDs is required")
	}
	return &Module{db: db, ids: deps.IDs, pub: deps.Publisher}, nil
}

func (m *Module) ModelName() string {
	return "clinical_encounter"
}
