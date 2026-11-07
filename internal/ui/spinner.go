package ui

import (
	"fmt"
	"io"
	"time"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const spinnerInterval = 80 * time.Millisecond

func WithSpinner[T any](out io.Writer, enabled bool, title string, work func() T) T {
	if !enabled {
		return work()
	}

	done := make(chan T, 1)
	go func() { done <- work() }()

	ticker := time.NewTicker(spinnerInterval)
	defer ticker.Stop()

	for frame := 0; ; frame++ {
		select {
		case result := <-done:
			fmt.Fprint(out, "\r\033[K")
			return result
		case <-ticker.C:
			fmt.Fprintf(out, "\r%s %s", Title.Render(spinnerFrames[frame%len(spinnerFrames)]), title)
		}
	}
}
