package main

import (
	"fmt"
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
	peers map[string]participant
	cr    *ChatRoom

	menu   list.Model
	choice string
	table  table.Model

	description     textinput.Model
	editDescription bool
	showDescription bool
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
				i, ok := m.menu.SelectedItem().(item)
				if ok {
					m.choice = string(i)
				}

				if m.choice == OPTION_SET_DESCRIPTION {
					m.editDescription = true
					m.description.Focus()
				}
			}
			return m, nil
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case receiveMsg:
		switch msg.MessageType {
		case Heartbeat:
			m.updateParticipants(msg)
		case SetDescription:
			m.description.SetValue(msg.Message)
			m.updateDescription(false)
		}
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

func (m *model) updateDescription(sendDescription bool) {
	m.showDescription = m.description.Value() != ""
	m.editDescription = false
	m.description.Blur()
	if !sendDescription {
		return
	}
	m.sendDescription()
}

func (m model) View() string {
	header := fmt.Sprintf("\n  Welcome to <%s>\n", m.cr.roomName)
	t := m.table.View()
	tableRendered := fmt.Sprintf("%s\n", baseStyle.Render(t))

	descriptionRendered := "> description not set <"
	if m.editDescription || m.showDescription {
		descriptionRendered = m.description.View()
	}
	tableAndDescriptionTitle := lipgloss.JoinVertical(lipgloss.Center, tableRendered, descriptionRendered)
	return header + "\n" + lipgloss.JoinHorizontal(lipgloss.Top, tableAndDescriptionTitle, m.menu.View())
}

func (m *model) sendHeartbeat() tea.Cmd {
	err := m.cr.Publish(Heartbeat, "")
	if err != nil {
		panic(err)
	}
	return nil
}

func (m *model) sendDescription() tea.Cmd {
	err := m.cr.Publish(SetDescription, m.description.Value())
	if err != nil {
		panic(err)
	}
	return nil
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) receiveMsgCmd() tea.Cmd {
	return func() tea.Msg {
		msg := <-m.cr.Messages
		return receiveMsg(msg)
	}
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
		peers: map[string]participant{
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
