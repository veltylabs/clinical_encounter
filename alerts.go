package clinical_encounter

import "webtyp.com/notify"

const (
	AlertSupplyPending = "clinical_encounter.supply.pending_deduction"
)

const (
	alertSubject = "Supplies pending deduction"
	alertBody    = "An attention was completed with supplies that could not be taken from stock. Open Supplies to retry them."
)

func AlertKinds(areaCode string) []notify.Kind {
	return []notify.Kind{
		{
			Code:        AlertSupplyPending,
			Area:        areaCode,
			Name: "Atenciones con insumos sin rebajar",
		},
	}
}
