package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/libp2p/go-libp2p/core/peer"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type participant struct {
	id              peer.ID
	nick            string
	heartbeatMisses int
}

type model struct {
	peers map[string]participant
	table table.Model
	cr    *ChatRoom
}

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
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case receiveMsg:
		m.updatePeer(msg)
		return m, m.receiveMsgCmd()
	case tickMsg:
		m.sendHeartbeat()
		m.refreshPeers()
		m.table.SetHeight(len(m.peers))
		m.table, cmd = m.table.Update(msg)
		return m, tea.Batch(tickCmd(), cmd)
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	t := m.table.View()
	return fmt.Sprintf("\n  Welcome to <%s>\n%s\n", m.cr.roomName, baseStyle.Render(t))
}

func (m *model) updatePeer(msg *ChatMessage) {
	sid := shortID(peer.ID(msg.SenderID))
	m.peers[sid] = participant{
		id:              peer.ID(msg.SenderID),
		nick:            msg.SenderNick,
		heartbeatMisses: 0,
	}
}

func (m *model) refreshPeers() {
	rows := []table.Row{}
	for k, p := range m.peers {
		if p.id == m.cr.self {
			continue
		}

		if p.heartbeatMisses > 5 {
			delete(m.peers, k)
			continue
		}
		rows = append(rows, table.Row{p.nick, "-"})
		p.heartbeatMisses++
		m.peers[k] = p
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	self := table.Row{m.peers[shortID(m.cr.self)].nick, "-"}
	rowSelf := []table.Row{self}
	rows = append(rowSelf, rows...)
	m.table.SetRows(rows)
}

func (m *model) sendHeartbeat() tea.Cmd {
	err := m.cr.Publish("heartbeat")
	if err != nil {
		panic(err)
	}
	return nil
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type receiveMsg *ChatMessage

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
	columns := []table.Column{
		{Title: "Today we have with us", Width: 50},
		{Title: "Estimation", Width: 10},
	}

	rows := []table.Row{}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
		table.WithHeight(1),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{
		table: t,
		cr:    cr,
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
