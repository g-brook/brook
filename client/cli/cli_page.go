/*
 * Copyright ©  sixh sixh@apache.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cli

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
