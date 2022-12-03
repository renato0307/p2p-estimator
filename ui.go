package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	participants map[string]participant
	cr           *ChatRoom

	menu   list.Model
	choice string
	table  table.Model

	description     textinput.Model
	editDescription bool
	showDescription bool

	showVotes bool
}

type tickMsg time.Time
type receiveMsg *ChatMessage

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		m.receiveMsgCmd(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.editDescription {
				m.updateDescription(true)
			} else {
				m.handleMenuEvents()
			}
			return m, nil
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case receiveMsg:
		m.handleNewMessage(msg)
		return m, m.receiveMsgCmd()
	case tickMsg:
		m.sendHeartbeat()
		cmd = m.updateParticipantsTable(msg)
		return m, tea.Batch(tickCmd(), cmd)
	}

	if m.editDescription {
		m.description, cmd = m.description.Update(msg)
		return m, cmd
	}

	cmdList := m.updateMenu(msg)
	return m, cmdList
}

func (m model) View() string {
	header := fmt.Sprintf("\n  Welcome to <%s>\n", m.cr.roomName)
	t := m.table.View()
	tableRendered := baseStyle.Render(t)

	descriptionRendered := "> description not set <"
	if m.editDescription || m.showDescription {
		descriptionRendered = m.description.View()
	}

	averageString := ""
	if m.showVotes {
		avg := m.calculateVotesAverage()
		averageString = fmt.Sprintf("Average: %s\n", strconv.FormatFloat(float64(avg), 'f', 2, 32))
	}

	leftSize := lipgloss.JoinVertical(lipgloss.Center, tableRendered)
	leftSize = lipgloss.JoinVertical(lipgloss.Center, leftSize, averageString)
	leftSize = lipgloss.JoinVertical(lipgloss.Center, leftSize, descriptionRendered)
	return header + "\n" + lipgloss.JoinHorizontal(lipgloss.Top, leftSize, m.menu.View())
}

func (m *model) sendHeartbeat() tea.Cmd {
	err := m.cr.Publish(Heartbeat, "")
	if err != nil {
		panic(err)
	}
	return nil
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type EstimatorUI struct {
	p *tea.Program
	m *model
}

func NewEstimationUI(cr *ChatRoom) *EstimatorUI {
	m := model{
		menu:        NewMenu(),
		table:       NewTable(),
		description: NewDescriptionInput(),
		cr:          cr,
		participants: map[string]participant{
			shortID(cr.self): {
				id:   cr.self,
				nick: cr.nick + " (you)",
			},
		},
	}
	ui := EstimatorUI{
		p: tea.NewProgram(m),
		m: &m,
	}

	return &ui
}

func (ui *EstimatorUI) Run() error {
	_, err := ui.p.Run()
	return err
}

func (m *model) self() *participant {
	peer := m.participants[shortID(m.cr.self)]
	return &peer
}
