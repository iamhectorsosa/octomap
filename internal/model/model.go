package model

import (
	"fmt"
	"strings"
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
	go repository.ProcessRepo(
		m.config,
		m.ch,
		5*time.Millisecond,
	)
	return tea.Batch(
		m.spinner.Tick,
		m.updateProcess(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case processUpdateMsg:
		up := entity.Update{
			Description: msg.Description,
			Err:         msg.Err,
		}

		m.updates = append(m.updates, up)
		if len(m.updates) > 6 {
			m.updates = m.updates[1:]
		}
		return m, m.updateProcess()
	case processEndMsg:
		m.complete = true
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder
	s.WriteString("\n")

	if !m.complete {
		s.WriteString(m.spinner.View())
	} else {
		s.WriteString("  ")
	}

	s.WriteString("🐙 Mapping repository...\n\n")

	for _, res := range m.updates {
		mark := checkMark
		if res.Err != nil {
			mark = errorMark
		}
		s.WriteString(fmt.Sprintf("%s %s\n", mark, res.Description))
	}

	if !m.complete {
		s.WriteString(helpStyle("\nPress any key to exit"))
	} else {
		s.WriteString("\nProcess finished!\n\n")
	}

	return mainStyle.Render(s.String())
}

type (
	processEndMsg    struct{}
	processUpdateMsg entity.Update
)

func (m model) updateProcess() tea.Cmd {
	return func() tea.Msg {
		for update := range m.ch {
			return processUpdateMsg(update)
		}

		return processEndMsg{}
	}
}
