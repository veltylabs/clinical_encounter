//go:build !browser

package module

import (
	"errors"

	"webtyp.com/appkit"
	"webtyp.com/model"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
)

var server = appkit.ModuleServer{Build: build}
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

	// El uso de insumos es opcional y va completo o no va: sin uno de los tres
	// módulos (catálogo, salas, insumos) no se conecta ninguno.
	recipes, errRecipes := optional[clinicalencounter.RecipeReader](d)
	rooms, errRooms := optional[clinicalencounter.RoomFinder](d)
	supplies, errSupplies := optional[clinicalencounter.SupplyStock](d)
	for _, err := range []error{errRecipes, errRooms, errSupplies} {
		if err != nil {
			return nil, err
		}
	}
	if recipes == nil || rooms == nil || supplies == nil {
		recipes, rooms, supplies = nil, nil, nil
	}

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

// optional busca un contrato que la app puede no tener: sin proveedor devuelve nil
// sin error; con más de uno, el error de appkit.
func optional[T any](d appkit.Deps) (T, error) {
	v, err := appkit.Find[T](d)
	if errors.Is(err, appkit.ErrNoProvider) {
		return v, nil
	}
	return v, err
}
