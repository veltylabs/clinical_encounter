//go:build browser

package main

import (
	. "webtyp.com/dom"
	"webtyp.com/events/mock"
	"webtyp.com/layout/app"
	"webtyp.com/layout/platformd"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/router/loopback"
	"webtyp.com/storage/mem"
	tinytime "webtyp.com/time"
	"webtyp.com/unixid"

	"webtyp.com/auth/trusted_ip"

	ab "github.com/veltylabs/appointment_booking"
	bookingseed "github.com/veltylabs/appointment_booking/seed"
	businesscalendar "github.com/veltylabs/business_calendar"
	businesscalendarmigrate "github.com/veltylabs/business_calendar/migrate"
	calendarseed "github.com/veltylabs/business_calendar/seed"
	clinicalencounter "github.com/veltylabs/clinical_encounter"
	clinicalencountermigrate "github.com/veltylabs/clinical_encounter/migrate"
	ceseed "github.com/veltylabs/clinical_encounter/seed"
	"github.com/veltylabs/clinical_encounter/ui"
	devicemanager "github.com/veltylabs/device_manager"
	devicemanagermigrate "github.com/veltylabs/device_manager/migrate"
	deviceseed "github.com/veltylabs/device_manager/seed"
	patientdirectory "github.com/veltylabs/patient_directory"
	patientdirectorymigrate "github.com/veltylabs/patient_directory/migrate"
	patientseed "github.com/veltylabs/patient_directory/seed"
	servicecatalog "github.com/veltylabs/service_catalog"
	servicecatalogmigrate "github.com/veltylabs/service_catalog/migrate"
	catalogseed "github.com/veltylabs/service_catalog/seed"
	staffmanager "github.com/veltylabs/staff_manager"
	staffmanagermigrate "github.com/veltylabs/staff_manager/migrate"
	staffseed "github.com/veltylabs/staff_manager/seed"
	abmigrate "github.com/veltylabs/appointment_booking/migrate"
	"webtyp.com/ddl"
	"webtyp.com/router"
)

// demoTenantID es el único tenant de esta demo en el navegador.
const demoTenantID = "demo"

// demoUser es la identidad fija que muestra la demo: la demo no tiene inicio de sesión.
type demoUser struct{}

func (demoUser) UserName() string    { return "Demo" }
func (demoUser) UserAvatar() string  { return "" }
func (demoUser) UserRoles() []string { return []string{"Administrador"} }

