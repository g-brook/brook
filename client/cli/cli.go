package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

type tickMsg time.Time

// ---- 消息类型 ----
type (
	fetchSuccessMsg struct{ data string }
	fetchFailedMsg  struct{ err error }
)

type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

func tick() tea.Cmd {
	// 每 2 秒返回一次 tickMsg
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		return m, tick()
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

func (m model) View() string {
	// The header
	sb := GetViewPage()
	sb.WriteString("\nPress q to quit.\n")
	// Send the UI for rendering
	return sb.String()
}

func InitModel() model {
	return model{
		choices:  []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},
		cursor:   0,
		selected: make(map[int]struct{}),
	}
}
