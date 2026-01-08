package utils

import (
	"github.com/schollz/progressbar/v3"
)

// NewProgressBar creates a new progress bar for tracking byte-based operations.
// The max parameter specifies the total number of bytes, and description
// provides a label for the progress bar. Returns a progress bar configured
// for displaying byte-sized progress (e.g., "10 MB / 100 MB").
func NewProgressBar(max int64, description string) *progressbar.ProgressBar {
	return progressbar.DefaultBytes(max, description)
}

// NewSpinner creates a spinner for indeterminate operations where progress
// cannot be measured. The description parameter provides a label displayed
// next to the spinner. Returns a progress bar configured in spinner mode
// that continuously animates until stopped. Use this for operations like
// SSH connections or waiting for remote commands where the duration is unknown.
func NewSpinner(description string) *progressbar.ProgressBar {
	return progressbar.NewOptions(-1,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSpinnerType(14),
	)
}
