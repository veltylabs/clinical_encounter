//go:build wasm

package clinical_encounter

import (
	"github.com/tinywasm/dom"
)

// ViewCtrl manages interactions between views and business logic.
type ViewCtrl struct {
	historyList *HistoryListView
	visitForm   *VisitFormView
	allItems    []*MedicalHistory
}

// NewViewCtrl initializes the view controller with initial data.
func NewViewCtrl(items []*MedicalHistory) *ViewCtrl {
	return &ViewCtrl{
		historyList: &HistoryListView{Items: items},
		visitForm:   &VisitFormView{Record: nil},
		allItems:    items,
	}
}

// SetupEventHandlers registers click handlers for history cards.
// Must be called after rendering to wire up WASM event listeners.
func (vc *ViewCtrl) SetupEventHandlers() {
	for _, item := range vc.allItems {
		vc.registerCardClick(item.ID)
	}
}

// registerCardClick registers a click handler for a single card.
func (vc *ViewCtrl) registerCardClick(cardID string) {
	ref, ok := dom.Get(cardID)
	if !ok {
		return
	}
	ref.On("click", func(e dom.Event) {
		vc.onCardClick(cardID)
	})
}

// onCardClick handles a history card click.
func (vc *ViewCtrl) onCardClick(cardID string) {
	record := vc.findByID(cardID)
	if record == nil {
		return
	}
	vc.visitForm.Record = record
	dom.Update(vc.visitForm)
}

// findByID finds a medical history record by ID.
func (vc *ViewCtrl) findByID(id string) *MedicalHistory {
	for _, item := range vc.allItems {
		if item.ID == id {
			return item
		}
	}
	return nil
}

// GetHistoryListView returns the history list view.
func (vc *ViewCtrl) GetHistoryListView() *HistoryListView {
	return vc.historyList
}

// GetVisitFormView returns the visit form view.
func (vc *ViewCtrl) GetVisitFormView() *VisitFormView {
	return vc.visitForm
}
