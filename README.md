# clinical_encounter
<img src="docs/img/badges.svg">

Contexto de gestión de encuentros clínicos para el ecosistema Velty.

## ¿Por qué existe?

Este módulo gestiona la historia clínica y las visitas de pacientes, permitiendo crear y gestionar encuentros, suministros médicos y atención en salud.

## Antes de comenzar

Este módulo exporta su pantalla de navegador bajo `ui/` y datos de demostración bajo `seed/`:

- `ui.ID`: `"clinical_encounter"`
- `ui.Label`: `"Historial Clínico"`
- `ui.Browser(caller, ids, tenantID, userID)`: devuelve el `platformd.UIModule` para el navegador.
- `seed.Load(module, upstream)`: puebla visitas de demostración utilizando los propios métodos del módulo.

Ejecute `webtyp` en la raíz del repositorio para abrir la demostración — en el navegador, en memoria, sin inicio de sesión.

## Guía paso a paso

El método `ImportVisit` permite importar una visita cerrada de un sistema previo con sus fechas reales (`AttentionAt`, `StartedAt`, `FinishedAt`). No usa la fecha actual "now", evitando que el historial de visitas pasadas aparezcan cerradas el día de la importación.

Ejemplo:
```go
v, err := enc.ImportVisit(MedicalHistory{
    Id: "8000000000000000005", TenantId: t, PatientId: p, DoctorId: d, AttentionAt: at,
    Reason: "control", PatientNameSnapshot: "Ana", PatientRutSnapshot: "12345678-5",
    DoctorNameSnapshot: "Dra. Pérez"
})
```

## Mapa de documentación

- [Arquitectura (ARCHITECTURE.md)](docs/ARCHITECTURE.md)
- [Pruebas (TESTS.md)](docs/TESTS.md)
- [UI (UI.md)](docs/UI.md)