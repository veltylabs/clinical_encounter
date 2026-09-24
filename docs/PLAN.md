---
PLAN: "feat(ui): el Historial Clínico con la agenda del día del doctor, su semilla y su demo viven en el módulo"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 6100775699553555915
PR: https://github.com/veltylabs/clinical_encounter/pull/5
---

> Este plan se despacha con el flujo CodeJob. Ver skill: agents-workflow.
> Todo lo que necesitas está en este plan y en `_temp/` (ver §2, nota final).

# Plan — `clinical_encounter` trae su vista (`ui/`), sus datos de demo (`seed/`) y su demo (`web/`)

## 1. Por qué

La vista de este módulo vivía en `veltylabs/mjosefa-cms/modules/clinical_encounter` (la app de
producción) y otra versión distinta en `webtyp/app-demo` (ya archivado). Cada pantalla
se escribía dos veces y probar una en la app obligaba a iniciar sesión. Decisión del
dueño (2026-09-24): **la vista vive en el módulo que posee los datos**, en un
subpaquete `ui/` del mismo repo, y el módulo trae su propia demo ejecutable en
`web/client.go`. Las apps solo inyectan base de datos y montan `ui.Browser`.

Reglas cerradas que este plan aplica (no se re-litigan):

- **Sin repos nuevos.** `ui/` es un paquete del mismo `go.mod`. El paquete raíz de
  dominio **sigue sin importar** `layout`/`components`: así el binario de un servidor
  (que importa el dominio) nunca enlaza la UI.
- **La fuente es la versión de mjosefa-cms**, no la de app-demo.
- **Grafo de dependencias entre módulos (DAG)**; un `ui/` importa el dominio y el `ui/`
  de sus módulos **upstream**, nunca de uno downstream:
  `device_manager, item_catalog, patient_directory, business_calendar` (hojas) ←
  `staff_manager` (→ device_manager) ← `appointment_booking` (→ staff_manager,
  item_catalog, patient_directory, business_calendar) ← `clinical_encounter`
  (→ appointment_booking, patient_directory, staff_manager).
- **Una pantalla que mezcla módulos vive en el módulo más downstream que toca.**
- **La demo corre entera en el navegador**: dominio real sobre `orm.New(mem.New())`
  (sin DDL: `storage/mem` no tiene esquema) + `router/loopback` + datos semilla. Sin
  servidor, sin login. `web/client.go` en la raíz de una librería es la convención que
  el daemon ya ejecuta (`layout/web/client.go`, `components/*/web/client.go`): se abre
  `webtyp` en la raíz del módulo y listo.

## 2. Contrato común (idéntico en los 7 módulos)

### 2.1 Estructura que queda en el repo

```
<modulo>/
├── (paquete de dominio, sin cambios salvo lo que diga §4)
├── ui/                 package ui — la vista (neutro: sin build tag)
│   ├── module.go       const ID, Label (identidad RBAC + ítem de nav)
│   ├── browser.go      func Browser(...) (platformd.UIModule, error) + helpers
│   ├── <otros>.go      componentes de la vista (nombres de §4)
│   ├── css.go          //go:build !wasm (solo si §4 lo trae)
│   └── svg.go          //go:build !wasm — ícono del ítem de nav
├── seed/               package seed — datos de demo (neutro: sin build tag)
│   └── seed.go         type Data struct{…}; func Load(...) (Data, error)
├── web/                package main — la demo
│   └── client.go       //go:build wasm
└── tests/              tests de la vista movidos desde mjosefa-cms (§4)
```

### 2.2 Reglas de código (obligatorias)

- **Build tags asimétricos, a propósito**: `ui/*.go` y `seed/*.go` **sin** tag (si
  `browser.go` llevara `wasm`, moriría el carril stdlib de los tests de vista);
  `ui/css.go` y `ui/svg.go` con `//go:build !wasm` (nunca llegan al binario wasm; el
  extractor SSR los descubre por import); `web/client.go` con `//go:build wasm`.
- `css.go` y `svg.go` se llaman **exactamente así**: otro nombre es invisible para el
  extractor SSR.
