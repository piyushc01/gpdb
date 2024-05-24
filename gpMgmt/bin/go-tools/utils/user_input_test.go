package utils_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/greenplum-db/gpdb/gp/testutils"
	"github.com/greenplum-db/gpdb/gp/utils"
)

func TestAskUserYesOrNo(t *testing.T) {
	t.Run("correctly displays the prompt text", func(t *testing.T) {
		input := fmt.Sprintln("y")
		resetStdin := testutils.MockStdin(t, input)
		defer resetStdin()

		buffer, writer, resetStdout := testutils.CaptureStdout(t)
		defer resetStdout()

		promptText := "some prompt text"
		utils.AskUserYesOrNo(promptText)
		writer.Close()

		stdout := <-buffer
		expectedStdout := fmt.Sprintf("\n%s  Yy|Nn: ", promptText)
		if stdout != expectedStdout {
			t.Fatalf("got %s, want %s", stdout, expectedStdout)
		}
	})

	t.Run("returns true when the user wants to proceed", func(t *testing.T) {
		for _, input := range []string{fmt.Sprintln("y"), fmt.Sprintln("Y")} {
			resetStdin := testutils.MockStdin(t, input)
			defer resetStdin()

			result := utils.AskUserYesOrNo("")
			if !result {
				t.Fatalf("got %t, want true", result)
			}
		}
	})

	t.Run("returns false when the user wants to cancel", func(t *testing.T) {
		for _, input := range []string{fmt.Sprintln("n"), fmt.Sprintln("N")} {
			resetStdin := testutils.MockStdin(t, input)
			defer resetStdin()

			result := utils.AskUserYesOrNo("")
			if result {
				t.Fatalf("got %t, want false", result)
			}
		}
	})

	t.Run("returns false when fails to read the input", func(t *testing.T) {
		resetStdin := testutils.MockStdin(t, "")
		defer resetStdin()

		buffer, writer, resetStdout := testutils.CaptureStdout(t)
		defer resetStdout()

		result := utils.AskUserYesOrNo("")
		writer.Close()

		if result {
			t.Fatalf("got %t, want false", result)
		}

		stdout := <-buffer
		expectedStdout := `Failed to read input: EOF, defaulting to no`
		if !strings.Contains(stdout, expectedStdout) {
			t.Fatalf("got %q, want %q", stdout, expectedStdout)
		}
	})

	t.Run("retries when input is invalid", func(t *testing.T) {
		input := "invalid\ninvalid\ny\n"
		resetStdin := testutils.MockStdin(t, input)
		defer resetStdin()

		buffer, writer, resetStdout := testutils.CaptureStdout(t)
		defer resetStdout()

		result := utils.AskUserYesOrNo("")
		writer.Close()

		if !result {
			t.Fatalf("got %t, want true", result)
		}

		stdout := <-buffer
		expectedStdout := `invalid input "invalid"`
		expectedCount := 2
		if strings.Count(stdout, expectedStdout) != expectedCount {
			t.Fatalf("got %q, want %q to be present %d times", stdout, expectedStdout, expectedCount)
		}
	})
}
