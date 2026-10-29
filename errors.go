package clinical_encounter
import "webtyp.com/model"
const (
	ErrTenantRequired model.Invalid  = "tenant_id is required"
	ErrQuantity       model.Invalid  = "quantity must be greater than zero"
	ErrSupplyNotFound model.NotFound = "supply use line not found"
	ErrVisitClosed    model.Conflict = "cannot modify supply use of a closed visit"
	ErrSupplyUseOff   model.Invalid  = "supply use is not enabled"
	ErrLineDeducted   model.Conflict = "cannot modify a deducted line"
	ErrLineNotPending model.Conflict = "line is not pending deduction"
)
