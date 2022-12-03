package main

import (
	"sort"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/libp2p/go-libp2p/core/peer"
)

type participant struct {
	id              peer.ID
	nick            string
	heartbeatMisses int
}

func NewTable() table.Model {
	columns := []table.Column{
		{Title: "Today we have with us", Width: 50},
		{Title: "Estimation", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
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

	return t
}

func (m *model) updateParticipants(msg *ChatMessage) {
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

func (m *model) updateParticipantsTable(msg tea.Msg) (cmd tea.Cmd) {
	m.refreshPeers()
	m.table.SetHeight(len(m.peers))
	m.table, cmd = m.table.Update(msg)

	return
}
