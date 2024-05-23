package utils_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/greenplum-db/gpdb/gp/utils"
)

func TestProgressBar(t *testing.T) {
	t.Run("creates progress bar with the correct label and total count", func(t *testing.T) {
		var buf bytes.Buffer
		instance := utils.NewProgressInstance(&buf)
		bar := utils.NewProgressBar(instance, "Test", 10)

		for i := 0; i < 10; i++ {
			bar.Increment()
		}
		instance.Wait()

		expectedLabel := "Test"
		if !bytes.Contains(buf.Bytes(), []byte(expectedLabel)) {
			t.Fatalf("expected string %q not present in progress bar", expectedLabel)
		}

		expectedTotal := "/10"
		if !bytes.Contains(buf.Bytes(), []byte(expectedTotal)) {
			t.Fatalf("expected string %q not present in progress bar", expectedTotal)
		}
	})

	t.Run("has the correct bar style", func(t *testing.T) {
		var buf bytes.Buffer
		instance := utils.NewProgressInstance(&buf)
		bar := utils.NewProgressBar(instance, "Test", 5)

		for i := 0; i < 5; i++ {
			bar.Increment()
		}
		instance.Wait()

		expectedProgress := "==="
		expectedClose := "["
		expectedStart := "]"

		if !bytes.Contains(buf.Bytes(), []byte(expectedProgress)) {
			t.Fatalf("expected string %q not present in progress bar", expectedProgress)
		}
		if !bytes.Contains(buf.Bytes(), []byte(expectedClose)) {
			t.Fatalf("expected string %q not present in progress bar", expectedProgress)
		}
		if !bytes.Contains(buf.Bytes(), []byte(expectedStart)) {
			t.Fatalf("expected string %q not present in progress bar", expectedProgress)
		}
	})

	t.Run("appends done once the bar is completed", func(t *testing.T) {
		var buf bytes.Buffer
		instance := utils.NewProgressInstance(&buf)
		bar := utils.NewProgressBar(instance, "Test", 10)

		for i := 0; i < 10; i++ {
			bar.Increment()
		}
		instance.Wait()

		expected := "\033[32mdone\033[0m"
		if !bytes.Contains(buf.Bytes(), []byte(expected)) {
			t.Fatalf("expected string %q not present in progress bar", expected)
		}
	})

	t.Run("appends error if the bar is aborted", func(t *testing.T) {
		var buf bytes.Buffer
		instance := utils.NewProgressInstance(&buf)
		bar := utils.NewProgressBar(instance, "Test", 10)

		for i := 0; i < 10; i++ {
			bar.Increment()
			if i == 5 {
				bar.Abort(false)
			}
		}
		instance.Wait()

		expected := "\033[31merror\033[0m"
		if !bytes.Contains(buf.Bytes(), []byte(expected)) {
			t.Fatalf("expected string %q not present in progress bar", expected)
		}
	})
}

func TestProgressBarContainer(t *testing.T) {
	t.Run("updates the progress when there is single bar", func(t *testing.T) {
		var buf bytes.Buffer
		container := utils.NewProgressContainer(&buf)

		expectedLabel := "Bar 1"
		for i := 0; i < 2; i++ {
			container.Update(expectedLabel, i, 3)
			time.Sleep(200 * time.Millisecond)

			expectedTotal := fmt.Sprintf("%d/3", i)
			if bytes.Count(buf.Bytes(), []byte(expectedTotal)) != 1 {
				t.Fatalf("expected string %q not present in progress bar", expectedTotal)
			}
		}

		if !bytes.Contains(buf.Bytes(), []byte(expectedLabel)) {
			t.Fatalf("expected string %q not present in progress bar", expectedLabel)
		}
	})

	t.Run("updates the progress when there are multiple bars", func(t *testing.T) {
		var buf bytes.Buffer
		container := utils.NewProgressContainer(&buf)

		expectedLabel1 := "Bar 1"
		expectedLabel2 := "Bar 2"
		for i := 0; i < 2; i++ {
			container.Update(expectedLabel1, i, 3)
			container.Update(expectedLabel2, i, 3)
			time.Sleep(200 * time.Millisecond)

			expectedTotal := fmt.Sprintf("%d/3", i)
			if bytes.Count(buf.Bytes(), []byte(expectedTotal)) != 2 {
				t.Fatalf("expected string %q not present in progress bar", expectedTotal)
			}
		}

		if !bytes.Contains(buf.Bytes(), []byte(expectedLabel1)) {
			t.Fatalf("expected string %q not present in progress bar", expectedLabel1)
		}
		if !bytes.Contains(buf.Bytes(), []byte(expectedLabel2)) {
			t.Fatalf("expected string %q not present in progress bar", expectedLabel2)
		}
	})

	t.Run("refreshes the container when all the bars are complete", func(t *testing.T) {
		var buf bytes.Buffer
		container := utils.NewProgressContainer(&buf)

		label1 := "Bar 1"
		label2 := "Bar 2"
		container.Update(label1, 0, 3)
		container.Update(label2, 0, 3)

		count := len(container.GetBars())
		if count != 2 {
			t.Fatalf("got %d, want 2 bars to be present in the container", count)
		}

		container.Update(label1, 3, 3)
		container.Update(label2, 3, 3)

		expected := "done"
		if !bytes.Contains(buf.Bytes(), []byte(expected)) {
			t.Fatalf("expected string %q not present in progress bar", expected)
		}

		count = len(container.GetBars())
		if count != 0 {
			t.Fatalf("got %d, want no bars to be present in the container", count)
		}
	})

	t.Run("reports if there are any running bars", func(t *testing.T) {
		var buf bytes.Buffer
		container := utils.NewProgressContainer(&buf)

		label := "Bar 1"
		container.Update(label, 0, 3)
		if !container.IsRunning() {
			t.Fatalf("expected progress bars to be running")
		}

		container.Update(label, 3, 3)
		if container.IsRunning() {
			t.Fatalf("expected progress bars to not be running")
		}
	})

	t.Run("aborts and refreshes the container", func(t *testing.T) {
		var buf bytes.Buffer
		container := utils.NewProgressContainer(&buf)

		// abort on empty container does nothing
		container.Abort()

		label1 := "Bar 1"
		label2 := "Bar 2"
		container.Update(label1, 0, 3)
		container.Update(label2, 0, 3)

		count := len(container.GetBars())
		if count != 2 {
			t.Fatalf("got %d, want 2 bars to be present in the container", count)
		}

		if !container.IsRunning() {
			t.Fatalf("expected progress bars to be running")
		}

		container.Abort()

		expected := "error"
		if !bytes.Contains(buf.Bytes(), []byte(expected)) {
			t.Fatalf("expected string %q not present in progress bar", expected)
		}

		if container.IsRunning() {
			t.Fatalf("expected progress bars to not be running")
		}

		count = len(container.GetBars())
		if count != 0 {
			t.Fatalf("got %d, want no bars to be present in the container", count)
		}
	})

	t.Run("write headers on top of the progress bars", func(t *testing.T) {
		var buf bytes.Buffer
		container := utils.NewProgressContainer(&buf)

		label := "Bar 1"
		container.Update(label, 0, 3)

		expected := "foobar"
		container.WriteHeader("foobar")
		container.Update(label, 3, 3)

		if !bytes.Contains(buf.Bytes(), []byte(expected)) {
			t.Fatalf("expected string %q not present in progress bar", expected)
		}
	})
}
