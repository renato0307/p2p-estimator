package main

import "github.com/charmbracelet/bubbles/textinput"

func NewDescriptionInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "What are we estimating?"
	ti.CharLimit = 156
	ti.Width = 20

	return ti
}
