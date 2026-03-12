package presentation

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/openmarkers/openmarkers-cli/internal/shared/ui"
)

func NewSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = ui.SpinnerStyle
	return s
}
