package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/renato0307/p2p-estimator/pkg/chatroom"
)

type participant struct {
	id              peer.ID
	nick            string
	currentVote     string
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
		Foreground(lipgloss.NoColor{}).
		Bold(false)
	t.SetStyles(s)

	return t
}

func (m *model) updateParticipants(msg *chatroom.ChatMessage) {
	sid := shortID(peer.ID(msg.SenderID))
	m.participants[sid] = participant{
		id:              peer.ID(msg.SenderID),
		nick:            msg.SenderNick,
		heartbeatMisses: 0,
		currentVote:     m.participants[sid].currentVote,
	}
}

func (m *model) refreshPeers() {
	rows := []table.Row{}
	for k, p := range m.participants {
		if p.id == m.cr.Self {
			continue
		}
		if p.heartbeatMisses > 15 {
			delete(m.participants, k)
			continue
		}
		rows = append(rows, table.Row{p.nick, m.estimationStatus(&p)})
		p.heartbeatMisses++
		m.participants[k] = p
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	self := table.Row{m.self().nick, m.estimationStatus(m.self())}
	rowSelf := []table.Row{self}
	rows = append(rowSelf, rows...)
	m.table.SetRows(rows)
}

func (m *model) updateParticipantsTable(msg tea.Msg) (cmd tea.Cmd) {
	m.refreshPeers()
	m.table.SetHeight(len(m.participants))
	m.table, cmd = m.table.Update(msg)

	return
}

func (m *model) estimationStatus(p *participant) string {
	if p.currentVote == "" {
		return "-"
	}

	if !m.showVotes {
		return "✅"
	}

	val, err := parseVote(p.currentVote)
	if err != nil {
		return "-"
	}

	return fmt.Sprintf("%d", val)
}

func parseVote(currVote string) (int, error) {
	vote := strings.Split(currVote, " ")[0]
	val, err := strconv.Atoi(vote)
	return val, err
}
