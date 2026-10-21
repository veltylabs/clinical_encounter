with open("patients.go", "r") as f:
    text = f.read()

text = text.replace('func (m *Module) ListRecentPatients(limit int) ([]*MedicalHistory, error) {', """func (m *Module) ListRecentPatients(tenantId string, limit int) ([]*MedicalHistory, error) {
	if tenantId == "" {
		return nil, ErrTenantRequired
	}""")

text = text.replace('qb := m.db.Query(&rec).OrderBy(MedicalHistory_.AttentionAt).Desc()', """qb := m.db.Query(&rec).
		Where(MedicalHistory_.TenantId).Eq(tenantId).
		OrderBy(MedicalHistory_.AttentionAt).Desc()""")

with open("patients.go", "w") as f:
    f.write(text)
