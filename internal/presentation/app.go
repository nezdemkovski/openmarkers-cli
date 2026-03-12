package presentation

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/openmarkers/openmarkers-cli/internal/infrastructure/api"
)

type App struct {
	client  *api.Client
	current tea.Model
}

func NewApp(client *api.Client) *App {
	return &App{
		client:  client,
		current: NewProfilesView(client),
	}
}

func (a *App) Init() tea.Cmd {
	return a.current.Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.current, cmd = a.current.Update(msg)
	return a, cmd
}

func (a *App) View() string {
	return a.current.View()
}

func Run(client *api.Client) error {
	p := tea.NewProgram(NewApp(client), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
