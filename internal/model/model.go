package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iamhectorsosa/octomap/pkg/processor"
)

var (
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	mainStyle = lipgloss.NewStyle().MarginLeft(1)
	checkMark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	errorMark = lipgloss.NewStyle().Foreground(lipgloss.Color("160")).SetString("x")
)

type model struct {
	err       error
	config    *processor.Config
	updatesCh chan processor.Update
	updates   []processor.Update
	spinner   spinner.Model
	complete  bool
}

func New(config *processor.Config) model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		config:    config,
		spinner:   sp,
		updates:   []processor.Update{},
		updatesCh: make(chan processor.Update),
	}
}

func (m model) Init() tea.Cmd {
	processor := processor.New(m.config, m.updatesCh)
	go processor.Process(time.Millisecond)
	return tea.Batch(m.spinner.Tick, m.updateProcess())
}

type (
	errMsg    struct{ err error }
	updateMsg processor.Update
	endMsg    struct{}
)

func (e errMsg) Error() string { return e.err.Error() }

func (m model) updateProcess() tea.Cmd {
	return func() tea.Msg {
		update, ok := <-m.updatesCh
		if !ok {
			return endMsg{}
		}
		if update.Err != nil {
			return errMsg{update.Err}
		}
		return updateMsg(update)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case errMsg:
		m.err = msg
		return m, tea.Quit
	case updateMsg:
		m.updates = append(m.updates, processor.Update(msg))
		if len(m.updates) > 6 {
			m.updates = m.updates[1:]
		}
		return m, m.updateProcess()
	case endMsg:
		m.complete = true
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder
	s.WriteString("\n")

	if m.complete || m.err != nil {
		s.WriteString("  ")
	} else {
		s.WriteString(m.spinner.View())
	}

	s.WriteString("🐙 Mapping repository...\n\n")
	for _, res := range m.updates {
		s.WriteString(fmt.Sprintf("%s %s\n", checkMark, res.Description))
	}

	if m.err != nil {
		s.WriteString(fmt.Sprintf("%s %s\n", errorMark, m.err.Error()))
	}

	if m.complete || m.err != nil {
		s.WriteString("\nProcess finished!\n\n")
	} else {
		s.WriteString(helpStyle("\nPress any key to exit"))
	}

	return mainStyle.Render(s.String())
}
