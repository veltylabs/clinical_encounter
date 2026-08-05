package tests

import (
	"testing"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"github.com/tinywasm/model"
	"github.com/tinywasm/view"
	"github.com/tinywasm/view/conformance"
)

func TestView_ListPopulatesItems(t *testing.T) {
	caller := &conformance.FakeCaller{
		Reply: func(op string, into model.Decodable) {
			if op != clinicalencounter.OpListVisitsByPatient {
				return
			}
			list := into.(*clinicalencounter.MedicalHistoryList)
			rec := list.Append().(*clinicalencounter.MedicalHistory)
			rec.Id, rec.Reason, rec.Status = "mh_1", "Control", clinicalencounter.StatusCompleted
		},
	}
	p := clinicalencounter.NewView(caller)
	if err := p.Reload(); err != nil {
		t.Fatalf("Reload: %v", err)
	}
	items := p.Items()
	if len(items) != 1 || items[0].ID != "mh_1" || items[0].Label != "Control" {
		t.Fatalf("unexpected items: %+v", items)
	}
	if _, ok := p.(view.Saver); !ok {
		t.Error("expected Saver capability (WithSaveOp is configured)")
	}
	if _, ok := p.(view.Deleter); ok {
		t.Error("expected no Deleter capability (no WithDeleteOp by design)")
	}
}
