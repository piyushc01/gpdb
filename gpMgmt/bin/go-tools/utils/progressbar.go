package utils

import (
	"fmt"
	"io"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

const ( // ASCII color codes
	red   = "\033[31m"
	green = "\033[32m"
	reset = "\033[0m"
)

func NewProgressInstance(output io.Writer) *mpb.Progress {
	return mpb.New(
		mpb.WithWidth(64),
		mpb.PopCompletedMode(),
		mpb.WithOutput(output),
		mpb.WithAutoRefresh(),
	)
}

func NewProgressBar(instance *mpb.Progress, label string, size int) *mpb.Bar {
	bar := instance.AddBar(int64(size),
		mpb.PrependDecorators(
			decor.Name(label, decor.WC{W: len(label) + 1, C: decor.DidentRight}),
			decor.CountersNoUnit("%d/%d"),
			decor.Elapsed(decor.ET_STYLE_GO, decor.WC{W: 6}),
		),
		mpb.AppendDecorators(
			decor.OnAbort(
				decor.OnComplete(
					decor.Percentage(decor.WC{W: 4}), fmt.Sprintf("%sdone%s", green, reset),
				),
				fmt.Sprintf("%serror%s", red, reset),
			),
		),
	)

	return bar
}

// progressContainer represents a container for managing progress bars
type progressContainer struct {
	output   io.Writer
	instance *mpb.Progress
	barMap   map[string]*mpb.Bar
}

// NewProgressContainer creates a new progressContainer object with
// the given output writer
func NewProgressContainer(output io.Writer) *progressContainer {
	return &progressContainer{
		output:   output,
		instance: NewProgressInstance(output),
		barMap:   make(map[string]*mpb.Bar),
	}
}

// refresh resets the progress container.
// This should be called once the progress container is completed or aborted
func (p *progressContainer) refresh() {
	p.instance = NewProgressInstance(p.output)
	p.barMap = make(map[string]*mpb.Bar)
}

// GetBars returns a list of all the running progress bar objects in the container
func (p *progressContainer) GetBars() []*mpb.Bar {
	result := []*mpb.Bar{}
	for _, bar := range p.barMap {
		result = append(result, bar)
	}

	return result
}

// Update updates the progress of a specific label with the current value.
// If the progress bar for the given label does not exist, it creates a new one.
// It checks if all progress bars are completed, and if so, waits for the instance to complete and refreshes the progress bars.
func (p *progressContainer) Update(label string, current, total int) {
	if _, ok := p.barMap[label]; !ok {
		p.barMap[label] = NewProgressBar(p.instance, label, total)
	}

	p.barMap[label].SetCurrent(int64(current))

	for _, bar := range p.barMap {
		if !bar.Completed() {
			return
		}
	}

	p.instance.Wait()
	p.refresh()
}

// Abort cancels all the progress bars and waits for them to finish.
// It also waits for the instance to finish and refreshes the progress container.
func (p *progressContainer) Abort() {
	if len(p.barMap) == 0 {
		return
	}

	for _, bar := range p.barMap {
		bar.Abort(true)
		bar.Wait()
	}
	p.instance.Wait()
	p.refresh()
}

// IsRunning checks if any progress bar in the progress container is currently running.
func (p *progressContainer) IsRunning() bool {
	for _, bar := range p.barMap {
		if bar.IsRunning() {
			return true
		}
	}

	return false
}

// WriteHeader writes the header text to the progress container.
// It takes a string parameter `text` representing the header text.
func (p *progressContainer) WriteHeader(text string) {
	p.instance.Write([]byte(text)) // nolint
}