- Sin stdlib en código que compila a wasm (`ui/`, `seed/`, `web/`, dominio): usar
  `webtyp.com/fmt` en vez de `fmt`/`strings`/`strconv`/`errors`, `webtyp.com/time` en
  vez de `time`, `webtyp.com/json` en vez de `encoding/json`.
- **Sin `map[K]V`** en ningún archivo, tests incluidos: slice + búsqueda lineal o
  `fmt.KeyValue`.
- `dom.Element` se embebe **por valor**, nunca como puntero.
- Sin strings repetidos: los ids de tenant, ids de pestañas y etiquetas son
  constantes.
- **Datos semilla**: `seed.Load` escribe **por los métodos del módulo** (no con
  `db.Create` directo), así cada fila pasa la validación del modelo. Si una fila de §4
  no valida, `Load` devuelve ese error: **ajustar el valor semilla, nunca la
  validación del modelo**. Un `seed.Load` que falla hace `panic` en `web/client.go`.
- **Nunca** un `replace` en `go.mod`. Las dependencias nuevas se agregan con su
  **último tag publicado** (`go get <pkg>@latest`). Riesgo conocido: mjosefa-cms
  compila contra copias locales de `webtyp/components`, `webtyp/layout` y
  `webtyp/widget` (tiene `replace`). Si al portar falta un símbolo en el último tag
  publicado, **parar** y decirlo en la descripción del PR (qué símbolo, de qué paquete). Nunca
  copiar el símbolo aquí ni agregar un `replace`.
- Tests con `gotest` (nunca `go test`). `gotest` corre las dos suites: nativa y
  navegador (wasm).
- **Idioma**: si la sección "Domain-specific notes" del `AGENTS.md` de este repo fija
  uno, manda ese (`staff_manager` y `appointment_booking` exigen español en docs y
  comentarios Go). Si no fija ninguno, se usa el idioma del archivo que se edita (los
  `README.md`/`AGENTS.md` de estos módulos están en inglés). Los snippets de este plan
  traen comentarios en inglés: traducirlos cuando el repo exige español.
- **Anti-footgun**: la lista blanca de `AGENTS.md` sigue prohibiendo `layout`,
  `components`, drivers y transportes en el **paquete raíz de dominio**. No mover
  código de la vista al paquete raíz "para simplificar".

### 2.3 Cambios en `AGENTS.md` (primer paso de la etapa 1)

Aplicar al `AGENTS.md` de este repo **estos tres cambios, textuales** (en inglés, como
el resto de ese archivo por encima de la línea; la sección "Domain-specific notes" no
se toca). Es el mismo texto que ya tiene la plantilla canónica de todos los módulos:

1. En "Model definitions", la frase final `Cross-module references stay soft (a plain
   string id/SKU) — modules never import each other.` →
   `Cross-module references stay soft (a plain string id/SKU) — the root domain
   package never imports another module (see "ui/, seed/ and web/" for the DAG rule).`
2. En "Identity, persistence, …", la frase final de **Cross-module wiring**
   `The composition root wires concrete instances together. Modules never import each
   other.` → `The composition root wires concrete instances together. Root domain
   packages never import each other; only ui/, seed/ and web/ may, upstream only.`
3. Nueva sección, justo antes de "## Testing": el texto citado abajo sin los `>`, con
   su primera línea en negrita convertida en título `## ui/, seed/ and web/ — the
   module's own view and demo`.

