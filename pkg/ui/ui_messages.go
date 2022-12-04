package ui

import (
	"log"

	"github.com/renato0307/p2p-estimator/pkg/chatroom"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (m *model) handleNewMessage(msg receiveMsg) {
	switch msg.MessageType {
	case chatroom.Heartbeat:
		m.updateParticipants(msg)
	case chatroom.SetDescription:
		m.description.SetValue(msg.Message)
		m.updateDescription(false)
	case chatroom.SendVote:
		m.updateVote(peer.ID(msg.SenderID), msg.Message, false)
	case chatroom.ClearVotes:
		m.clearVotes(false)
	case chatroom.ShowVotes:
		m.displayVotes(false)
	}
}

func (m *model) receiveMsgCmd() tea.Cmd {
	return func() tea.Msg {
		msg := <-m.cr.Messages
		log.Printf("received message %s: %s", msg.MessageType, msg.Message)
		return receiveMsg(msg)
	}
}

func (m *model) sendHeartbeat() tea.Cmd {
	err := m.cr.Publish(chatroom.Heartbeat, "")
	if err != nil {
		panic(err)
	}

	log.Printf("sent heartbeat 💓")
	return nil
}
