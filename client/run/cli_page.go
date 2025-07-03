package run

import (
	"github.com/charmbracelet/lipgloss"
	"strconv"
	"strings"
)

type MainPage struct {
	Title string

	RemoteAddress string

	Status string

	Latency int
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

var CliMainPage = MainPage{
	Title: style.Render("=========== Welcome Your Come  Brook (version:0.0.1) ==========="),
}

func UpdateStatus(status string) {
	switch status {
	case "online":
		CliMainPage.Status = greenStyle.Render(status)
	case "offline":
		CliMainPage.Status = redStyle.Render(status)
	default:
		CliMainPage.Status = status
	}
}

func GetViewPage() *strings.Builder {
	var sb strings.Builder
	writeLine("", &sb, CliMainPage.Title, false)
	writeLine("Status:", &sb, CliMainPage.Status, true)
	writeLine("Remote Address:", &sb, CliMainPage.RemoteAddress, true)
	writeLine("Latency:", &sb, strconv.Itoa(CliMainPage.Latency)+"ms", true)
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
