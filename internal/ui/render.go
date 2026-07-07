package ui

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// DeviceState represents the execution status of a single node in the network.
type DeviceState struct {
	Name   string
	Status string // "Waiting", "Running", "Success", "Failed", "Warning"
	Detail string // Additional details like timing or warning reasons
}

// RenderDeviceStates updates the terminal output for target nodes in-place using ANSI escapes.
func RenderDeviceStates(w io.Writer, states []DeviceState, isFirst bool) {
	if NoColor {
		// Non-ANSI fallback: just print line-by-line without terminal rewrites
		if isFirst {
			for _, s := range states {
				fmt.Fprintf(w, "  %-12s %s %s\n", s.Name, s.Status, s.Detail)
			}
		}
		return
	}

	if !isFirst {
		// Move cursor up by the number of states to overwrite in-place
		fmt.Fprintf(w, "\033[%dA", len(states))
	}

	for _, s := range states {
		var indicator string
		switch s.Status {
		case "Success":
			indicator = Success()
		case "Failed":
			indicator = ErrorSymbol()
		case "Warning":
			indicator = Warning()
		case "Running":
			indicator = Blue("●")
		default: // Waiting
			indicator = Gray("○")
		}
		// Clear line and print state details
		fmt.Fprintf(w, "\033[2K  %-12s %s %s\n", s.Name, indicator, s.Detail)
	}
}

// Header renders a standardized section header with 1 leading space and bold typography.
func Header(w io.Writer, title string) {
	fmt.Fprintf(w, " %s\n", Bold(title))
}

// PrintDetail prints a standard indented key-value detail pair with 3 leading spaces.
func PrintDetail(w io.Writer, label, value string) {
	fmt.Fprintf(w, "   %-11s %s\n", label, value)
}

// Separator prints a visual divider.
func Separator(w io.Writer) {
	if NoColor {
		fmt.Fprintln(w, "──────────────────────────────────────────────────────────")
	} else {
		fmt.Fprintln(w, colorGray+"──────────────────────────────────────────────────────────"+colorReset)
	}
}

// Suggestion renders a next-action suggestion block.
func Suggestion(w io.Writer, command string) {
	fmt.Fprintf(w, "   %s\n", Blue(command))
}

// PrintError formats a premium, structured error message matching the design system.
func PrintError(w io.Writer, summary, reason, tryCmd string) {
	fmt.Fprintf(w, " %s %s\n", ErrorSymbol(), Bold(summary))
	if reason != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  "+Bold(Gray("Reason")))
		fmt.Fprintf(w, "    %s\n", reason)
	}
	if tryCmd != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  "+Bold(Gray("Try")))
		fmt.Fprintf(w, "    %s\n", Blue(tryCmd))
	}
}

// PrintTable formats and prints a grid of data with consistent column widths.
func PrintTable(w io.Writer, headers []string, rows [][]string) {
	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, val := range row {
			if i < len(colWidths) && len(val) > colWidths[i] {
				colWidths[i] = len(val)
			}
		}
	}

	// Print headers
	for i, h := range headers {
		padding := colWidths[i] - len(h)
		fmt.Fprint(w, Bold(Gray(h))+strings.Repeat(" ", padding+4))
	}
	fmt.Fprintln(w)

	// Print rows
	for _, row := range rows {
		for i, val := range row {
			if i < len(colWidths) {
				padding := colWidths[i] - len(val)
				fmt.Fprint(w, val+strings.Repeat(" ", padding+4))
			}
		}
		fmt.Fprintln(w)
	}
}

// SimulateSpinner simulates a brief loading action with elapsed time.
func SimulateSpinner(w io.Writer, message string, duration time.Duration) {
	if duration <= 0 {
		return
	}
	
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	if NoColor {
		// Non-ANSI: print a static line and return
		fmt.Fprintf(w, "  %s...\n", message)
		time.Sleep(duration)
		return
	}

	steps := int(duration / (100 * time.Millisecond))
	startTime := time.Now()

	for i := 0; i < steps; i++ {
		frame := frames[i%len(frames)]
		elapsed := time.Since(startTime).Truncate(100 * time.Millisecond)
		fmt.Fprintf(w, "\r  %s %s (%s)", Blue(frame), message, elapsed)
		time.Sleep(100 * time.Millisecond)
	}
	// Clear the spinner line
	fmt.Fprintf(w, "\r\033[2K")
}
