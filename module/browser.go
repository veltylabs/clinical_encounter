//go:build browser

package module

import (
	"webtyp.com/appkit"
	"webtyp.com/view"

	"github.com/veltylabs/clinical_encounter/ui"
)

var server appkit.ModuleServer

var browser = view.ModuleBrowser{
	Screens: func(d view.BrowserDeps) []view.Screen {
		return ui.Screens(d)
	},
}
