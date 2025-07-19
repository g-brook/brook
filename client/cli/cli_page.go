package cli

import (
	"github.com/charmbracelet/lipgloss"
	"strconv"
	"strings"
)

type MainPage struct {
	Title string

	RemoteAddress string

	Status string

	Latency int64
}

var style = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#2E8B57")).
	Bold(true).
	Italic(false).
	Underline(true)

// #E32636
var redStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#F20C00"))

var greenStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#228B22")).Bold(true)

var Page = MainPage{
	Title:   style.Render("=========== Welcome Your Come  Brook (version:0.0.1) ==========="),
	Latency: 0,
	Status:  "offline",
}

func UpdateStatus(status string) {
	switch status {
	case "online":
		Page.Status = greenStyle.Render(status)
	case "offline":
		Page.Status = redStyle.Render(status)
	default:
		Page.Status = status
	}
}

func UpdateSpell(ms int64) {
	Page.Latency = ms
}

func GetViewPage() *strings.Builder {
	var sb strings.Builder
	writeLine("", &sb, Page.Title, false)
	writeLine("Status:", &sb, Page.Status, true)
	writeLine("Remote Address:", &sb, Page.RemoteAddress, true)
	writeLine("Latency:", &sb, strconv.FormatInt(Page.Latency, 10)+"ms", true)
	return &sb
}

func writeLine(prefix string, sb *strings.Builder, text string, leftPad bool) {
	// 补充前缀长度到16字符，实现右对齐效果
	for len(prefix) < 30 && leftPad {
		prefix += " "
	}
	sb.WriteString(prefix)
	sb.WriteString(text)
	sb.WriteString("\n")
}
