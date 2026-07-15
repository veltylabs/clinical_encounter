package clinical_encounter

import (
	"github.com/tinywasm/dom"
	"github.com/tinywasm/html"
)

// PatientHeadView displays patient name and RUT in the header slot.
type PatientHeadView struct {
	dom.Element // value, never nil
	Name        string
	RUT         string
}

func (p *PatientHeadView) Render() *dom.Element {
	return html.Span().
		Class("ce-patient-head").
		Text(p.Name + " — " + p.RUT)
}

// VisitFormView is a placeholder stub.
// In production, use tinywasm/form.New(parentID, CreateVisitArgs{})
// to build the form dynamically.
type VisitFormView struct {
	dom.Element // value, never nil
	Record      *MedicalHistory
	DoctorName  string
	onSave      func(args CreateVisitArgs)
	onCancel    func()
}

func (v *VisitFormView) Render() *dom.Element {
	return html.Div().Class("ce-form-container").
		Text("Formulario renderizado por tinywasm/form")
}

// HistoryListView displays a list of medical history cards in the aside slot.
type HistoryListView struct {
	dom.Element // value, never nil
	Items       []*MedicalHistory
	onSelect    func(id string)
}

func (h *HistoryListView) Render() *dom.Element {
	list := html.Ul().Class("ce-history-list")

	for _, item := range h.Items {
		list.Add(h.renderCard(item))
	}

	return list
}

func (h *HistoryListView) renderCard(item *MedicalHistory) *dom.Element {
	dateStr := formatDate(item.AttentionAt)
	statusClass := "ce-status-" + item.Status
	reasonText := item.Reason
	if len(reasonText) > 40 {
		reasonText = reasonText[:37] + "..."
	}

	return html.Li().Class("ce-history-card").ID(item.ID).
		Add(
			html.Div().Class("ce-card-date").Text(dateStr),
			html.Div().Class("ce-card-status").Class(statusClass).Text(item.Status),
			html.Div().Class("ce-card-reason").Text(reasonText),
		)
}

// HistorySearchView displays a search input in the aside-controls slot.
type HistorySearchView struct {
	dom.Element // value, never nil
}

func (h *HistorySearchView) Render() *dom.Element {
	return html.Div().Class("ce-search-container").
		Add(html.P().Text("🔍 Buscar..."))
}

// formatDate converts Unix timestamp to a display string.
func formatDate(unix int64) string {
	if unix == 0 {
		return "—"
	}
	return "DD/MM/YYYY"
}