> **`ui/`, `seed/` and `web/` — the module's own view and demo.** The whitelist and
> blacklist above apply to the **root domain package**. Three sub-packages are
> exempt, and only them:
>
> - `ui/` (package `ui`, no build tag; `css.go`/`svg.go` tagged `!wasm`) may import
>   `webtyp.com/layout/*`, `webtyp.com/components/*`, `dom`, `html`, `css`, `svg`,
>   `widget`. It is the module's screen: `const ID`, `const Label` and
>   `Browser(caller router.Caller, ids model.IDGenerator, tenantID string)
>   (platformd.UIModule, error)`.
> - `seed/` (package `seed`, no build tag) holds demo data: `Load(...) (Data, error)`,
>   which writes through the module's own methods so every row is validated, and
>   returns the rows it created so downstream demos can reference them.
> - `web/` (package `main`, `web/client.go` tagged `wasm`) is the runnable demo and
>   may additionally import the concrete `webtyp.com/storage/mem`,
>   `webtyp.com/router/loopback`, `webtyp.com/events/mock`, `webtyp.com/unixid` and
>   `webtyp.com/auth/trusted_ip` (for `ValidateRUT`).
>
> **Dependencies between modules form a DAG.** The root domain package still imports
> no sibling module. `ui/`, `seed/` and `web/` may import the domain, `ui/` and
> `seed/` packages of **upstream** modules only. Current graph: `device_manager`,
> `item_catalog`, `patient_directory`, `business_calendar` (leaves) ←
> `staff_manager` ← `appointment_booking` ← `clinical_encounter`
> (`appointment_booking` also depends on `item_catalog`, `patient_directory`,
> `business_calendar`; `clinical_encounter` also on `patient_directory` and
> `staff_manager`). A screen that mixes modules lives in the most-downstream module
> it touches.

### 2.4 La demo `web/client.go` — esqueleto común

Todas las demos tienen esta forma. §4 dice qué módulos construye y qué `seed.Load`
llama:

```go
//go:build wasm

package main

import (
	. "webtyp.com/dom"
	"webtyp.com/events/mock"
	"webtyp.com/layout/platformd"
	"webtyp.com/orm"
	"webtyp.com/router/loopback"
	"webtyp.com/storage/mem"
	"webtyp.com/unixid"
	// + los paquetes de dominio, seed y ui que diga §4
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

	// 1. construir los módulos (§4) — panic en error: una demo que no arma no sirve
	// 2. seed.Load de cada uno, upstream primero (§4) — panic en error;
	//    un seed downstream recibe el seed.Data de sus upstream
	// 3. caller := loopback.WithTenant(demoTenantID, <módulos>...)
	// 4. v, err := ui.Browser(caller, ids, demoTenantID) — panic en error

	p := &platformd.Platform{
		AppName:   "<Label del módulo> — demo",
		User:      demoUser{},
		Modules:   []platformd.UIModule{v},
		DefaultID: ui.ID,
	}
	Append("body", p)
	select {}
}
```

`loopback.WithTenant(tenantID string, mods ...router.OperationModule) router.Caller`
monta los módulos en proceso; `Brand` de `platformd` es opcional y se omite. Las
funciones que bloquean esperando una llamada (`blockingCall` en clinical_encounter)
son válidas aquí por la misma razón que en la app: `ui.Browser` corre en `main`, antes
de que arranque el event loop del navegador.

### 2.5 Design gate (común)

1. **Prior art.** Odoo: cada módulo trae modelos + **vistas** + **datos de demo**
   (`'demo': [...]` en el manifiesto, cargados solo en modo demo) y declara `depends`.
   Django: cada app trae `views`/`templates` junto a sus modelos. Storybook: cada
   componente trae su propia historia ejecutable. Aquí: `ui/` = vistas, `seed/` = datos
   de demo, `web/client.go` = la historia ejecutable, con la convención de entrada que
   el daemon ya tiene.
2. **Prueba del nombre.** `ui.Browser(caller, ids, tenantID)` → "la vista de este módulo
   para el navegador" (el mismo nombre que ya usa `mjosefa-cms/modules/browser.go`).
   `seed.Load(...)` → "cargar los datos semilla"; devuelve `seed.Data` (las filas
   creadas). `ui.ID`, `ui.Label` sin cambios.
3. **Contabilidad de complejidad.** Conceptos +1 (`seed`). Archivos que una app toca
   para mostrar este módulo: de ~4 (`module.go`, `browser.go`, `svg.go`, helpers) a 1
   línea en su registro. Formas de construir esta pantalla: de 2 (mjosefa + app-demo) a
   1. Costo honesto: el `go.mod` del módulo gana `layout`/`components`, que un consumidor
   solo-servidor descarga pero nunca enlaza.
4. **Dónde vive.** En el módulo que posee los datos (D1 del plan maestro).
5. **Qué borra.** `mjosefa-cms/modules/clinical_encounter/` completo (lo borra el plan de
   mjosefa-cms, fase C, después del tag de este repo) y los tests de vista de
   mjosefa-cms que se mueven aquí.