func main() {
	ids, err := unixid.NewUnixID()
	if err != nil {
		panic(err)
	}
	db := orm.New(mem.New())
	broker := &mock.Broker{}

	// 1. Construir módulos en orden de dependencia
	dm, err := devicemanager.New(db, devicemanager.Deps{
		IDs: ids, Publisher: broker, TenantID: demoTenantID,
	})
	if err != nil {
		panic(err)
	}
	sm, err := staffmanager.New(db, staffmanager.Deps{
		IDs:         ids,
		Publisher:   broker,
		TenantID:    demoTenantID,
		ValidateRUT: trustedip.ValidateRUT,
		Devices:     devicemanager.IPLocator{Devices: dm, TenantID: demoTenantID},
	})
	if err != nil {
		panic(err)
	}
	ic, err := servicecatalog.New(db, servicecatalog.Deps{IDs: ids, Publisher: broker})
	if err != nil {
		panic(err)
	}
	pd, err := patientdirectory.New(db, patientdirectory.Deps{
		Publisher: broker, TenantID: demoTenantID, ValidateRUT: trustedip.ValidateRUT,
	})
	if err != nil {
		panic(err)
	}
	bc, err := businesscalendar.New(db, businesscalendar.Deps{IDs: ids, Publisher: broker})
	if err != nil {
		panic(err)
	}
	abMod, err := ab.New(db, ab.Deps{
		Staff: sm, Catalog: ic, Directory: pd, Bounds: bc, IDs: ids, Publisher: broker,
	})
	if err != nil {
		panic(err)
	}
	ceMod, err := clinicalencounter.New(db, clinicalencounter.Deps{Publisher: broker})
	if err != nil {
		panic(err)
	}

	// 2. Cargar semillas, primero los dependientes
	if _, err := deviceseed.Load(dm, demoTenantID); err != nil {
		panic(err)
	}
	staffData, err := staffseed.Load(sm, demoTenantID)
	if err != nil {
		panic(err)
	}
	catalogData, err := catalogseed.Load(ic, demoTenantID)
	if err != nil {
		panic(err)
	}
	patientData, err := patientseed.Load(pd, demoTenantID)
	if err != nil {
		panic(err)
	}
	if _, err := calendarseed.Load(bc); err != nil {
		panic(err)
	}
	bookingData, err := bookingseed.Load(abMod, demoTenantID, bookingseed.Upstream{
		Staff:    staffData,
		Catalog:  catalogData,
		Patients: patientData,
	})
	if err != nil {
		panic(err)
	}
	if _, err := ceseed.Load(ceMod, demoTenantID, ceseed.Upstream{
		Patients: patientData,
		Staff:    staffData,
	}); err != nil {
		panic(err)
	}

	// 3. Agregar una reserva confirmada para hoy a las 16:00 hora de Santiago si es posible (omitir si es fin de semana/feriado)
	nowSec := tinytime.NowUnix()
	isoDate := tinytime.FormatISO8601(nowSec)[:10]
	if daySec, err := tinytime.ParseDate(isoDate); err == nil {
		todaySec := daySec
		if len(patientData.Patients) > 0 && len(bookingData.ServiceConfigs) > 0 {
			resToday, err := abMod.CreateReservation(ab.CreateReservationCmd{
				TenantId:                demoTenantID,
				ClientId:                patientData.Patients[0].Id,
				EmployeeServiceConfigId: bookingData.ServiceConfigs[0].Id,
				SlotStartUtc:            ab.LocalIntToUnixUTC(todaySec, 960, "America/Santiago"),
				Origin:                  ab.OriginCounter,
			})
			if err == nil {
				_ = abMod.ChangeReservationStatus(ab.ChangeStatusCmd{
					TenantId: demoTenantID,
					Id:       resToday.Id,
					Event:    ab.EventConfirm,
					Revision: int(resToday.Revision),
				})
			}
		}
	}

	// 4. Montar módulos en el enrutador loopback
	caller := loopback.WithTenant(demoTenantID, ceMod, abMod, sm, dm, ic, pd, bc)

	// 5. Construir vista del navegador WASM para la doctora 'demo-user-ana'
	screens := ui.Screens(caller, ids, demoTenantID, "demo-user-ana")

	steps := []ddl.Step{
		{Name: devicemanager.ModelName, Run: devicemanagermigrate.Migrate},
		{Name: staffmanager.ModelName, Run: staffmanagermigrate.Migrate},
		{Name: servicecatalog.ModelName, Run: servicecatalogmigrate.Migrate},
		{Name: patientdirectory.ModelName, Run: patientdirectorymigrate.Migrate},
		{Name: businesscalendar.ModelName, Run: businesscalendarmigrate.Migrate},
		{Name: ab.ModelName, Run: abmigrate.Migrate},
		{Name: clinicalencounter.ModelName, Run: clinicalencountermigrate.Migrate},
	}
	tools, err := app.LocalTools([]router.OperationModule{ceMod, abMod, sm, dm, ic, pd, bc}, steps...)
	if err != nil {
		panic(err)
	}

	menu, err := app.BuildMenu(screens, ids, func(model.Resource) bool { return true }, tools)
	if err != nil {
		panic(err)
	}

	p := &platformd.Platform{
		AppName:   ui.Label + " — demo",
		User:      demoUser{},
		Modules:   menu.Modules,
		CanView:   menu.CanView,
		DefaultID: ui.ID,
	}
	Append("body", p)
	select {}
}
