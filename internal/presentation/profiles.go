package presentation

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/openmarkers/openmarkers-cli/internal/infrastructure/api"
	"github.com/openmarkers/openmarkers-cli/internal/shared/models"
	"github.com/openmarkers/openmarkers-cli/internal/shared/ui"
)

type ProfilesView struct {
	client   *api.Client
	profiles []ProfileItem
	cursor   int
	loading  bool
	spinner  spinner.Model
	err      error
	width    int
	height   int
}

func NewProfilesView(client *api.Client) *ProfilesView {
	return &ProfilesView{
		client:  client,
		loading: true,
		spinner: NewSpinner(),
	}
}

func (v *ProfilesView) Init() tea.Cmd {
	return tea.Batch(
		v.spinner.Tick,
		v.fetchProfiles(),
	)
}

func (v *ProfilesView) fetchProfiles() tea.Cmd {
	return func() tea.Msg {
		var profiles []models.Profile
		if err := v.client.Get(context.Background(), "/api/profiles", &profiles); err != nil {
			return ErrMsg{Err: err}
		}
		items := make([]ProfileItem, len(profiles))
		for i, p := range profiles {
			items[i] = ProfileItem{
				ID:   p.ID,
				Name: p.Name,
				DOB:  p.DateOfBirth,
				Sex:  p.Sex,
			}
		}
		return ProfilesMsg{Profiles: items}
	}
}

func (v *ProfilesView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return v, tea.Quit
		case "up", "k":
			if v.cursor > 0 {
				v.cursor--
			}
		case "down", "j":
			if v.cursor < len(v.profiles)-1 {
				v.cursor++
			}
		}

	case ProfilesMsg:
		v.loading = false
		v.profiles = msg.Profiles
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

func (v *ProfilesView) View() string {
	s := ui.Title.Render("OpenMarkers — Profiles") + "\n\n"

	if v.loading {
		s += v.spinner.View() + " Loading profiles..."
		return ui.Container.Render(s)
	}

	if v.err != nil {
		s += ui.ErrorStyle.Render("Error: "+v.err.Error()) + "\n"
		return ui.Container.Render(s)
	}

	if len(v.profiles) == 0 {
		s += "No profiles found. Create one with: openmarkers profile create --name \"Name\" --sex M\n"
		return ui.Container.Render(s)
	}

	for i, p := range v.profiles {
		cursor := "  "
		style := lipgloss.NewStyle()
		if i == v.cursor {
			cursor = fmt.Sprintf("%s ", ui.SymActive)
			style = ui.SelectedStyle
		}
		line := fmt.Sprintf("%s%d  %-20s  %s  %s", cursor, p.ID, p.Name, p.DOB, p.Sex)
		s += style.Render(line) + "\n"
	}

	s += ui.HelpStyle.Render("\n↑/↓ navigate • q quit")
	return ui.Container.Render(s)
}
