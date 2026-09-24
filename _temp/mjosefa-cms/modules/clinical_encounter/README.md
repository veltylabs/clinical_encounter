# clinical_encounter — cómo debería funcionar la agenda del doctor

Este módulo es la ficha clínica (`medical_history`): lo que el doctor escribe
durante y después de atender a un paciente. Este README documenta cómo
funciona "qué paciente estoy atendiendo" — ya implementado — para que quede
una sola fuente de verdad y no se repita el error de diseño que motivó este
documento (ver "Error de diseño descartado" abajo).

## La idea

El doctor NUNCA busca un paciente. Abre su agenda del día y ve únicamente los
pacientes que tienen una **reserva confirmada** con él para hoy — el mismo
principio que ya aplica `webtyp/app-demo/modules/medicalhistory` (demo de
referencia, con datos falsos) pero con datos reales:

```
appointment_booking.list_reservations_by_staff(staff_id, from=hoy, to=hoy)
    → filtrar Status == CONFIRMED
    → para cada reserva: patient_directory.get_patient(reservation.client_id)
    → picker: "09:00  Juan Pérez  12.345.678-9"
```

Seleccionar una fila del picker:
1. Filtra la lista de fichas al historial de ESE paciente
   (`clinical_encounter.list_medical_history` con `patient_id`).
2. Pre-llena `patient_id`, `patient_name_snapshot` y `patient_rut_snapshot`
   en el formulario de "nueva ficha" — el doctor no vuelve a escribir un dato
   que el sistema ya sabe.

`doctor_id` / `doctor_name_snapshot` / `doctor_specialty_snapshot` se
resuelven una sola vez al construir la vista, buscando en
`staff_manager.list_staff` la fila cuyo `user_id` coincide con el usuario
autenticado (`auth.OpMe` → `profile.Id`) — el doctor tampoco escribe su
propio nombre.

## Por qué el picker no puede ser síncrono contra el servidor

`view.Presenter.Filter(term string) []view.Item` es **síncrono** —
`crudview` lo llama directamente para redibujar la lista
(`v.list.SetItems(v.Presenter.Filter(term))`). Una llamada de red real
(`router.Caller.Call`) es asíncrona por contrato, y no existe combinación de
canal + espera que la vuelva síncrona sin bloquear el event loop del
navegador — WASM en el navegador es de un solo hilo cooperativo; bloquear
dentro de un callback ya en curso lo cuelga en silencio (sin panic, sin log).
La única excepción documentada en este repo es `config/client.go`'s
`auth.OpMe`, y es legal únicamente porque corre en `main()`, ANTES de que el
event loop arranque — no aplica aquí, porque el picker vive dentro de un
callback de render que sí corre después.

La solución (igual patrón que `appointment_booking/bookingview.go`'s
`freeSlotsCache`): **pre-cargar todo el historial de cada paciente de la
agenda al construir la vista**, una sola vez, antes de que el picker exista.
El día de un doctor es acotado (unas decenas de pacientes como mucho), así
que N llamadas de red al abrir el módulo es aceptable; `Filter(term)` después
solo hace un scan lineal sobre datos ya en memoria — ninguna llamada de red
dentro de `Filter`.

## ~~Gap~~ RESUELTO 2026-09-23: crudview siembra el borrador nuevo

`newAction()` limpiaba el formulario y llamaba a `OnNew func()` — sin
argumentos y sin acceso al formulario — así que la app no tenía dónde inyectar
nada. Era un hueco del framework, no de esta app, y se cerró ahí y no con un
parche local. `crudview.Config.NewRecord` (layout v0.2.52) cerró el hueco: es una fábrica que
devuelve el registro con el que arranca el "+". Acá se cablea con el paciente
que el picker dejó elegido más el doctor de la sesión, así que la ficha nueva
nace con `patient_id`, nombre, RUT, `doctor_id`, nombre y especialidad ya
puestos. Los campos siguen siendo `input.Text()` editables a propósito: sembrar
no es bloquear, y una corrección puntual sigue siendo posible.

## Bug ya corregido: `attention_at`

`MedicalHistoryModel.attention_at` es `model.Int()` (no `input.*`), así que
`webtyp/form` nunca lo dibuja — toda ficha nueva llegaba al servidor con
`AttentionAt == 0`. `CreateVisit` lo rechazaba (`ErrMissingArgs`), lo que
significaba que el botón "+" de este módulo **nunca guardaba nada** en
producción. Corregido en `veltylabs/clinical_encounter` v0.1.9: `AttentionAt`
por defecto es `time.Now()` — una atención se crea porque está pasando ahora,
no es un dato que el doctor deba escribir.

## Error de diseño descartado

Una iteración anterior de este trabajo propuso un buscador de paciente por
RUT (reutilizando `patient_directory.find_patient_by_rut`). Es el patrón
equivocado: obliga al doctor a escribir algo que el sistema ya sabe por la
reserva confirmada. El diseño correcto es agenda-primero, cero búsqueda —
documentado arriba.

## Estado

- [x] `attention_at` default a `time.Now()` (clinical_encounter v0.1.9).
- [x] Picker de agenda del día (reservas confirmadas, pre-carga de historial
      por paciente).
- [x] Resolución de identidad del doctor vía `staff_manager` + `auth.OpMe`.
- [x] Pre-llenado del borrador nuevo (`crudview.Config.NewRecord`,
      layout v0.2.52), fijado por
      `tests/clinical_encounter_agenda_wasm_test.go`.

Lo único que este módulo NO valida todavía es que el picker se vea con datos
reales en el navegador: la base de desarrollo no tiene reservas confirmadas
(ver **B3** en `docs/MASTER.md`).
