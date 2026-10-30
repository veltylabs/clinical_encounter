package migrate

import (
	"webtyp.com/ddl"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
)

// Migrate reconcilia el esquema de la base de datos de clinical_encounter:
// MedicalHistory y EncounterSupply.
//
// Deliberadamente NO es llamado por New, y deliberadamente vive en su propio
// paquete en lugar de un archivo nuevo en el paquete raíz: nada en un
// árbol de construcción WASM de una aplicación consumidora (su view.go, el cual importa el paquete raíz
// de clinical_encounter para clinicalencounter.NewView) importa
// "github.com/veltylabs/clinical_encounter/migrate" jamás — por lo que webtyp.com/ddl
// nunca entra en ese árbol de construcción, independiente de los build tags del
// lado del consumidor.
//
// conn es un ddl.Execer, no un *orm.DB, así que un transporte en tiempo de despliegue que puede
// ejecutar solo DDL lo satisface. El RawConn() de un *orm.DB también lo satisface,
// para los llamadores locales/de test:
//
//	conn, _ := postgres.Open(dsn)
//	compiler, _ := conn.(ddl.Compiler)
//	err := migrate.Migrate(conn, compiler)
//
// Ejemplo: err := migrate.Migrate(conn, compiler)
func Migrate(conn ddl.Execer, ddlCompiler ddl.Compiler) error {
	return ddl.New(conn, ddlCompiler).Sync(&clinicalencounter.MedicalHistory{}, &clinicalencounter.EncounterSupply{})
}
