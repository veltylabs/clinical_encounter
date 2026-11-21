# Bugs de UI y mejoras — `clinical_encounter`

> Auditoría en vivo y revisión de `clinical_encounter/ui/browser.go` (2026-11-19).

## Bugs encontrados

| # | Sev | Pantalla | Qué se ve en pantalla | Causa raíz verificada |
|---|---|---|---|---|
| CE-1 | 🔴 | Menú / `#clinical_encounter` | **El módulo completo de atención clínica («Ficha médica» / `clinical_encounter`) NO aparece en absoluto en la aplicación**, aunque está registrado en `config/app.go:39` y `config/screens/browser.go:42`. | `clinical_encounter/ui/browser.go:101,176-179` pasa explícitamente `IDs: nil` a `crudview.New(cfg)`, el cual rechaza `IDs == nil` (`webtyp/layout/crudview/crud.go:73-75`), y `Screens()` silencia el error devolviendo `return nil` sin mostrar aviso ni pantalla. |
| CE-2 | 🟡 | `clinical_encounter/ui` | **Placeholder de selector de paciente en inglés literal sin traducir:** `"Select a patient from today's agenda..."`, y ayuda genérica `Help: Label`. | `clinical_encounter/ui/browser.go:146,186` usa texto literal en inglés en `selectsearch.SelectSearch` sin `lang.Text` ni clave en `lang.json`. |
| CE-3 | 🟡 | Arranque de la app | **Hace llamadas RPC síncronas bloqueantes (`blockingCall` + canal) dentro de `Screens()` durante el arranque del WASM**, una por la agenda del día y otra por cada paciente en agenda (`N+1`). Si falla o tarda la red, bloquea todo el arranque o desaparece el módulo entero. | `clinical_encounter/ui/browser.go:114-130` consulta `my_agenda_result` y luego hace un `medical_history` read en un bucle antes de devolver `[]view.Screen`. |

## Sugerencias de mejora justificadas

1. **Corregir `IDs: d.IDs` en `clinical_encounter/ui/browser.go:101` y nunca silenciar un error de `crudview.New` devolviendo `nil`:** mostrar una pantalla de error legible si la configuración falla. *Justificación:* devuelve inmediatamente el módulo clínico completo a la app sin desaparecer en silencio.
2. **Mover la carga de la agenda del día y del historial fuera de `Screens()` (carga diferida al abrir la pantalla o en `Presenter`):** *Justificación:* evita el patrón N+1 bloqueante durante el arranque global de `mjosefa-cms` y permite actualizar la agenda del día sin recargar la página cuando llegan nuevos pacientes durante la jornada.
3. **Traducir el placeholder de `SelectSearch` mediante `lang.Text` y redactar `Help` descriptivo de la pantalla clínica:** *Justificación:* cumplimiento de `translations` y `ui-consistency`.
