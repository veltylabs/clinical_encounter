package tests

import (
	"testing"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"github.com/tinywasm/form"
	"github.com/tinywasm/model"
)

var _ model.Model = (*clinicalencounter.CreateVisitArgs)(nil)

// TestCreateVisitArgsHasWidgets guards against CreateVisitArgsModel silently losing its
// widgets again — form.New must never render an empty form for it.
//
// API note (`tinywasm/form@v0.3.13`, verified against the real published module — this plan
// originally targeted the now-stale `v0.2.16`, where `form.New` returned a single `*Form`):
// `form.New` now returns `(*Form, error)`.
func TestCreateVisitArgsHasWidgets(t *testing.T) {
	f, err := form.New("clinical_encounter", &clinicalencounter.CreateVisitArgs{})
	if err != nil {
		t.Fatalf("form.New: %v", err)
	}
	if len(f.Inputs) == 0 {
		t.Fatal("CreateVisitArgsModel has no widgets — form.New returned an empty form")
	}
}
