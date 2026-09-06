package tests

import (
	"testing"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

type fakeCaller struct {
	reply func(op string, into model.Decodable)
}

func (f *fakeCaller) Call(op string, args model.Encodable, into model.Decodable, done func(err error)) {
	if f.reply != nil {
		f.reply(op, into)
	}
	if done != nil {
		done(nil)
	}
}

func (f *fakeCaller) Dispatch(op string, args model.Encodable) {}

var _ router.Caller = (*fakeCaller)(nil)

func TestView_ListPopulatesItems(t *testing.T) {
	caller := &fakeCaller{
		reply: func(op string, into model.Decodable) {
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
