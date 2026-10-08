package tests

import (
	"testing"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
)

func TestSentinelErrorTexts(t *testing.T) {
	if got := clinicalencounter.ErrNotFound.Error(); got != "visit not found" {
		t.Errorf("expected ErrNotFound to be 'visit not found', got %q", got)
	}
	if got := clinicalencounter.ErrMissingArgs.Error(); got != "missing required arguments" {
		t.Errorf("expected ErrMissingArgs to be 'missing required arguments', got %q", got)
	}
}
