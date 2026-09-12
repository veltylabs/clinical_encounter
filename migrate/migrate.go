package migrate

import (
	"webtyp.com/ddl"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
)

// Migrate reconciles the database schema clinical_encounter owns:
// MedicalHistory.
//
// It is deliberately NOT called by New, and deliberately lives in its own
// package rather than a new file in the root package: nothing on a
// consuming app's WASM build path (its view.go, which imports the root
// clinical_encounter package for clinicalencounter.NewView) ever imports
// "github.com/veltylabs/clinical_encounter/migrate" — so webtyp.com/ddl
// never enters that build graph, regardless of build tags on the
// consumer's side.
//
// conn is a ddl.Execer, not an *orm.DB, so a deploy-time transport that can
// only execute DDL satisfies it. An *orm.DB's RawConn() also satisfies it,
// for local/test callers:
//
//	conn, _ := postgres.Open(dsn)
//	compiler, _ := conn.(ddl.Compiler)
//	err := migrate.Migrate(conn, compiler)
func Migrate(conn ddl.Execer, ddlCompiler ddl.Compiler) error {
	return ddl.New(conn, ddlCompiler).CreateTable(&clinicalencounter.MedicalHistory{})
}