> **La fuente está en `_temp/` de este mismo repo.** `mjosefa-cms` es un repo privado
> al que no tienes acceso, así que el dueño copió aquí **solo** los archivos que este
> plan porta: `_temp/mjosefa-cms/modules/...` (el código de la vista) y
> `_temp/mjosefa-cms/tests/...` (los tests a mover y sus helpers `idgen_test.go` →
> `testIDGen`, `item_catalog_view_mock_test.go` → `mockCaller`). Go ignora los
> directorios que empiezan con `_`, así que `_temp/` no compila ni rompe `go build
> ./...`. Esos archivos importan `github.com/veltylabs/mjosefa-cms/...`: esos imports
> son justo lo que este plan reescribe al copiarlos fuera de `_temp/`.
> **La última etapa borra `_temp/` completo**: el PR no debe contenerlo.

## 3. Fuente

Espera el tag de `appointment_booking` (B3).

| Origen | Destino | Cambios |
|---|---|---|
| `modules/clinical_encounter/module.go` | `ui/module.go` | `package ui`. `ID = "clinical_encounter"`, `Label = "Historial Clínico"`. |
| `modules/clinical_encounter/browser.go` | `ui/browser.go` | `package ui`. Conserva su firma **con `userID`**: `Browser(caller, ids, tenantID, userID string)` (mapea la cuenta con sesión a un funcionario para armar la agenda de hoy). Conserva `blockingCall`, `requirePatient`, `loadDoctorIdentity`, `loadTodayAgenda`. |
| `modules/clinical_encounter/svg.go` | `ui/svg.go` | `package ui`, conserva `//go:build !wasm`. |
| `modules/clinical_encounter/README.md` | `docs/UI.md` | explica el picker y por qué no puede ser síncrono; rutas internas `modules/clinical_encounter/...` → `ui/...`. |

Todos bajo `_temp/mjosefa-cms/`. `server.go` no se copia. Los imports de dominio
(`appointmentbooking`, `patientdirectory`, `staffmanager`) se quedan: los permite la
regla del DAG en `ui/`.

**Fuera de alcance, no tocar:** `patients.go` (`ListRecentPatients`) todavía dice que
no existe un directorio de pacientes. Reemplazarlo por un puerto `DirectoryReader` es
un cambio de dominio con su propio plan posterior.


## 4. Etapas

### Etapa 1 — `AGENTS.md`

Aplicar §2.3.

### Etapa 2 — `ui/`

Copiar lo de §3.

### Etapa 3 — `seed/seed.go`

```go
type Upstream struct {
	Patients patientseed.Data
	Staff    staffseed.Data
}

type Data struct {
	Visits []*clinicalencounter.MedicalHistory
}

func Load(m *clinicalencounter.Module, up Upstream) (Data, error)
```

Con `m.CreateVisit(CreateVisitArgs{...})`: una ficha anterior (fecha de hace 30 días)
para cada uno de los dos primeros pacientes de `up.Patients`, atendida por el primer
funcionario de `up.Staff`, con motivo `Control` y diagnóstico `Sin hallazgos`. Los
campos exactos salen de `CreateVisitArgs` (`visit.go`); los que el modelo exige y esta
tabla no nombra, llenarlos igual que `TestWASM_ClinicalEncounter_NewDraftIsSeededFromThePicker`.

### Etapa 4 — `web/client.go`

Esqueleto de §2.4. Construir los mismos seis módulos y cargar las mismas semillas que
la demo **ya publicada** de `appointment_booking` — copiar su construcción de
`github.com/veltylabs/appointment_booking/web/client.go` (tag v0.1.14 o posterior;
está en el caché de módulos tras `go mod download github.com/veltylabs/appointment_booking`,
o en GitHub). Eso incluye `bookingseed.Load` (`github.com/veltylabs/appointment_booking/seed`),
que ya crea configuraciones de servicio y dos reservas confirmadas en el próximo día
hábil. Luego `clinicalencounter.New(db, clinicalencounter.Deps{IDs: ids, Publisher:
broker})` y `seed.Load(ce, ...)`.

