package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (m *model) calculateVotesAverage() float32 {
	var validVotes float32
	var sum float32
	for _, p := range m.participants {
		val, err := parseVote(p.currentVote)
		if err != nil {
			continue
		}
		validVotes++
		sum += float32(val)
	}
	return sum / validVotes
}

func (m *model) updateVote(pid peer.ID, vote string, sendMsg bool) tea.Cmd {
	m.setVote(pid, vote)

	var numberOfVotes int
	for _, p := range m.participants {
		if p.currentVote != "" {
			numberOfVotes++
		}
	}
	m.showVotes = numberOfVotes == len(m.participants)
	if m.showVotes {
		m.refreshPeers()
		m.updateParticipantsTable(nil)
	}

	if !sendMsg {
		return nil
	}

	err := m.cr.Publish(SendVote, vote)
	if err != nil {
		panic(err)
	}
	return nil
}

func (m *model) clearVotes(sendMsg bool) tea.Cmd {
	for i, p := range m.participants {
		p.currentVote = ""
		m.participants[i] = p
	}
	m.showVotes = false

	if !sendMsg {
		return nil
	}

	err := m.cr.Publish(ClearVotes, "")
	if err != nil {
		panic(err)
	}
	return nil
}

func (m *model) displayVotes(sendMsg bool) tea.Cmd {
	m.showVotes = true

	if !sendMsg {
		return nil
	}

	err := m.cr.Publish(ShowVotes, "")
	if err != nil {
		panic(err)
	}
	return nil
}

func (m *model) setVote(pid peer.ID, vote string) {
	p := m.participants[shortID(pid)]
	p.currentVote = vote
	m.participants[shortID(pid)] = p
}
