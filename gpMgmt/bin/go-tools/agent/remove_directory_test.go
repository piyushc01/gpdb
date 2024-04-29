package agent_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/greenplum-db/gpdb/gp/agent"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/testutils"
	"github.com/greenplum-db/gpdb/gp/testutils/exectest"
	"github.com/greenplum-db/gpdb/gp/utils"
)

var (
	dummyDir    = "/tmp/xyz"
	nonExistent = "/path/to/nonexistent/directory"
	errDir      = "/tmp/errDir"
)

// Mocked Stat function
func mockStat(name string) (os.FileInfo, error) {
	if name == dummyDir {
		return nil, nil // Directory exists
	} else if name == nonExistent {
		return nil, os.ErrNotExist // Directory doesn't exist
	} else {
		return nil, fmt.Errorf("mocked error") // Simulate any other error
	}
}

func TestRemoveDirectory(t *testing.T) {

	agentServer := agent.New(agent.Config{
		GpHome: "gpHome",
	})

	err := testutils.CreateDirectoryWithRemoveFail(errDir)
	if err != nil {
		t.Fatalf("failed to create dummy error directory err: %v", err)
	}

	utils.System.Stat = mockStat

	utils.ResetSystemFunctions()

	t.Run("Stat succeeds if directory does not exist", func(t *testing.T) {

		req := &idl.RemoveDirectoryRequest{
			DataDirectory: nonExistent,
		}
		_, err := agentServer.RemoveDirectory(context.Background(), req)

		// Check error
		if err != nil {
			t.Fatalf("stat should ignore file not found error err: %v", err)
		}

	})

	t.Run("Stat succeeds and pgctl status fails but remove directory succeeds", func(t *testing.T) {

		err := testutils.CreateDummyDir(dummyDir)
		if err != nil {
			t.Fatalf("failed to create dummy directory err: %v", err)
		}

		req := &idl.RemoveDirectoryRequest{
			DataDirectory: dummyDir,
		}

		utils.System.ExecCommand = exectest.NewCommand(exectest.Failure)

		_, err = agentServer.RemoveDirectory(context.Background(), req)

		// Check error
		if err != nil {
			t.Fatalf("unable to remove directory")
		}

	})

	t.Run("Stat succeeds and pgctl status succeeds but pgctl stop fails", func(t *testing.T) {

		err := testutils.CreateDummyDir(dummyDir)
		if err != nil {
			t.Fatalf("failed to create dummy directory err: %v", err)
		}

		req := &idl.RemoveDirectoryRequest{
			DataDirectory: dummyDir,
		}

		utils.System.ExecCommand = exectest.NewCommand(exectest.Success)
		utils.System.ExecCommand = exectest.NewCommand(exectest.Failure)

		_, err = agentServer.RemoveDirectory(context.Background(), req)

		// Check error
		expectedErrPrefix := "executing pg_ctl stop"
		if err != nil {
			if !strings.HasPrefix(err.Error(), expectedErrPrefix) {
				t.Fatalf("got %s, want %s", err.Error(), expectedErrPrefix)
			}
		}

	})

	t.Run("Stat succeeds and pgctl status succeeds but pgctl stop succeeds, remove directory fails", func(t *testing.T) {

		err := testutils.CreateDirectoryWithRemoveFail(errDir)
		if err != nil {
			t.Fatalf("failed to create dummy error directory err: %v", err)
		}

		req := &idl.RemoveDirectoryRequest{
			DataDirectory: errDir,
		}

		utils.System.ExecCommand = exectest.NewCommand(exectest.Success)
		utils.System.ExecCommand = exectest.NewCommand(exectest.Success)

		_, err = agentServer.RemoveDirectory(context.Background(), req)

		// Check error
		expectedErrPrefix := "could not remove directory"
		if err != nil {
			if !strings.HasPrefix(err.Error(), expectedErrPrefix) {
				t.Fatalf("got %s, want %s", err.Error(), expectedErrPrefix)
			}
		}

	})

	t.Run("Stat succeeds and pgctl status succeeds but pgctl stop succeeds, remove directory succeeds", func(t *testing.T) {

		err := testutils.CreateDummyDir(dummyDir)
		if err != nil {
			t.Fatalf("failed to create dummy directory err: %v", err)
		}

		req := &idl.RemoveDirectoryRequest{
			DataDirectory: dummyDir,
		}

		utils.System.ExecCommand = exectest.NewCommand(exectest.Success)
		utils.System.ExecCommand = exectest.NewCommand(exectest.Success)

		_, err = agentServer.RemoveDirectory(context.Background(), req)

		// Check error
		expectedErrPrefix := "could not remove directory"
		if err != nil {
			if !strings.HasPrefix(err.Error(), expectedErrPrefix) {
				t.Fatalf("got %s, want %s", err.Error(), expectedErrPrefix)
			}
		}

	})

}