El `userID` que recibe `ui.Browser` es el `UserId` del primer funcionario sembrado
(`demo-user-ana`). Para que su agenda de **hoy** no salga vacía, agregar en
`web/client.go` (no en `seed/`) una reserva confirmada **para hoy a las 16:00 hora de
Santiago** con el mismo camino que usa `bookingseed.Load`
(`CreateReservation` con `ab.LocalIntToUnixUTC(díaDeHoy, 960, "America/Santiago")`,
`Origin: ab.OriginCounter`, luego `ChangeReservationStatus` con `ab.EventConfirm`).
**Si hoy es sábado, domingo o feriado, esa reserva falla: en ese caso no hacer
`panic`, solo omitirla** (la agenda de hoy queda vacía, lo que es correcto).

### Etapa 5 — tests movidos desde mjosefa-cms

Mover `_temp/mjosefa-cms/tests/clinical_encounter_agenda_wasm_test.go` completo (`//go:build wasm`,
tres tests) a `tests/ui_agenda_wasm_test.go`. Import
`github.com/veltylabs/mjosefa-cms/modules/clinical_encounter` →
`github.com/veltylabs/clinical_encounter/ui`, selector → `ui.`. Aserciones sin
cambios. Agregar `tests/seed_test.go`: 2 fichas sin error.

### Etapa 6 — `go.mod` y README

1. `go get webtyp.com/layout@latest webtyp.com/components@latest webtyp.com/dom@latest
   webtyp.com/html@latest webtyp.com/svg@latest webtyp.com/css@latest
   webtyp.com/widget@latest webtyp.com/storage@latest webtyp.com/events@latest
   webtyp.com/unixid@latest webtyp.com/auth@latest webtyp.com/time@latest github.com/veltylabs/appointment_booking@latest github.com/veltylabs/staff_manager@latest github.com/veltylabs/device_manager@latest github.com/veltylabs/item_catalog@latest github.com/veltylabs/patient_directory@latest github.com/veltylabs/business_calendar@latest` y `go mod tidy`. Sin `replace` (ver §2.2).
2. `README.md`, sección nueva **"View and demo"** (en inglés, como el README): qué
   exporta `ui` (`ID`, `Label`, `Browser`), qué carga `seed.Load`, y "run
   `webtyp` at the repository root to open the demo — in-browser, in-memory, no login".

### Etapa 7 — borrar `_temp/`

Cuando todo lo anterior está verde: `rm -rf _temp` y commitear ese borrado en el mismo
PR. El PR no debe contener ningún archivo bajo `_temp/`.

## 5. Verificación

```bash
gotest                                                        # verde (nativo + navegador)
GOOS=js GOARCH=wasm go build -o /dev/null ./web/              # la demo compila
go list -deps . | grep -c 'webtyp.com/layout\|webtyp.com/components'   # 0: el dominio no arrastra UI
grep -rln 'webtyp.com/layout\|webtyp.com/components' --include=*.go . \
  | grep -v '^./ui/\|^./web/\|^./tests/'                      # vacío
grep -rn 'mjosefa-cms' --include=*.go .                        # vacío
grep -rn 'map\[' --include=*.go ui seed web                    # vacío
```

```bash
test ! -e _temp && echo ok                                     # _temp/ ya no existe
```

Prueba visual (la hace el dueño al revisar el PR): `webtyp` en la raíz del repo → el
navegador muestra la pantalla de §4 con los datos semilla, sin login.

## 6. Tabla de etapas

| # | Etapa | Archivos |
|---|---|---|
| 1 | AGENTS.md | `AGENTS.md` |
| 2 | vista | `ui/module.go`, `ui/browser.go`, `ui/svg.go`, `docs/UI.md` |
| 3 | semilla | `seed/seed.go` |
| 4 | demo | `web/client.go` |
| 5 | tests | `tests/ui_agenda_wasm_test.go`, `tests/seed_test.go` |
| 6 | deps + README | `go.mod`, `go.sum`, `README.md` |
| 7 | borrar `_temp/` | `_temp/` |
