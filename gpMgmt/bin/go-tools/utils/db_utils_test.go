package utils_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpdb/gp/testutils"
	"github.com/greenplum-db/gpdb/gp/utils"
	"github.com/jmoiron/sqlx"
)

func TestExecOnDatabase(t *testing.T) {
	testhelper.SetupTestLogger()

	t.Run("succesfully executes the query on a given database", func(t *testing.T) {
		conn, mock := testutils.CreateMockDBConnWithContext(t, context.Background())
		testhelper.ExpectVersionQuery(mock, "7.0.0")

		mock.ExpectExec("SOME QUERY").WillReturnResult(testhelper.TestResult{Rows: 0})
		err := utils.ExecOnDatabaseFunc(conn, "postgres", "SOME QUERY")
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("errors out when fails to connect to the database", func(t *testing.T) {
		var mockdb *sqlx.DB
		expectedErr := errors.New("connection error")

		conn, mock := testutils.CreateMockDBConnWithContext(t, context.Background())
		testhelper.ExpectVersionQuery(mock, "7.0.0")
		conn.DB.Driver = &testhelper.TestDriver{ErrToReturn: expectedErr, DB: mockdb, User: "testrole"}

		err := utils.ExecOnDatabaseFunc(conn, "postgres", "SOME QUERY")
		if !strings.Contains(err.Error(), expectedErr.Error()) {
			t.Fatalf("got %#v, want %#v", err, expectedErr)
		}
	})

	t.Run("errors out when fails to execute the query", func(t *testing.T) {
		expectedErr := errors.New("execution error")

		conn, mock := testutils.CreateMockDBConnWithContext(t, context.Background())
		testhelper.ExpectVersionQuery(mock, "7.0.0")
		mock.ExpectExec("SOME QUERY").WillReturnError(expectedErr)

		err := utils.ExecOnDatabaseFunc(conn, "postgres", "SOME QUERY")
		if !strings.Contains(err.Error(), expectedErr.Error()) {
			t.Fatalf("got %#v, want %#v", err, expectedErr)
		}
	})
}
