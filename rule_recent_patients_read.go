package clinical_encounter
import (
    "webtyp.com/crud"
    "webtyp.com/model"
)
func (p *RecentPatients) Read(a crud.CurrentUser, filter *RecentPatients, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	limit := 50
	if filter.Limit > 0 {
		limit = int(filter.Limit)
	}
	res, err := m.ListRecentPatients(filter.TenantId, limit)
	if err != nil {
		return nil, err
	}
	return (*MedicalHistoryList)(&res), nil
}
