//go:build !browser

package module

import (
	"webtyp.com/appkit"
	"webtyp.com/model"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"github.com/veltylabs/clinical_encounter/migrate"
)

var server = appkit.ModuleServer{Migrate: migrate.Migrate, Build: build}
var browser appkit.ModuleBrowser

func build(d appkit.Deps) (model.ModuleNaming, error) {
	doc, err := appkit.Find[clinicalencounter.DoctorReader](d)
	if err != nil {
		return nil, err
	}
	ag, err := appkit.Find[clinicalencounter.AgendaReader](d)
	if err != nil {
		return nil, err
	}
	pat, err := appkit.Find[clinicalencounter.PatientReader](d)
	if err != nil {
		return nil, err
	}

	recipes, _ := appkit.Find[clinicalencounter.RecipeReader](d)
	rooms, _ := appkit.Find[clinicalencounter.RoomFinder](d)
	supplies, _ := appkit.Find[clinicalencounter.SupplyStock](d)

	return clinicalencounter.New(d.DB, clinicalencounter.Deps{
		IDs:       d.IDs,
		Publisher: d.Events,
		TenantID:  d.TenantID,
		Zone:      d.Zone,
		Doctors:   doc,
		Agenda:    ag,
		Patients:  pat,
		Recipes:   recipes,
		Rooms:     rooms,
		Supplies:  supplies,
	})
}
