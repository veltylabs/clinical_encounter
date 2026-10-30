# Arquitectura de Clinical Encounter

## Alcance del dominio

`clinical_encounter` es el propietario del registro de la visita médica para el ecosistema Velty: el historial de las
atenciones de un paciente (`MedicalHistory`) y los argumentos para abrir uno nuevo (`CreateVisitArgs`). Es
la fuente de verdad de la que otros módulos leerían si necesitaran el historial de visitas de un paciente — ningún
otro módulo lo lee o escribe hoy.

## Entidades Principales

- **`MedicalHistory`** (tabla `medical_history`): una fila por visita de paciente. Identidad (`Id`,
  `PatientId`, `DoctorId`, opcional suave `ReservationId`), estado FSM (`Status`, ver abajo),
  programación (`AttentionAt` — marca de tiempo Unix, fecha y hora unificadas), contenido clínico (`Reason`,
  `Diagnostic`, `Prescription`, `Cie10Code`), tiempo (`StartedAt`, `FinishedAt`), capturas inmutables
  en el tiempo tomadas en la creación de la visita (`PatientNameSnapshot`, `PatientRutSnapshot`,
  `DoctorNameSnapshot`, `DoctorSpecialtySnapshot`) para que el registro se lea correctamente aunque el perfil del
  paciente/doctor cambie después, y `UpdatedAt`. No contiene widgets de interfaz de usuario — se muestra,
  nunca se edita a través de un formulario de forma directa.
- **`CreateVisitArgs`**: los argumentos de transporte para abrir un nuevo `MedicalHistory` — los mismos campos
  clínicos menos `Status`/`StartedAt`/`FinishedAt`/`Cie10Code`/`UpdatedAt` (asignado por el servidor o aún
  no recopilado en el momento de la creación). `CreateVisitArgs.Id` es generado por el llamador. Repetir la creación con el mismo ID devuelve la visita guardada. Cada campo tiene un widget UI (`input.Text()`/
  `input.Number()`/`input.Textarea()` para `Diagnostic`/`Prescription`) para que un formulario renderizado
  por `webtyp/form` nunca esté vacío silenciosamente.

## El Visit FSM (`const.go` — no modificado por esta migración)

`const.go` declara el ciclo de vida alineado con FHIR como simples constantes de cadena — este módulo no
toca ese archivo, solo se construye sobre este:

```go
StatusCreated    = "created"     // agendado, aún no en clínica
StatusArrived    = "arrived"     // paciente registrado en recepción
StatusTriaged    = "triaged"     // enfermera tomó signos vitales, esperando al doctor
StatusInProgress = "in_progress" // doctor comenzó la consulta
StatusCompleted  = "completed"
StatusCancelled  = "cancelled"
```

`MedicalHistory.Status` contiene uno de estos. Al momento de escribir esto `const.go` declara sólo el
conjunto `Status*` — no hay un grupo de constantes `Event*` y no hay tabla de transición (`visitTransitions`) en
el repo; un doc previo referenció a una pero no existe en el código. `CreateVisit` es el único método
que asigna un estado hoy (`StatusCreated`, en la creación) — no existen métodos de transición
(arrive/triage/start/complete/cancel) aún. Por consiguiente, `router.OpModule` en este módulo expone
solo las operaciones respaldadas por métodos reales (create/list/get) — no inventa operaciones de transición para
estados FSM que no tienen lógica de negocios correspondiente aún. Añadir ésos es un cambio a futuro, separado.

## Patrones

