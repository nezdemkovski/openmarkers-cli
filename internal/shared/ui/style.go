package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Color palette (ANSI 256 — terminal-safe)
const (
	ColorPrimary   = lipgloss.Color("111") // Soft cyan — titles, active
	ColorSecondary = lipgloss.Color("250") // Light gray — headers, labels
	ColorNeutral   = lipgloss.Color("245") // Muted gray — metadata
	ColorAccent    = lipgloss.Color("109") // Soft green — success
	ColorWarning   = lipgloss.Color("180") // Soft amber — caution
	ColorError     = lipgloss.Color("167") // Soft red — failure
	ColorHighlight = lipgloss.Color("147") // Soft purple — selection
	ColorMuted     = lipgloss.Color("8")   // Dim gray — separators
)

// Typography roles
var (
	Title    = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	Subtitle = lipgloss.NewStyle().Foreground(ColorNeutral)
	Label    = lipgloss.NewStyle().Foreground(ColorSecondary)
	Value    = lipgloss.NewStyle()
	Meta     = lipgloss.NewStyle().Foreground(ColorNeutral)
)

// State roles
var (
	ActiveStyle  = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	SuccessStyle = lipgloss.NewStyle().Foreground(ColorAccent)
	WarningStyle = lipgloss.NewStyle().Foreground(ColorWarning)
	ErrorStyle   = lipgloss.NewStyle().Foreground(ColorError).Bold(true)
)

// Component styles
var (
	SpinnerStyle  = lipgloss.NewStyle().Foreground(ColorPrimary)
	SelectedStyle = lipgloss.NewStyle().Foreground(ColorHighlight).Bold(true)
	SepStyle      = lipgloss.NewStyle().Foreground(ColorMuted)
	HelpStyle     = lipgloss.NewStyle().Foreground(ColorNeutral).MarginTop(1)
	Container     = lipgloss.NewStyle().Padding(1, 2)
)

// Table styles
var (
	TableHeader = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	TableCell   = lipgloss.NewStyle()
	TableSep    = lipgloss.NewStyle().Foreground(ColorMuted)
)

// Status symbols (no emojis)
const (
	SymDone    = "•"
	SymActive  = "→"
	SymRunning = "◉"
	SymPending = "○"
	SymWarning = "!"
	SymError   = "✗"
	SymInfo    = "·"
)

// StatusError prints an error and exits.
func ExitWithError(msg string, err error) {
	fmt.Fprintln(os.Stderr, ErrorStyle.Render(fmt.Sprintf("%s %s: %v", SymError, msg, err)))
	os.Exit(1)
}

// StatusMessage prints a formatted status message to stderr.
func StatusMessage(symbol string, style lipgloss.Style, msg string) {
	fmt.Fprintln(os.Stderr, style.Render(fmt.Sprintf("%s %s", symbol, msg)))
}
