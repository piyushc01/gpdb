package cli_test

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpdb/gp/cli"
	"github.com/greenplum-db/gpdb/gp/testutils"
)

func TestHandleSignal(t *testing.T) {
	_, _, logfile := testhelper.SetupTestLogger()

	// initialize the states based on the stream controller implementation
	const (
		streamRunning = iota
		streamPaused
		streamDiscard
	)

	t.Run("prompts the user and correctly sets the TerminationRequested flag when the signal is SIGINT", func(t *testing.T) {
		defer func() {
			cli.TerminationRequested = false
		}()

		buffer, writer, resetStdout := testutils.CaptureStdout(t)
		defer resetStdout()

		// when user selects yes
		resetStdin := testutils.MockStdin(t, "y\n")
		defer resetStdin()

		cli.HandleSignal(syscall.SIGINT, nil)
		writer.Close()

		testutils.AssertLogMessage(t, logfile, `\[WARNING\]:-received an interrupt signal`)

		stdout := <-buffer
		expectedPromptText := "Do you want to continue terminating the current execution?"
		if !strings.Contains(stdout, expectedPromptText) {
			t.Fatalf("got %s, want %s", stdout, expectedPromptText)
		}

		if !cli.TerminationRequested {
			t.Fatalf("got %t, want true", cli.TerminationRequested)
		}

		// when user selects no
		resetStdin = testutils.MockStdin(t, "n\n")
		defer resetStdin()

		cli.HandleSignal(syscall.SIGINT, nil)
		if cli.TerminationRequested {
			t.Fatalf("got %t, want false", cli.TerminationRequested)
		}
	})

	t.Run("correctly sets the stream controller states when user selects yes", func(t *testing.T) {
		defer func() {
			cli.TerminationRequested = false
		}()

		writer, resetStdin := testutils.MockStdinWithWriter(t)
		defer resetStdin()

		ctrl := cli.NewStreamController()
		go cli.HandleSignal(syscall.SIGINT, ctrl) // the code will be blocked until we indicate that the stream is paused using ctrl.SetPaused()

		// the stream should be paused initially to display the prompt
		time.Sleep(100 * time.Millisecond)
		if ctrl.State() != streamPaused {
			t.Fatalf("expected stream to be paused")
		}

		ctrl.Paused() // indicate that the stream has been paused so we can display the user prompt

		// enter yes in the prompt
		_, err := writer.WriteString(fmt.Sprintln("y"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		writer.Close()

		// the stream should be discarded if user wants to continue
		time.Sleep(100 * time.Millisecond)
		if ctrl.State() != streamDiscard {
			t.Fatalf("expected stream to be discarded")
		}

		testutils.AssertLogMessage(t, logfile, `\[WARNING\]:-received an interrupt signal`)

		if !cli.TerminationRequested {
			t.Fatalf("got %t, want true", cli.TerminationRequested)
		}
	})

	t.Run("correctly sets the stream controller states when user selects no", func(t *testing.T) {
		defer func() {
			cli.TerminationRequested = false
		}()

		writer, resetStdin := testutils.MockStdinWithWriter(t)
		defer resetStdin()

		ctrl := cli.NewStreamController()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			// the code will be blocked until we indicate that the stream is paused using ctrl.SetPaused()
			cli.HandleSignal(syscall.SIGINT, ctrl)
		}()

		// the stream should be paused initially to display the prompt
		time.Sleep(100 * time.Millisecond)
		if ctrl.State() != streamPaused {
			t.Fatalf("expected stream to be paused")
		}

		ctrl.Paused() // indicate that the stream has been paused so we can display the user prompt

		// enter no in the prompt
		_, err := writer.WriteString(fmt.Sprintln("n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		writer.Close()
		wg.Wait()

		// the stream should be resumed if user does not want to continue
		if ctrl.State() != streamRunning {
			t.Fatalf("expected stream to be resumed")
		}

		testutils.AssertLogMessage(t, logfile, `\[WARNING\]:-received an interrupt signal`)

		if cli.TerminationRequested {
			t.Fatalf("got %t, want false", cli.TerminationRequested)
		}
	})

	t.Run("when there is a SIGTERM signal", func(t *testing.T) {
		defer func() {
			cli.TerminationRequested = false
		}()

		cli.HandleSignal(syscall.SIGTERM, nil)
		if !cli.TerminationRequested {
			t.Fatalf("got %t, want true", cli.TerminationRequested)
		}
	})

	t.Run("when there is an unhandled signal", func(t *testing.T) {
		defer func() {
			cli.TerminationRequested = false
		}()

		for _, sig := range []os.Signal{syscall.SIGALRM, syscall.SIGABRT} {
			cli.HandleSignal(sig, nil)

			// do not modify the flag
			if cli.TerminationRequested {
				t.Fatalf("got %t, want false", cli.TerminationRequested)
			}

			expectedLog := fmt.Sprintf(`\[DEBUG\]:-signal received: %s`, sig)
			testutils.AssertLogMessage(t, logfile, expectedLog)
		}
	})
}

func TestSetSignalHandler(t *testing.T) {
	_, _, logfile := testhelper.SetupTestLogger()

	t.Run("correctly sets the signal handler", func(t *testing.T) {
		defer func() {
			cli.TerminationRequested = false
		}()

		cli.SetSignalHandler(nil)

		// ignores SIGHUP
		err := syscall.Kill(os.Getpid(), syscall.SIGHUP)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		time.Sleep(100 * time.Millisecond)

		// do not modify the flag
		if cli.TerminationRequested {
			t.Fatalf("got %t, want false", cli.TerminationRequested)
		}
		testutils.AssertLogMessageNotPresent(t, logfile, fmt.Sprint(syscall.SIGHUP))

		// handles SIGTERM
		err = syscall.Kill(os.Getpid(), syscall.SIGTERM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		time.Sleep(100 * time.Millisecond)

		if !cli.TerminationRequested {
			t.Fatalf("got %t, want true", cli.TerminationRequested)
		}
		testutils.AssertLogMessage(t, logfile, `\[WARNING\]:-received a termination signal`)
	})
}
