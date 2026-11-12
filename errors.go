package clinical_encounter

import (
	"webtyp.com/fmt"
	"webtyp.com/model"
)

var (
	ErrSupplyNotFound = model.NotFound(fmt.Err("supply", "use", "line", "not", "found").Error())
	ErrVisitClosed    = model.Conflict(fmt.Err("cannot", "modify", "supply", "use", "of", "a", "closed", "visit").Error())
	ErrSupplyUseOff   = model.Invalid(fmt.Err("supply", "use", "is", "not", "enabled").Error())
	ErrLineDeducted   = model.Conflict(fmt.Err("cannot", "modify", "a", "deducted", "line").Error())
)

const (
	ErrTenantRequired  model.Invalid  = "tenant_id is required"
	ErrQuantity        model.Invalid  = "quantity must be greater than zero"
	ErrLineNotPending  model.Conflict = "line is not pending deduction"
	ErrNotProfessional model.Invalid  = "only the professional of the visit can complete it"
)
