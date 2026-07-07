package ui

import (
	"fmt"
	"os"
)

var (
	// NoColor can be set to true to force disable ANSI color codes
	NoColor = false

	// ANSI Escape sequences
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

func init() {
	// Respect NO_COLOR environment variable (https://no-color.org/)
	if os.Getenv("NO_COLOR") != "" {
		NoColor = true
	}
}

// Green wraps string in green color ANSI codes if colors are enabled.
func Green(s string) string {
	if NoColor {
		return s
	}
	return colorGreen + s + colorReset
}

// Yellow wraps string in yellow color ANSI codes if colors are enabled.
func Yellow(s string) string {
	if NoColor {
		return s
	}
	return colorYellow + s + colorReset
}

// Red wraps string in red color ANSI codes if colors are enabled.
func Red(s string) string {
	if NoColor {
		return s
	}
	return colorRed + s + colorReset
}

// Blue wraps string in blue color ANSI codes if colors are enabled.
func Blue(s string) string {
	if NoColor {
		return s
	}
	return colorBlue + s + colorReset
}

// Gray wraps string in gray color ANSI codes if colors are enabled.
func Gray(s string) string {
	if NoColor {
		return s
	}
	return colorGray + s + colorReset
}

// Bold wraps string in bold ANSI codes if colors are enabled.
func Bold(s string) string {
	if NoColor {
		return s
	}
	return colorBold + s + colorReset
}

// Success returns a green checkmark symbol or "OK" depending on NO_COLOR.
func Success() string {
	if NoColor {
		return "[OK]"
	}
	return colorGreen + "✓" + colorReset
}

// Warning returns a yellow warning symbol or "WARN" depending on NO_COLOR.
func Warning() string {
	if NoColor {
		return "[WARN]"
	}
	return colorYellow + "⚠" + colorReset
}

// ErrorSymbol returns a red error symbol or "ERROR" depending on NO_COLOR.
func ErrorSymbol() string {
	if NoColor {
		return "[ERR]"
	}
	return colorRed + "✗" + colorReset
}

// ColorizeNodePrefix returns a formatted, styled prefix for parallel command streaming output.
func ColorizeNodePrefix(nodeName string, hash int) string {
	if NoColor {
		return fmt.Sprintf("[%s] ", nodeName)
	}
	// Select a color based on the hash of the node name to keep it consistent
	colors := []string{colorBlue, colorYellow, colorGreen, colorGray}
	color := colors[hash%len(colors)]
	return color + "[" + nodeName + "]" + colorReset + " "
}

// BrandHeader returns a typographical header for Orion with aligned version.
func BrandHeader(version string) string {
	title := "ORION"
	if !NoColor {
		title = colorBold + colorBlue + "ORION" + colorReset
	}
	// Fixed padding length to align version on the right
	return fmt.Sprintf(" %s                                         %s", title, Gray(version))
}
