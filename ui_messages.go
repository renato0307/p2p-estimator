package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (m *model) handleNewMessage(msg receiveMsg) {
	switch msg.MessageType {
	case Heartbeat:
		m.updateParticipants(msg)
	case SetDescription:
		m.description.SetValue(msg.Message)
		m.updateDescription(false)
	case SendVote:
		m.updateVote(peer.ID(msg.SenderID), msg.Message, false)
	case ClearVotes:
		m.clearVotes(false)
	case ShowVotes:
		m.displayVotes(false)
	}
}

func (m *model) receiveMsgCmd() tea.Cmd {
	return func() tea.Msg {
		msg := <-m.cr.Messages
		return receiveMsg(msg)
	}
}
