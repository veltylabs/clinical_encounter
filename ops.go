package clinical_encounter

import (
	"github.com/tinywasm/model"
	"github.com/tinywasm/router"
)

const (
	OpCreateVisit          = "create_visit"
	OpGetVisit             = "get_medical_history"
	OpListVisitsByPatient  = "list_medical_history"
)

func (m *Module) MountOps(reg router.OpRegistry) {
	reg.Op(OpCreateVisit, m.opCreateVisit).Requires("medical_history", model.Create).Accepts(&CreateVisitArgs{})
	reg.Op(OpGetVisit, m.opGetVisit).Requires("medical_history", model.Read).Accepts(&GetVisitArgs{})
	reg.Op(OpListVisitsByPatient, m.opListVisits).Requires("medical_history", model.Read).Accepts(&ListVisitsArgs{})
}

var _ router.OpModule = (*Module)(nil)

func (m *Module) opCreateVisit(ctx router.Context) {
	var args CreateVisitArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	record, err := m.CreateVisit(args)
	if err != nil {
		// Status convention (ecosystem-wide): 400 = invalid input, 404 = not found,
		// 500 = genuine internal error only — never collapse client errors into 500.
		if err == ErrMissingArgs {
			ctx.WriteStatus(400)
			return
		}
		ctx.WriteStatus(500)
		return
	}
	if err := ctx.Encode(record); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opGetVisit(ctx router.Context) {
	var args GetVisitArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	record, err := m.GetVisit(args.Id)
	if err != nil {
		if err == ErrNotFound {
			ctx.WriteStatus(404)
			return
		}
		ctx.WriteStatus(500)
		return
	}
	if err := ctx.Encode(record); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opListVisits(ctx router.Context) {
	var args ListVisitsArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	records, err := m.ListVisitsByPatient(args.PatientId)
	if err != nil {
		ctx.WriteStatus(500)
		return
	}
	list := make(MedicalHistoryList, len(records))
	for i, r := range records {
		list[i] = r
	}
	if err := ctx.Encode(&list); err != nil {
		ctx.WriteStatus(500)
	}
}
