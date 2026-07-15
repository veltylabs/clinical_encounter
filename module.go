package clinical_encounter

import (
	"github.com/tinywasm/orm"
	"github.com/tinywasm/unixid"
)

type Module struct {
	DB  *orm.DB
	UID *unixid.UnixID
	Pub EventPublisher
}

func (m *Module) ModelName() string {
	return "clinical_encounter"
}
