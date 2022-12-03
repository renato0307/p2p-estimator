package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

const (
	OPTION_SET_DESCRIPTION = "Set description 🖍"
	OPTION_CLEAR_VOTES     = "Clear votes 🗑"
	OPTION_SHOW_VOTES      = "Show votes 🔎"
	OPTION_UPDATE_JIRA     = "Update jira 🧙"
)

type item string
type itemDelegate struct{}

func NewMenu() list.Model {
	items := []list.Item{
		item(OPTION_SET_DESCRIPTION),
		item(OPTION_CLEAR_VOTES),
		item(OPTION_SHOW_VOTES),
		item(OPTION_UPDATE_JIRA),
		item("0 points"),
		item("1/2 point"),
		item("1 point"),
		item("2 points"),
		item("3 points"),
		item("5 points"),
		item("8 points"),
		item("13 points"),
		item("20 points"),
		item("40 points"),
		item("No clue 🤷"),
	}

	const defaultWidth = 30

	l := list.New(items, itemDelegate{}, defaultWidth, len(items)+6)
	l.Title = "What do you want to do?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return l
}

func (i item) FilterValue() string                               { return "" }
func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := string(i)
	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + str)
		}
	}

	fmt.Fprint(w, fn(str))
}

func (m *model) updateMenu(msg tea.Msg) tea.Cmd {
	var cmdList tea.Cmd
	m.menu, cmdList = m.menu.Update(msg)
	return cmdList
}
