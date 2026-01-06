package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

// Spinner wraps progressbar with timing functionality
type Spinner struct {
	bar       *progressbar.ProgressBar
	startTime time.Time
	desc      string
	done      chan struct{}
}

// NewSpinner creates a new spinner with description
func NewSpinner(description string) *Spinner {
	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(10),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetPredictTime(false),
	)

	s := &Spinner{
		bar:       bar,
		startTime: time.Now(),
		desc:      description,
		done:      make(chan struct{}),
	}

	// Start spinner animation
	go func() {
		for {
			select {
			case <-s.done:
				return
			default:
				s.bar.Add(1)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	return s
}

// Stop stops the spinner and prints elapsed time
func (s *Spinner) Stop() time.Duration {
	close(s.done)
	s.bar.Finish()

	elapsed := time.Since(s.startTime)
	fmt.Fprintf(os.Stderr, "%s done (%s)\n", s.desc, formatDuration(elapsed))

	return elapsed
}

// StopWithError stops the spinner without success message
func (s *Spinner) StopWithError() time.Duration {
	close(s.done)
	s.bar.Finish()
	return time.Since(s.startTime)
}

// formatDuration formats duration in human-readable form
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)

	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}

	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60

	if seconds == 0 {
		return fmt.Sprintf("%dm", minutes)
	}

	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

// FormatDuration exports duration formatting for use in main
func FormatDuration(d time.Duration) string {
	return formatDuration(d)
}
