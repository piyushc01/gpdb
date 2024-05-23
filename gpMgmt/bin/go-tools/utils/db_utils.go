package utils

import (
	"context"

	"github.com/greenplum-db/gp-common-go-libs/dbconn"
)

var (
	ExecOnDatabaseFunc       = execOnDatabase
	newDBConnFromEnvironment = dbconn.NewDBConnFromEnvironment
)

type DBConnWithContext struct {
	DB  *dbconn.DBConn
	Ctx context.Context
}

func NewDBConnWithContext(ctx context.Context, dbname string) DBConnWithContext {
	return DBConnWithContext{
		DB:  newDBConnFromEnvironment(dbname),
		Ctx: ctx,
	}
}

func execOnDatabase(conn *DBConnWithContext, dbname string, query string) error {
	conn.DB.DBName = dbname
	if err := conn.DB.Connect(1); err != nil {
		return err
	}
	defer conn.DB.Close()

	if _, err := conn.DB.ExecContext(conn.Ctx, query); err != nil {
		return err
	}

	return nil
}

// used only for testing
func SetExecOnDatabase(customFunc func(*DBConnWithContext, string, string) error) {
	ExecOnDatabaseFunc = customFunc
}

func ResetExecOnDatabase() {
	ExecOnDatabaseFunc = execOnDatabase
}

func SetNewDBConnFromEnvironment(customFunc func(dbname string) *dbconn.DBConn) {
	newDBConnFromEnvironment = customFunc
}

func ResetNewDBConnFromEnvironment() {
	newDBConnFromEnvironment = dbconn.NewDBConnFromEnvironment
}
