package presentation

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/openmarkers/openmarkers-cli/internal/infrastructure/api"
	"github.com/openmarkers/openmarkers-cli/internal/shared/models"
	"github.com/openmarkers/openmarkers-cli/internal/shared/ui"
)

type ResultsView struct {
	client    *api.Client
	profileID string
	results   []ResultItem
	cursor    int
	loading   bool
	spinner   spinner.Model
	err       error
	width     int
	height    int
}

func NewResultsView(client *api.Client, profileID string) *ResultsView {
	return &ResultsView{
		client:    client,
		profileID: profileID,
		loading:   true,
		spinner:   NewSpinner(),
	}
}

func (v *ResultsView) Init() tea.Cmd {
	return tea.Batch(
		v.spinner.Tick,
		v.fetchResults(),
	)
}

func (v *ResultsView) fetchResults() tea.Cmd {
	return func() tea.Msg {
		var results []models.Result
		if err := v.client.Get(context.Background(), "/api/results?profile_id="+v.profileID, &results); err != nil {
			return ErrMsg{Err: err}
		}
		items := make([]ResultItem, len(results))
		for i, r := range results {
			items[i] = ResultItem{
				ID:          r.ID,
				BiomarkerID: r.BiomarkerID,
				Date:        r.Date,
				Value:       r.Value,
			}
		}
		return ResultsMsg{Results: items}
	}
}

func (v *ResultsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return v, tea.Quit
		case "up", "k":
			if v.cursor > 0 {
				v.cursor--
			}
		case "down", "j":
			if v.cursor < len(v.results)-1 {
				v.cursor++
			}
		}

	case ResultsMsg:
		v.loading = false
		v.results = msg.Results
		return v, nil

	case ErrMsg:
		v.loading = false
		v.err = msg.Err
		return v, nil

	case spinner.TickMsg:
		if v.loading {
			var cmd tea.Cmd
			v.spinner, cmd = v.spinner.Update(msg)
			return v, cmd
		}
	}

	return v, nil
}

func (v *ResultsView) View() string {
	s := ui.Title.Render(fmt.Sprintf("OpenMarkers — Results (Profile %s)", v.profileID)) + "\n\n"

	if v.loading {
		s += v.spinner.View() + " Loading results..."
		return ui.Container.Render(s)
	}

	if v.err != nil {
		s += ui.ErrorStyle.Render("Error: "+v.err.Error()) + "\n"
		return ui.Container.Render(s)
	}

	if len(v.results) == 0 {
		s += "No results found.\n"
		return ui.Container.Render(s)
	}

	header := fmt.Sprintf("  %-6s  %-25s  %-12s  %s", "ID", "Biomarker", "Date", "Value")
	s += ui.TableHeader.Render(header) + "\n"
	s += ui.TableSep.Render("  ──────  ─────────────────────────  ────────────  ──────────") + "\n"

	for i, r := range v.results {
		cursor := "  "
		style := ui.Value
		if i == v.cursor {
			cursor = fmt.Sprintf("%s ", ui.SymActive)
			style = ui.SelectedStyle
		}
		line := fmt.Sprintf("%s%-6d  %-25s  %-12s  %s", cursor, r.ID, r.BiomarkerID, r.Date, r.Value)
		s += style.Render(line) + "\n"
	}

	s += ui.HelpStyle.Render("\n↑/↓ navigate • esc back • q quit")
	return ui.Container.Render(s)
}
