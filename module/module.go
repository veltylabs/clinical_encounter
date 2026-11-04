package module

import "webtyp.com/appkit"

// Module conecta clinical_encounter en App.Modules de una aplicación.
// Ejemplo: Modules: []appkit.Module{clinical.Module}
var Module = appkit.Module{Name: "clinical_encounter", Server: server, Browser: browser}
