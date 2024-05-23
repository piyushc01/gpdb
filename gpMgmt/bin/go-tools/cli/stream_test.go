package cli_test

import (
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpdb/gp/cli"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/testutils"
)

const (
	streamRunning = iota
	streamPaused
	streamDiscard
)

type msgStream struct {
	msg   []*idl.HubReply
	err   error
	sent  int
	delay int
}

func (m *msgStream) Recv() (*idl.HubReply, error) {
	if len(m.msg) == 0 {
		if m.err == nil {
			return nil, io.EOF
		}

		return nil, m.err
	}

	nextMsg := (m.msg)[0]
	m.msg = (m.msg)[1:]
	m.sent++

	if m.delay != 0 {
		time.Sleep(time.Duration(m.delay) * time.Millisecond)
	}

	return nextMsg, nil
}

func (m *msgStream) Sent() int {
	return m.sent
}

func TestParseStreamResponse(t *testing.T) {
	t.Run("displays the correct stream responses to the user", func(t *testing.T) {
		_, _, logfile := testhelper.SetupTestLogger()

		infoLogMsg := "info log message"
		warnLogMsg := "warning log message"
		errLogMsg := "error log message"
		dbgLogMsg := "debug log message"
		msg := []*idl.HubReply{
			{
				Message: &idl.HubReply_StdoutMsg{
					StdoutMsg: "stdout message",
				},
			},
			{
				Message: &idl.HubReply_LogMsg{
					LogMsg: &idl.LogMessage{Message: infoLogMsg, Level: idl.LogLevel_INFO},
				},
			},
			{
				Message: &idl.HubReply_LogMsg{
					LogMsg: &idl.LogMessage{Message: warnLogMsg, Level: idl.LogLevel_WARNING},
				},
			},
			{
				Message: &idl.HubReply_LogMsg{
					LogMsg: &idl.LogMessage{Message: errLogMsg, Level: idl.LogLevel_ERROR},
				},
			},
			{
				Message: &idl.HubReply_LogMsg{
					LogMsg: &idl.LogMessage{Message: dbgLogMsg, Level: idl.LogLevel_DEBUG},
				},
			},
			{
				Message: &idl.HubReply_ProgressMsg{
					ProgressMsg: &idl.ProgressMessage{
						Label:   "progress message",
						Current: 0,
						Total:   1,
					},
				},
			},
			{
				Message: &idl.HubReply_ProgressMsg{
					ProgressMsg: &idl.ProgressMessage{
						Label:   "progress message",
						Current: 1,
						Total:   1,
					},
				},
			},
		}

		buffer, writer, resetStdout := testutils.CaptureStdout(t)
		defer resetStdout()

		err := cli.ParseStreamResponse(&msgStream{msg: msg}, cli.NewStreamController())
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
		writer.Close()

		testutils.AssertLogMessage(t, logfile, infoLogMsg)
		testutils.AssertLogMessage(t, logfile, warnLogMsg)
		testutils.AssertLogMessage(t, logfile, errLogMsg)
		testutils.AssertLogMessage(t, logfile, dbgLogMsg)

		stdout := <-buffer
		expectedStdoutMsg := "stdout message"
		if !strings.Contains(stdout, expectedStdoutMsg) {
			t.Fatalf("got %v, want %v", stdout, expectedStdoutMsg)
		}

		expectedProgressContents := []string{"progress message", "done", "1/1"}
		for _, expected := range expectedProgressContents {
			if !strings.Contains(stdout, expected) {
				t.Fatalf("got %v, want %v", stdout, expected)
			}
		}
	})

	t.Run("returns non EOF errors and aborts any running progress bars", func(t *testing.T) {
		msg := []*idl.HubReply{
			{
				Message: &idl.HubReply_ProgressMsg{
					ProgressMsg: &idl.ProgressMessage{
						Label:   "progress message",
						Current: 0,
						Total:   5,
					},
				},
			},
			{
				Message: &idl.HubReply_ProgressMsg{
					ProgressMsg: &idl.ProgressMessage{
						Label:   "progress message",
						Current: 1,
						Total:   5,
					},
				},
			},
		}

		buffer, writer, resetStdout := testutils.CaptureStdout(t)
		defer resetStdout()

		expectedErr := errors.New("error")
		err := cli.ParseStreamResponse(&msgStream{
			msg: msg,
			err: expectedErr,
		}, cli.NewStreamController())
		if !errors.Is(err, expectedErr) {
			t.Fatalf("got %#v, want %#v", err, expectedErr)
		}

		writer.Close()
		stdout := <-buffer

		expectedProgressContents := []string{"progress message", "error", "1/5"}
		for _, expected := range expectedProgressContents {
			if !strings.Contains(stdout, expected) {
				t.Fatalf("got %v, want %v", stdout, expected)
			}
		}
	})

	t.Run("controller should be able to pause and resume the stream", func(t *testing.T) {
		_, _, logfile := testhelper.SetupTestLogger()

		msg := []*idl.HubReply{}
		for i := 0; i < 5; i++ {
			msg = append(msg, &idl.HubReply{
				Message: &idl.HubReply_LogMsg{
					LogMsg: &idl.LogMessage{Message: "log message", Level: idl.LogLevel_INFO},
				},
			})
		}

		var wg sync.WaitGroup
		errCh := make(chan error, 1)
		ctrl := cli.NewStreamController()
		streamer := &msgStream{
			msg:   msg,
			delay: 100,
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cli.ParseStreamResponse(streamer, ctrl)

			if err != nil {
				errCh <- err
			}
		}()

		for {
			if streamer.Sent() == 3 {
				ctrl.SetState(streamPaused)
				ctrl.Paused()
				break
			}
		}

		// we assert one less than the time the stream was paused because
		// the response will be recieved but not processed
		testutils.AssertLogMessageCount(t, logfile, "log message", 2)

		ctrl.SetState(streamRunning)

		wg.Wait()
		close(errCh)

		err := <-errCh
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// check if we have processed all the responses
		testutils.AssertLogMessageCount(t, logfile, "log message", 5)
	})

	t.Run("controller should be able to discard the stream", func(t *testing.T) {
		_, _, logfile := testhelper.SetupTestLogger()

		msg := []*idl.HubReply{}
		for i := 0; i < 5; i++ {
			msg = append(msg, &idl.HubReply{
				Message: &idl.HubReply_LogMsg{
					LogMsg: &idl.LogMessage{Message: "log message", Level: idl.LogLevel_INFO},
				},
			})
		}

		var wg sync.WaitGroup
		errCh := make(chan error, 1)
		ctrl := cli.NewStreamController()
		streamer := &msgStream{
			msg:   msg,
			delay: 100,
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cli.ParseStreamResponse(streamer, ctrl)

			if err != nil {
				errCh <- err
			}
		}()

		for {
			if streamer.Sent() == 3 {
				ctrl.SetState(streamPaused)
				ctrl.Paused()
				break
			}
		}

		// we assert one less than the time the stream was paused because
		// the response will be recieved but not processed
		testutils.AssertLogMessageCount(t, logfile, "log message", 2)

		ctrl.SetState(streamDiscard)

		wg.Wait()
		close(errCh)

		err := <-errCh
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// check that we do not have any additional responses
		testutils.AssertLogMessageCount(t, logfile, "log message", 2)
	})
}
