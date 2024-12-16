package model

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iamhectorsosa/octomap/internal/entity"
	"github.com/iamhectorsosa/octomap/internal/repository"
)

var (
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	mainStyle = lipgloss.NewStyle().MarginLeft(1)
	checkMark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	errorMark = lipgloss.NewStyle().Foreground(lipgloss.Color("160")).SetString("x")
)

type model struct {
	config   *entity.Config
	ch       chan entity.Update
	updates  []entity.Update
	spinner  spinner.Model
	complete bool
	quitting bool
}

func New(cfg *entity.Config) model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		config:  cfg,
		spinner: sp,
		updates: []entity.Update{},
		ch:      make(chan entity.Update),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.runPretendProcess(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.quitting = true
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case processFinishedMsg:
		up := entity.Update{
			Description: msg.Description,
			Err:         msg.Err,
		}

		m.updates = append(m.updates, up)
		if len(m.updates) > 6 {
			m.updates = m.updates[1:]
		}
		return m, m.runPretendProcess()
	case processEndMsg:
		m.complete = true
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	s := "\n"
	if !m.complete {
		s += m.spinner.View()
	} else {
		s += "  "
	}
	s += "🐙 Mapping repository...\n\n"

	for _, res := range m.updates {
		mark := checkMark
		if res.Err != nil {
			mark = errorMark
		}
		s += fmt.Sprintf("%s %s\n", mark, res.Description)
	}

	if !m.complete {
		s += helpStyle("\nPress any key to exit")
	} else {
		s += "\nProcess finished!\n\n"
	}

	return mainStyle.Render(s)
}

type (
	processEndMsg      struct{}
	processFinishedMsg entity.Update
)

func (m model) runPretendProcess() tea.Cmd {
	return func() tea.Msg {
		if len(m.updates) == 0 {
			go repository.ProcessRepo(
				m.config.Repo,
				m.config.Url,
				m.config.Dir,
				m.config.Output,
				m.config.Include,
				m.config.Exclude,
				m.ch,
				25*time.Millisecond,
			)
		}

		msg, ok := <-m.ch
		if !ok {
			return processEndMsg{}
		}

		return processFinishedMsg{
			Description: msg.Description,
			Err:         msg.Err,
		}
	}
}
