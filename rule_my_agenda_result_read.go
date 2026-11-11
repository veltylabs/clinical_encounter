package clinical_encounter
import (
    "webtyp.com/crud"
    "webtyp.com/model"
)
func (p *MyAgendaResult) Read(a crud.CurrentUser, filter *MyAgendaResult, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	if filter.Date == "" {
		return nil, ErrMissingArgs
	}
	res, err := m.MyAgenda(MyAgendaArgs{TenantId: filter.TenantId, UserId: a.UserID, Date: filter.Date})
	if err != nil {
		return nil, err
	}
	return &MyAgendaResultList{*res}, nil
}
