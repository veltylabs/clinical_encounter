//go:build browser

package module

import (
	"webtyp.com/appkit"
	"webtyp.com/view"

	"github.com/veltylabs/clinical_encounter/ui"
)

var server appkit.ModuleServer

var browser = appkit.ModuleBrowser{
	Screens: func(d appkit.BrowserDeps) []view.Screen {
		return ui.Screens(d.Caller, d.Zone, nil, "", "")
	},
}
