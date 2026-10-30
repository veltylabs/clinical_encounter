package clinical_encounter

import "webtyp.com/orm"

// defaultRecentPatientsLimit limita el escaneo cuando el llamador pasa Limit<=0.
// medical_history comienza vacío en producción (ver docs/PLAN_MODEL_MIGRATION.md
// — este módulo no hereda las filas de la tabla heredada medicalhistory), por lo que
// un escaneo completo de tabla y deduplicación en Go es barato hoy; revisar con una
// consulta real una vez que el volumen haga que eso sea falso.
const defaultRecentPatientsLimit = 200

// ListRecentPatients devuelve los pacientes atendidos más recientemente, una fila
// por patient_id (deduplicado, la más reciente attention_at gana), para un picker que
// busca "qué paciente tiene una ficha existente" en lugar de buscar uno
// por un ID que nadie ha memorizado. Obtenido enteramente de los datos propios
// de este módulo — sin dependencia cruzada de módulos en un directorio de pacientes/clientes, que
// aún no existe (ver ARCHITECTURE.md).
// Ejemplo: module.ListRecentPatients("tenant1", 10)
func (m *Module) ListRecentPatients(tenantId string, limit int) ([]*MedicalHistory, error) {
	if tenantId == "" {
		return nil, ErrTenantRequired
	}
	if limit <= 0 {
		limit = defaultRecentPatientsLimit
	}

	var rec MedicalHistory
	qb := m.db.Query(&rec).
		Where(MedicalHistory_.TenantId).Eq(tenantId).
		OrderBy(MedicalHistory_.AttentionAt).Desc()
	rows, err := ReadAllMedicalHistory(qb)
	if err != nil {
		if orm.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	// Slice, no mapa — la regla de "cero mapas" en AGENTS.md no tiene excepciones.
	// patients está limitado a limit (≤200 por defecto), por lo que el escaneo lineal
	// ya visto a continuación es como máximo limit² comparaciones — el mismo
	// compromiso que el weeklyForDay propio de appointment_booking hace sobre un
	// slice similarmente pequeño y limitado.
	patients := make([]*MedicalHistory, 0, limit)
	for _, r := range rows {
		alreadySeen := false
		for _, p := range patients {
			if p.PatientId == r.PatientId {
				alreadySeen = true
				break
			}
		}
		if alreadySeen {
			continue
		}
		patients = append(patients, r)
		if len(patients) >= limit {
			break
		}
	}
	return patients, nil
}
