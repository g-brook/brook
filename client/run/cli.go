package run

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"io/ioutil"
	"net/http"
	"time"
)

const refreshInterval = 5 * time.Second

type status int

const (
	loading status = iota
	success
	failed
)

// ---- 消息类型 ----
type (
	fetchSuccessMsg struct{ data string }
	fetchFailedMsg  struct{ err error }
	tickMsg         struct{}
)

type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
	status   status
	content  string
	err      error
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

// fetchData 命令
func fetchData() tea.Msg {
	resp, err := http.Get("https://httpbin.org/get")
	if err != nil {
		return fetchFailedMsg{err}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fetchFailedMsg{err}
	}
	return fetchSuccessMsg{data: string(body)}
}

// tick 命令：每隔 interval 返回 tickMsg
func tick(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(_ time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m model) View() string {
	// The header
	s := "What should we buy at the market?\n\n"

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

func initModel() model {
	return model{
		choices:  []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},
		cursor:   0,
		selected: make(map[int]struct{}),
	}
}
