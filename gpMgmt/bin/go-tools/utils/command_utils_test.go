package utils_test

import (
	"errors"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpdb/gp/testutils/exectest"
	"github.com/greenplum-db/gpdb/gp/utils"
	"github.com/greenplum-db/gpdb/gp/utils/postgres"
)

func init() {
	exectest.RegisterMains(
		CommandSuccess,
		CommandFailure,
	)
}

func TestRunExecCommand(t *testing.T) {
	testhelper.SetupTestLogger()

	cmd := &postgres.Initdb{
		PgData:         "pgdata",
		Encoding:       "encoding",
		Locale:         "locale",
		MaxConnections: 50,
	}

	t.Run("succesfully runs the command", func(t *testing.T) {
		utils.System.ExecCommand = exectest.NewCommand(CommandSuccess)
		defer utils.ResetSystemFunctions()

		out, err := utils.RunExecCommand(cmd, "gphome")
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}

		expected := "success"
		if out.String() != expected {
			t.Fatalf("got %q, want %q", out, expected)
		}
	})

	t.Run("succesfully runs the gp sourced command", func(t *testing.T) {
		var calledUtility, calledArgs string
		utils.System.ExecCommand = exectest.NewCommandWithVerifier(CommandSuccess, func(utility string, args ...string) {
			calledUtility = utility
			calledArgs = strings.Join(args, " ")
		})
		defer utils.ResetSystemFunctions()

		out, err := utils.RunGpSourcedCommand(cmd, "gphome")
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}

		expectedOut := "success"
		if out.String() != expectedOut {
			t.Fatalf("got %q, want %q", out, expectedOut)
		}

		expectedUtility := "bash"
		if calledUtility != expectedUtility {
			t.Fatalf("got %q, want %q", calledUtility, expectedUtility)
		}

		expectedArgs := "-c source gphome/greenplum_path.sh &&"
		if !strings.HasPrefix(calledArgs, expectedArgs) {
			t.Fatalf("got %q, want prefix %q", calledArgs, expectedArgs)
		}
	})

	t.Run("when command fails to execute", func(t *testing.T) {
		utils.System.ExecCommand = exectest.NewCommand(CommandFailure)
		defer utils.ResetSystemFunctions()

		out, err := utils.RunExecCommand(cmd, "gphome")
		if err == nil {
			t.Fatalf("expected error")
		}

		var expectedErr *exec.ExitError
		if !errors.As(err, &expectedErr) {
			t.Errorf("got %T, want %T", err, expectedErr)
		}

		expectedOut := "failure"
		if out.String() != expectedOut {
			t.Fatalf("got %v, want %v", out, expectedOut)
		}
	})
}

func TestAppendIfNotEmpty(t *testing.T) {
	testhelper.SetupTestLogger()

	t.Run("appends the arguments correctly", func(t *testing.T) {
		args := []string{}

		args = utils.AppendIfNotEmpty(args, "string", "value")
		args = utils.AppendIfNotEmpty(args, "int", 1)
		args = utils.AppendIfNotEmpty(args, "float", 1.2)
		args = utils.AppendIfNotEmpty(args, "bool", true)
		args = utils.AppendIfNotEmpty(args, "boolFalse", false)

		expectedArgs := []string{"string", "value", "int", "1", "float", "1.2", "bool"}
		if !reflect.DeepEqual(args, expectedArgs) {
			t.Fatalf("got %+v, want %+v", args, expectedArgs)
		}
	})
}

func CommandSuccess() {
	os.Stdout.WriteString("success")
	os.Exit(0)
}

func CommandFailure() {
	os.Stderr.WriteString("failure")
	os.Exit(1)
}
