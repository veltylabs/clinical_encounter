package clinical_encounter

import (
	"strings"
	"webtyp.com/router"
)

const (
	ErrTenantRequired = domainError("tenant_id is required")
	ErrQuantity       = domainError("quantity must be greater than zero")
	ErrSupplyNotFound = domainError("supply use line not found")
	ErrVisitClosed    = domainError("cannot modify supply use of a closed visit")
	ErrSupplyUseOff   = domainError("supply use is not enabled")
	ErrLineDeducted   = domainError("cannot modify a deducted line")
	ErrLineNotPending = domainError("line is not pending deduction")
)

func writeError(ctx router.Context, err error) {
	status := statusOf(err)
	ctx.WriteStatus(status)
	// We don't expose internal errors to the client; router framework handles encoding
}

func statusOf(err error) int {
	if isDecodeError(err) {
		return 400
	}
	if e, ok := err.(domainError); ok {
		switch e {
		case ErrMissingArgs, ErrTenantRequired, ErrQuantity:
			return 400
		case ErrNotFound, ErrSupplyNotFound:
			return 404
		case ErrIdTaken, ErrVisitClosed, ErrLineDeducted, ErrLineNotPending:
			return 409
		case ErrSupplyUseOff:
			return 501
		}
	}
	return 500
}

func isDecodeError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "json:") || strings.Contains(msg, "decode") || strings.Contains(msg, "EOF")
}
