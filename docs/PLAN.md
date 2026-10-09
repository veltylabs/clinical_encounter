---
PLAN: "feat!: el id de la visita lo genera quien llama — CreateVisitArgs.Id obligatorio, reintento idempotente; se elimina Deps.IDs"
TAG: v0.2.0
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — `clinical_encounter`: IDs generados por quien llama (offline-first)

## 0. Contexto (leer primero)

Una ola offline-first (master plan:
<https://github.com/veltylabs/mjosefa-cms/blob/main/docs/OFFLINE_FIRST_MASTER_PLAN.md>) hace
que un médico pueda **abrir y escribir una ficha sin servidor**. El navegador del box ejecuta
este mismo módulo sobre su base local, guarda la operación en una cola (outbox) y la reenvía al
servidor cuando vuelve la conexión. Para eso:

- **El id de la visita lo genera quien llama** (decisión OF-5 del master): la visita creada sin
  conexión llega al servidor con el mismo id, y lo que se escriba después la referencia.
- **Reenviar la misma creación no duplica ni falla**: si existe una visita con ese id, se
  devuelve la guardada.
- **Sin retrocompatibilidad** (OF-6).

Fuera de este plan (vendrá en otro): versionar la ficha, que quede cerrada al completarse y la
adenda (OF-17).

Estado actual (verificado en `main`, v0.1.79):

- `CreateVisit` (`visit.go`) asigna `Id: m.ids.NewID()` y no acepta un id del que llama.
- `CreateVisitArgsModel` (`model.go`) no tiene campo `id`.
- `webtyp/form` ya genera en el cliente el id de un registro nuevo: a todo campo de texto que
  sea **PK** lo oculta y le asigna `IDGenerator.NewID()` al enviar, y **conserva ese id entre
  reintentos** hasta `Reset` (`form/sync.go`). Basta con que `CreateVisitArgs` declare su `id`
  como PK.

## Design gate

### 1. Prior art
- **Replicache / Zero**: ids generados en el cliente; la mutación de creación es idempotente.
- **Stripe `Idempotency-Key`**: repetir una creación devuelve el mismo objeto.
- **FHIR `PUT Encounter/{id}`** (creación con id del cliente, "update as create"): el cliente
  nombra el recurso clínico.

### 2. Prueba del nombre para un novato
`CreateVisit(CreateVisitArgs{Id: id, ...})`: "crea la visita con este id". Sin nombres nuevos.

### 3. Libro de complejidad
```
Conceptos que aprender              −1 (Deps.IDs)
Archivos que tocar para hacer X      0
Líneas en el sitio de llamada        +1 (Id en los args) / −1 (IDs en Deps)
Formas de hacer lo mismo             0
```

### 4. Dónde va
En este módulo (su regla de identidad). Generar el id corresponde a quien llama: el formulario
(`crudview`, que ya recibe `ids` en `ui.Browser`), el seed y la importación de datos.

### 5. Qué borra
`Deps.IDs`, `Module.ids`, la comprobación de `deps.IDs` en `New`, `Id: m.ids.NewID()` en
`CreateVisit`.

## 1. Cambios (normativos)

### 1.1 `model.go` + regenerar `model_orm.go`
- Agregar como **primer** campo de `CreateVisitArgsModel`:
  `{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, NotNull: true}`.
  Es el PK que `webtyp/form` oculta y asigna. Actualizar el comentario de `CreateVisitArgsModel`:
  `id` lo genera quien llama.
- Nuevo error de dominio junto a los existentes:
  `ErrIdTaken domainError = "visit id already used by another patient"`.
- Regenerar con `go install webtyp.com/ormc/cmd/ormc@latest && ormc` en la raíz del repo.
  Commitear el `model_orm.go` generado; nunca editarlo a mano.

### 1.2 `module.go`
`Deps` pierde `IDs`; `Module` pierde `ids`; `New` deja de exigirlo. Actualizar el comentario.

### 1.3 `visit.go` — `CreateVisit`
En este orden:
1. `args.Id == ""` o falta alguno de los obligatorios actuales → `ErrMissingArgs` (mismo error de
   hoy; `id` pasa a ser uno más de los obligatorios).
2. **Reintento**: si existe una visita con `args.Id` (`GetVisit`):
   - del mismo `PatientId` → devolverla tal como está guardada, sin escribir ni publicar evento;
   - de otro paciente → `ErrIdTaken`.
   - `ErrNotFound` → seguir. Distinguir por aserción de tipo `err.(domainError)`, nunca con `==`
     entre interfaces (en TinyGo arrastra `reflectlite`).
3. Resto igual que hoy, con `Id: args.Id`.

### 1.4 `ops.go`
`opCreateVisit`: `ErrIdTaken` → 409. El resto no cambia.

### 1.5 `seed/seed.go`
Cada visita de demo lleva un id determinístico (`"demo-visit-1"`, `"demo-visit-2"`, como
constantes del paquete `seed`), porque el módulo ya no los genera.

### 1.6 `web/client.go`
Quitar `IDs` de `clinicalencounter.Deps`. Las demás `Deps` de módulos de arriba quedan como
están (se actualizan cuando esos módulos cambien). `ids` se sigue pasando a `ui.Browser`.

### 1.7 `ui/`
Si `ui/browser.go` arma `CreateVisitArgs` a mano (fuera del formulario), asignarle
`Id: ids.NewID()` **una sola vez por intento de creación**, guardado mientras se reintenta, igual
que hace `form`. Si todas las creaciones pasan por el formulario, no hay que tocar nada.

### 1.8 Documentación
`README.md` y `docs/ARCHITECTURE.md`: quitar `IDs` del ejemplo de raíz de composición; en "Core
Entities" decir que `CreateVisitArgs.Id` lo genera quien llama y que repetir la creación devuelve
la visita guardada. Tabla de ops: `create_visit` responde 409 con `ErrIdTaken`.

## 2. Tests (`tests/`, paquete externo; `gotest`)

Actualizar los existentes (todo `CreateVisit` lleva `Id`; `setup` sin `IDs`; borrar `mockIDGen`
si queda sin uso). Agregar:

1. Sin `Id` → `ErrMissingArgs`.
2. Reintento: dos `CreateVisit` con el mismo id y paciente → la segunda devuelve la misma visita
   sin error; hay **una** fila; **un** evento publicado (contar con un publisher de test).
3. Mismo id con otro paciente → `ErrIdTaken`; por la op → 409.
4. `seed.Load` dos veces sobre la misma base → sin error y sin duplicados.
5. Vista wasm existente (`*_view_*_test.go`): crear una ficha desde el formulario sigue
   funcionando (el id lo asigna `form`).

## 3. Reglas de código (no negociables)
- Paquete raíz compila a TinyGo WASM: `webtyp.com/fmt`; nada de `errors`, `strconv`, `strings`.
- Errores como constantes `domainError`, comparadas por aserción de tipo.
- Tests solo en `tests/`; nunca exportar un símbolo para un test.

## 4. Criterios de aceptación
- `gotest ./...` en verde (stdlib y wasm).
- `grep -rn "m.ids\|NewID" visit.go module.go` → vacío.
- `grep -n "\"id\"" model.go` muestra el campo en `CreateVisitArgsModel`.

| Etapa | Archivos | Listo cuando |
|---|---|---|
| 1 | `model.go`, `model_orm.go` | `CreateVisitArgs.Id` existe (PK) |
| 2 | `module.go`, `visit.go`, `ops.go` | id obligatorio, reintento idempotente, sin `Deps.IDs` |
| 3 | `seed/seed.go`, `web/client.go`, `ui/` | demo funciona |
| 4 | `tests/*.go` | 5 casos + existentes en verde |
| 5 | `README.md`, `docs/ARCHITECTURE.md` | ejemplos sin `IDs` |

**Consumidor (no es trabajo de este plan):** `mjosefa-cms` pasa `IDs` a
`clinicalencounter.Deps`; se actualiza en la etapa I1 del master.
