package cli

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/utils"
)

const (
	streamRunning = iota
	streamPaused
	streamDiscard
)

// StreamController represents a controller for managing a stream.
// The stream can be controlled with the help of 3 states:
//   - streamRunning: indicates that the stream is currently running
//   - streamPaused: indicates that the stream is currently paused
//   - streamDiscard: discard any responses we have got except any errors
type StreamController struct {
	mu     sync.Mutex
	state  int
	paused chan struct{}
	resume chan struct{}
}

// NewStreamController creates a new instance of StreamController.
// It initializes the state to streamRunning and creates channels for pausing and resuming the stream.
func NewStreamController() *StreamController {
	return &StreamController{
		state:  streamRunning,
		paused: make(chan struct{}, 1),
		resume: make(chan struct{}, 1),
	}
}

// SetState sets the state of the StreamController to the specified state.
// If the state is set to streamRunning or streamDiscard, it sends a signal to the resume channel.
func (s *StreamController) SetState(state int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = state

	if state == streamRunning || state == streamDiscard {
		s.resume <- struct{}{}
	}
}

// State returns the current state of the StreamController.
func (s *StreamController) State() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

// Paused indicates that the stream has been paused.
// This is different from the SetState(streamPaused) method as
// this will tell the stream controller to pause the stream
// (the stream may take some time to pause), but Paused()
// will indicate the actual time when the stream is paused.
func (s *StreamController) Paused() {
	s.paused <- struct{}{}
}

// WaitUntilPaused waits till the stream is paused
func (s *StreamController) WaitUntilPaused() {
	<-s.paused
}

type StreamReceiver interface {
	Recv() (*idl.HubReply, error)
}

func ParseStreamResponseFn(stream StreamReceiver, ctrl *StreamController) error {
	progressContainer := utils.NewProgressContainer(os.Stdout)
	respCh := make(chan *idl.HubReply)
	errCh := make(chan error)

	go func() {
		for {
			resp, err := stream.Recv()
			if err != nil {
				errCh <- err
				return
			}

			respCh <- resp
		}
	}()

loop:
	for {
		switch state := ctrl.State(); state {
		case streamPaused:
			progressContainer.Abort()
			ctrl.Paused()
			<-ctrl.resume

		default:
			select {
			case resp := <-respCh:
				if state == streamDiscard {
					continue
				}
				msg := resp.Message
				switch msg.(type) {
				case *idl.HubReply_LogMsg:
					logMsg := resp.GetLogMsg()
					switch logMsg.Level {
					case idl.LogLevel_DEBUG:
						gplog.Verbose(logMsg.Message)
					case idl.LogLevel_WARNING:
						gplog.Warn(logMsg.Message)
					case idl.LogLevel_ERROR:
						gplog.Error(logMsg.Message)
					case idl.LogLevel_FATAL:
						gplog.Fatal(nil, logMsg.Message)
					default:
						gplog.Info(logMsg.Message)
					}

				case *idl.HubReply_StdoutMsg:
					fmt.Print(resp.GetStdoutMsg())

				case *idl.HubReply_ProgressMsg:
					progressMsg := resp.GetProgressMsg()
					progressContainer.Update(progressMsg.Label, int(progressMsg.Current), int(progressMsg.Total))
				}

			case err := <-errCh:
				if err == io.EOF {
					break loop
				} else if err != nil {
					progressContainer.Abort()

					return utils.FormatGrpcError(err)
				}

			default:
				continue
			}
		}
	}

	return nil
}
