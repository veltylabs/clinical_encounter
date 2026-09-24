//go:build wasm

package main

import (
	. "webtyp.com/dom"
	"webtyp.com/events/mock"
	"webtyp.com/layout/platformd"
	"webtyp.com/orm"
	"webtyp.com/router/loopback"
	"webtyp.com/storage/mem"
	tinytime "webtyp.com/time"
	"webtyp.com/unixid"

	"webtyp.com/auth/trusted_ip"

	ab "github.com/veltylabs/appointment_booking"
	bookingseed "github.com/veltylabs/appointment_booking/seed"
	businesscalendar "github.com/veltylabs/business_calendar"
	calendarseed "github.com/veltylabs/business_calendar/seed"
	clinicalencounter "github.com/veltylabs/clinical_encounter"
	ceseed "github.com/veltylabs/clinical_encounter/seed"
	"github.com/veltylabs/clinical_encounter/ui"
	devicemanager "github.com/veltylabs/device_manager"
	deviceseed "github.com/veltylabs/device_manager/seed"
	itemcatalog "github.com/veltylabs/item_catalog"
	catalogseed "github.com/veltylabs/item_catalog/seed"
	patientdirectory "github.com/veltylabs/patient_directory"
	patientseed "github.com/veltylabs/patient_directory/seed"
	staffmanager "github.com/veltylabs/staff_manager"
	staffseed "github.com/veltylabs/staff_manager/seed"
)

// demoTenantID is the only tenant of this in-browser demo.
const demoTenantID = "demo"

// demoUser is the fixed identity the demo shell shows: the demo has no login.
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

	// 1. Build modules in dependency order
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
	ic, err := itemcatalog.New(db, itemcatalog.Deps{IDs: ids, Publisher: broker})
	if err != nil {
		panic(err)
	}
	pd, err := patientdirectory.New(db, patientdirectory.Deps{
		IDs: ids, Publisher: broker, TenantID: demoTenantID, ValidateRUT: trustedip.ValidateRUT,
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
	ceMod, err := clinicalencounter.New(db, clinicalencounter.Deps{IDs: ids, Publisher: broker})
	if err != nil {
		panic(err)
	}

	// 2. Load seeds, upstream first
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
	if _, err := ceseed.Load(ceMod, ceseed.Upstream{
		Patients: patientData,
		Staff:    staffData,
	}); err != nil {
		panic(err)
	}

	// 3. Add a confirmed reservation for today at 16:00 Santiago time if possible (omit if weekend/holiday)
	nowSec := tinytime.Now() / 1e9
	isoDate := tinytime.FormatISO8601(nowSec * 1e9)[:10]
	if dayNano, err := tinytime.ParseDate(isoDate); err == nil {
		todaySec := dayNano / 1e9
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

	// 4. Mount modules in loopback router
	caller := loopback.WithTenant(demoTenantID, ceMod, abMod, sm, dm, ic, pd, bc)

	// 5. Build WASM browser view for doctor 'demo-user-ana'
	v, err := ui.Browser(caller, ids, demoTenantID, "demo-user-ana")
	if err != nil {
		panic(err)
	}

	p := &platformd.Platform{
		AppName:   ui.Label + " — demo",
		User:      demoUser{},
		Modules:   []platformd.UIModule{v},
		DefaultID: ui.ID,
	}
	Append("body", p)
	select {}
}
