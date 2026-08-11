package clinical_encounter

import "github.com/tinywasm/orm"

// defaultRecentPatientsLimit caps the scan when the caller passes Limit<=0.
// medical_history starts empty in production (see docs/PLAN_MODEL_MIGRATION.md
// — this module does not inherit the legacy medicalhistory table's rows), so
// a full-table scan-and-dedupe in Go is cheap today; revisit with a real
// query once volume makes that untrue.
const defaultRecentPatientsLimit = 200

// ListRecentPatients returns the most recently attended patients, one row
// per patient_id (deduped, most recent attention_at wins), for a picker that
// searches "which patient has an existing ficha" rather than looking one up
// by an ID nobody has memorized. Sourced entirely from this module's own
// data — no cross-module dependency on a patient/client directory, which
// does not exist yet (see ARCHITECTURE.md).
func (m *Module) ListRecentPatients(limit int) ([]*MedicalHistory, error) {
	if limit <= 0 {
		limit = defaultRecentPatientsLimit
	}

	var rec MedicalHistory
	qb := m.db.Query(&rec).OrderBy(MedicalHistory_.AttentionAt).Desc()
	rows, err := ReadAllMedicalHistory(qb)
	if err != nil {
		if err == orm.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	// Slice, no map — the "zero map" rule in AGENTS.md has no exceptions.
	// patients is capped at limit (≤200 by default), so the linear
	// already-seen scan below is at most limit² comparisons — the same
	// trade-off appointment_booking's own weeklyForDay makes over a
	// similarly small, bounded slice.
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
