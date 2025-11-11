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
	"fmt"
	"strconv"
	"strings"

	"github.com/brook/common/hash"
	"github.com/brook/common/version"
	"github.com/charmbracelet/lipgloss"
)

type MainPage struct {
	Title string

	RemoteAddress string

	Status string

	Latency int64

	Connections *hash.SyncMap[string, *TunnelCon]
}

type TunnelCon struct {
	Addr      string
	Port      string
	LocalAddr string
	Protocol  string
	State     string
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
	Title:       style.Render("=========== Welcome Your Come Brook (version:", version.GetBuildVersion(), ") ==========="),
	Latency:     0,
	Status:      "offline",
	Connections: hash.NewSyncMap[string, *TunnelCon](),
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

func UpdateConnections(add string, port int, localAddr string, protocol string, isClose bool) {
	ipAndPort := strings.Split(add, ":")
	s := "OK"
	if isClose {
		s = "NO"
	}
	Page.Connections.Store(strconv.Itoa(port), &TunnelCon{
		Addr:      ipAndPort[0],
		Port:      strconv.Itoa(port),
		Protocol:  protocol,
		LocalAddr: localAddr,
		State:     s,
	})
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
	// 分割线
	sb.WriteString(strings.Repeat("-", 70) + "\n")
	// Connections
	// 表头
	sb.WriteString(fmt.Sprintf("| %-4s | %-4s | %-38s | %-13s |\n", "No.", "P", "Target IP", "Status"))
	sb.WriteString(strings.Repeat("-", 70) + "\n")
	for i, c := range Page.Connections.Values() {
		sb.WriteString(fmt.Sprintf("| %-4d | %-4s | %-38s | %-13s |\n", i+1, c.Protocol, fmt.Sprintf("%s:%s->%s", c.Addr, c.Port, c.LocalAddr), c.State))
	}
	sb.WriteString(strings.Repeat("-", 70) + "\n")

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
