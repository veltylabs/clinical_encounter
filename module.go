package clinical_encounter

import (
	"github.com/tinywasm/ddl"
	"github.com/tinywasm/events"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/model"
	"github.com/tinywasm/orm"
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
	// ddl.Compiler es una capacidad opcional — solo los backends SQL (sqlt, postgres) la implementan.
	// storage/mem (las pruebas propias de este módulo) crea tablas de forma perezosa y no necesita DDL,
	// así que una aserción de tipo — no una llamada incondicional — es cómo el módulo se mantiene agnóstico aquí.
	if ddlCompiler, ok := db.RawConn().(ddl.Compiler); ok {
		if err := ddl.New(db.RawConn(), ddlCompiler).CreateTable(&MedicalHistory{}); err != nil {
			return nil, err
		}
	}
	return &Module{db: db, ids: deps.IDs, pub: deps.Publisher}, nil
}

func (m *Module) ModelName() string {
	return "clinical_encounter"
}
