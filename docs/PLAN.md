---
PLAN: "feat: the demo shows its API and its database (layout v0.3.99 developer tools)"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 10272247989571858546
PR: https://github.com/veltylabs/clinical_encounter/pull/13
REVIEWED_COMMIT: 5936884055420faa7e63f6b9936fedee2b20c5a8
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work.

# Herramientas de desarrollador en la demo

## Contexto

La demo de este repo (`web/browser.go`) se ejecuta enteramente en el navegador: sus operaciones pasan a través de `loopback` y sus datos viven
en memoria, sin servidor. Hasta ahora el menú siempre añadía «Developer tools» (Herramientas de desarrollador), y ambas pantallas preguntaban a un servidor que
no existe por `/_routes` y `/_schema` (404). `webtyp.com/layout` v0.3.99 arregla eso: `BuildMenu` ahora toma las
herramientas como un cuarto argumento **requerido**, y `app.LocalTools(modules, steps...)` las construye en el navegador con las
mismas piezas que usa el servidor — las rutas que las operaciones declaran y los pasos de migración. Así que esta demo deja de compilar
hasta que pase sus herramientas; este plan hace exactamente eso. En una demo el explorador muestra la tabla sin el formulario de «probar»
(no hay servidor a dónde enviar una solicitud).

## Herramientas

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go get webtyp.com/layout@v0.3.99
go mod tidy
```

Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`.

## APIs utilizadas

| Símbolo | Firma exacta | Tag |
|---|---|---|
| `app.LocalTools` | `func LocalTools(modules []router.OperationModule, steps ...ddl.Step) (Tools, error)` | `webtyp.com/layout/app` v0.3.99 |
| `app.BuildMenu` | `func BuildMenu(screens []view.Screen, ids model.IDGenerator, allows func(model.Resource) bool, tools Tools) (Menu, error)` | v0.3.99 |
| `ddl.Step` | `struct { Name string; Run func(conn Execer, compiler Compiler) error }` | `webtyp.com/ddl` |
| cada `migrate.Migrate` utilizado debajo | `func Migrate(conn ddl.Execer, ddlCompiler ddl.Compiler) error` — encaja en `ddl.Step.Run` | cada módulo |

En este archivo el paquete layout se importa como `app`. Si una firma difiere, detente y nombra el símbolo en
el PR; nunca lo adaptes adivinando.

## Etapa 1 — `web/browser.go`

Agrega las importaciones (mantén las existentes; usa los alias que ya están en el archivo para los paquetes raíz):

```go
	devicemanagermigrate "github.com/veltylabs/device_manager/migrate"
	staffmanagermigrate "github.com/veltylabs/staff_manager/migrate"
	servicecatalogmigrate "github.com/veltylabs/service_catalog/migrate"
	patientdirectorymigrate "github.com/veltylabs/patient_directory/migrate"
	businesscalendarmigrate "github.com/veltylabs/business_calendar/migrate"
	abmigrate "github.com/veltylabs/appointment_booking/migrate"
	clinicalencountermigrate "github.com/veltylabs/clinical_encounter/migrate"
	"webtyp.com/ddl"
	"webtyp.com/router"
```

Justo antes de la llamada existente `app.BuildMenu(...)`, construye las herramientas a partir de los **mismos módulos que el llamador loopback
monta** y los pasos de migración de los módulos, en este orden:

```go
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
```

y pasa `tools` como el cuarto argumento de la llamada existente `BuildMenu` (nada más en esa llamada cambia). Si `err`
ya está declarado en ese ámbito, usa `=` donde el compilador lo solicite.


Comprueba que compila para el navegador: `GOOS=js GOARCH=wasm go build -o /dev/null ./web`; luego ejecuta `gotest`.
Haz commit.

## Etapa 2 — idioma

Traduce a Español cada README.md, archivo `docs/*`, diagrama y comentario de código de este repo que aún esté en Inglés
(sólo esos; nunca reescribas lo que ya esté en Español), y agrega un ejemplo de uso (`// Ejemplo: …`) a los comentarios de
símbolos exportados que no tengan uno. El código, los identificadores y las claves de traducción permanecen en Inglés. Haz commit.

## Chequeo de entrega

1. `GOOS=js GOARCH=wasm go build -o /dev/null ./web` tiene éxito y `gotest` pasa. Un PR que no compila, o cuyas pruebas no pudieron ejecutarse, es una entrega fallida.
2. `grep -n "LocalTools" web/browser.go` → una coincidencia; toda llamada a `BuildMenu(` tiene cuatro argumentos.

## Reglas de entrega

- Haz commit luego de cada etapa. Nunca ejecutes `git reset`, `git restore`, `git checkout -- <file>` o `git stash` en tu propio trabajo.
- Atascado por 30 minutos: abre el PR de todas formas, listando qué falla.
- No cambies nada más en la demo (pantallas, semillas, eventos).

## Etapas

| # | Archivos | Resultado |
|---|---|---|
| 1 | `web/browser.go`, `go.mod`, `go.sum` | La demo pasa sus herramientas; el explorador y el diagrama muestran sus propias API y tablas |
| 2 | docs y comentarios aún en Inglés | Español |
## Executor notes
- None
