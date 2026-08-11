package clinical_encounter

import (
	"github.com/tinywasm/model"
	"github.com/tinywasm/router"
	"github.com/tinywasm/time"
	"github.com/tinywasm/view"
)

// weekdayAbbr/monthAbbr back leadFromUnix's LeadTop/LeadBottom — Spanish,
// matching this module's other user-facing strings. Same table as the
// tinywasm/layout/platformd/modules/medicalhistory demo this productionizes.
var weekdayAbbr = [7]string{"Dom", "Lun", "Mar", "Mié", "Jue", "Vie", "Sáb"} // time.Weekday: 0=Sunday
var monthAbbr = [12]string{"Ene", "Feb", "Mar", "Abr", "May", "Jun", "Jul", "Ago", "Sep", "Oct", "Nov", "Dic"}

// leadFromUnix turns a unix-seconds timestamp into the three-line badge a
// targethour-style list reads instead of a plain label — e.g. "Vie" / "20" /
// "Jul 26". Same shape as the demo's leadFromDate, but AttentionAt is
// already a unix timestamp here (not a "YYYY-MM-DD" string to parse), so
// this reads day/month off FormatCompact's fixed-width "YYYYMMDDHHMMSS"
// instead — `time` bans stdlib per AGENTS.md, and tinywasm/time exposes no
// day/month accessor of its own. A zero or malformed timestamp degrades to
// an empty badge rather than a panic — Item() runs over whatever the store
// holds, and one bad row must not take the whole list down.
func leadFromUnix(sec int64) (top, main, bottom string) {
	if sec <= 0 {
		return "", "", ""
	}
	compact := time.FormatCompact(sec * 1e9)
	if len(compact) != 14 {
		return "", "", ""
	}
	weekday := time.Weekday(sec)
	if weekday < 0 || weekday > 6 {
		return "", "", ""
	}
	day := compact[6:8]
	if day[0] == '0' {
		day = day[1:]
	}
	monthIdx := int(compact[4]-'0')*10 + int(compact[5]-'0') - 1
	if monthIdx < 0 || monthIdx > 11 {
		return "", "", ""
	}
	yy := compact[2:4]
	return weekdayAbbr[weekday], day, monthAbbr[monthIdx] + " " + yy
}

// Item implementa view.Itemizer — el ÚNICO código específico de view que carga este registro. El
// Presenter indexa las filas por ID a partir de esto durante Reload; no hay lookup manual byID/WithFill.
// El badge lateral lleva la fecha (leadFromUnix); Label el motivo, Description el estado — el nombre
// del doctor no aparece aquí porque, a diferencia de la demo, el consumidor decide qué snapshot mostrar
// (ver modules/clinical_encounter/view.go de mjosefa-cms, que sí conoce DoctorNameSnapshot).
func (it *MedicalHistory) Item() view.Item {
	top, main, bottom := leadFromUnix(it.AttentionAt)
	return view.Item{
		ID: it.Id, Label: it.Reason, Description: it.Status,
		LeadTop: top, LeadMain: main, LeadBottom: bottom,
	}
}

// NewView construye el Presenter del historial médico — el motor agnóstico de tecnología que envuelve
// un renderer (tinywasm/layout/rightpanel, o cualquier otro). Este módulo lo construye (solo
// view+model+router); la app decide qué renderer lo dibuja.
func NewView(caller router.Caller) view.Presenter {
	record := &MedicalHistory{}

	return view.New(
		caller,
		record,
		OpListVisitsByPatient,
		func() model.ModelSlice { return &MedicalHistoryList{} },
		view.WithTitle("Historial clínico"),
		view.WithSaveOp(OpCreateVisit),
	)
}
