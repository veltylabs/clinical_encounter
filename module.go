package clinical_encounter

import (
	"github.com/tinywasm/ddl"
	"github.com/tinywasm/events"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/model"
	"github.com/tinywasm/orm"
)

// Deps are the module's infrastructure ports — never a concrete implementation.
type Deps struct {
	IDs       model.IDGenerator // required — the module never builds its own
	Publisher events.Publisher  // optional — nil disables publishing silently
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
	// ddl.Compiler is an optional capability — only SQL backends (sqlt, postgres) implement it.
	// storage/mem (this module's own tests, Stage 6) creates tables lazily and needs no DDL, so a
	// type assertion — not an unconditional call — is how the module stays backend-agnostic here.
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
