package cli

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/utils"
)

var (
	TerminationRequested bool
	SigtermReceived      bool
	ch                   = make(chan os.Signal, 1)
)

// Specific type of error for interrupt
type ErrorUserTermination struct{}

func (*ErrorUserTermination) Error() string {
	return "program was terminated by the user"
}

func SetSignalHandler(ctrl *StreamController) {
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	signal.Ignore(syscall.SIGHUP)

	go func() {
		for sig := range ch {
			HandleSignal(sig, ctrl)
		}
	}()
}

// HandleSignal handles the given signal and performs the necessary actions based on the signal received.
// If the signal is SIGINT, it pauses the hub stream parsing and prompts the user to continue terminating the current execution.
// If the signal is SIGTERM, it sets the TerminationRequested flag to true.
// For any other signal, it logs the signal received.
// Note: This function assumes that the TerminationRequested flag is defined and accessible from the current scope.
func HandleSignal(sig os.Signal, ctrl *StreamController) {
	switch sig {
	case syscall.SIGINT:
		signal.Ignore(syscall.SIGINT)
		logMessage := "received an interrupt signal"
		promptText := "Do you want to continue terminating the current execution?"

		if ctrl != nil {
			// pause the hub stream parsing so we could display a prompt
			ctrl.SetState(streamPaused)
			// wait until the stream is paused
			ctrl.WaitUntilPaused()
			gplog.Warn(logMessage)
			terminate := utils.AskUserYesOrNo(promptText)
			if !terminate {
				signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
				// resume the stream
				ctrl.SetState(streamRunning)
			} else {
				TerminationRequested = true
				ctrl.SetState(streamDiscard) // discard all the stream responses we get while we waited for the user confirmation
			}
		} else {
			gplog.Warn(logMessage)
			TerminationRequested = utils.AskUserYesOrNo(promptText)
		}

	case syscall.SIGTERM:
		gplog.Warn("received a termination signal")
		TerminationRequested = true
		SigtermReceived = true

	default:
		gplog.Debug("signal received: %s\n", sig.String())
	}
}