- **Arnés de módulo reusable**: el módulo está acoplado solo a contratos publicados, no a infraestructura
  concreta — ver `AGENTS.md` (esta raíz del repo) para la lista blanca/lista negra completa:
    - `orm.DB` para el almacenamiento (agnóstico al backend — envuelve cualquier `storage.Conn` que la app inyecte).
    - `ddl.Sync` (sobre `db.RawConn()`, protegido por una aserción de tipo `ddl.Compiler`) para la
      migración del esquema propia del módulo con `Sync` (crea lo que falta y agrega columnas nuevas).
    - `router.OpModule` (`ModelName()` + `MountOps(reg router.OpRegistry)`) para el transporte — el
      módulo nunca importa `webtyp/mcp` y nunca implementa `router.Router`/`router.APIModule`.
    - `model.IDGenerator` para la identidad (`Deps.IDs`, requerido).
    - `events.Publisher` para actualizaciones orientadas a eventos (`Deps.Publisher`, opcional — `nil` deshabilita
      la publicación silenciosamente). La propia interfaz `EventPublisher` de este módulo (previamente en
      `publish.go`) está retirada en favor de este contrato.
    - `view.Presenter` (`NewView(caller router.Caller) view.Presenter`) para UI, construido solo con
      `view`+`model`+`router` — la aplicación elige el renderizador. Reemplaza al paquete anterior `web/`,
      que construía una UI concreta `webtyp/dom` + `webtyp/layout/rightpanel` directamente dentro del
      módulo y es eliminado — ver `docs/PLAN.md` Etapa 0/4 para el historial completo de reconciliación.
    - Las pruebas corren contra `storage/mem` (`orm.New(mem.New())`), nunca un driver concreto, y nunca un
      fake hand-rolled `orm.Executor`/`orm.Compiler` (el retirado paquete `web/mockdb` hizo esto
      redundantemente — `storage/mem` es el equivalente upstream mantenido).
- **Campos Snapshot, no live joins**: los campos `*Snapshot` en `MedicalHistory` congelan la identidad
  del paciente/doctor en el momento de la creación de la visita. Este módulo nunca consulta a una tabla
  `patient`/`doctor` — esos pertenecen a otros módulos que este no importa.
- **Sin campo multi-tenant hoy**: a diferencia de `service_catalog`, `MedicalHistory`/`CreateVisitArgs`
  no llevan campo `TenantId` en el esquema actual. Esta migración porta las columnas existentes mecánicamente
  (misma tabla, mismas columnas, mismo comportamiento) y no agrega uno — ver `docs/PLAN.md` Etapa 1. Si
  este módulo necesita aislamiento de tenant después, eso es un cambio de esquema explícito y separado, no un
  efecto secundario de la migración de arnés.

## Ops (mediante `MountOps`)

| Op | Acción | Recurso | Descripción |
|---|---|---|---|
| `create_visit` | `c` | `medical_history` | Crear una nueva visita desde `CreateVisitArgs` — asigna `Status = StatusCreated`. Responde 409 con `ErrIdTaken` por colisión de ID. |
| `get_medical_history` | `r` | `medical_history` | Obtener un registro de visita por ID |
| `list_medical_history` | `r` | `medical_history` | Listar el historial de visitas de un paciente |

No existen operaciones de transición de estado (`arrive`/`triage`/`start`/`complete`/`cancel`) aún — ver "El Visit
FSM" arriba. Añadirlas es un cambio a futuro con su propio plan, una vez que existan los métodos `*Module`
correspondientes.

## Ejemplo de Raíz de Composición

```go
enc, _ := clinicalencounter.New(db, clinicalencounter.Deps{
    Publisher: eventPublisher, // events.Publisher, nil deshabilita publicación
})
enc.MountOps(opRegistry) // router.OpRegistry
view := enc.NewView(caller) // router.Caller -> view.Presenter, renderizado por lo que la aplicación elija
```

## Fuera de Alcance

- `const.go` (constantes FSM) — descritas arriba, nunca editado por esta migración de arnés del módulo.
- `dicom.go` — actualmente un stub vacío (`package clinical_encounter`, sin código); dejado intacto.
- `clinical_encounterOld` — un directorio hermano heredado sin importadores en ninguna parte del ecosistema.
  No es parte de este repo y no debe ser leído, referenciado o migrado por ningún cambio aquí.
