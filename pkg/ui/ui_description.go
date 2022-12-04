package ui

import (
	"github.com/renato0307/p2p-estimator/pkg/chatroom"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func NewDescriptionInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "What are we estimating?"
	ti.CharLimit = 156
	ti.Width = 20

	return ti
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

func (m *model) sendDescription() tea.Cmd {
	err := m.cr.Publish(chatroom.SetDescription, m.description.Value())
	if err != nil {
		panic(err)
	}
	return nil
}
