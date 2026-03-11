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
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
	"github.com/g-brook/brook/common/log"
	"go.uber.org/zap/zapcore"
)

// level tag styles — matches the cyber palette in cli_page.go
var (
	tagDebug = lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true).Render("[DBG]")
	tagInfo  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00F5FF")).Bold(true).Render("[INF]")
	tagWarn  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C")).Bold(true).Render("[WRN]")
	tagError = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true).Render("[ERR]")
	tagFatal = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2D78")).Bold(true).Render("[FTL]")
)

// WriteSyncer implements zapcore.WriteSyncer and routes output to the TUI log panel.
type WriteSyncer struct {
	mu    sync.Mutex
	level zapcore.Level
}

// NewCLIWriteSyncer creates a new CLI write syncer.
func NewCLIWriteSyncer(level string) *WriteSyncer {
	if level == "" {
		level = "info"
	}
	return &WriteSyncer{
		level: log.ParseLevel(level),
	}
}

func (w *WriteSyncer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	msg := strings.TrimSpace(string(p))
	if msg != "" {
		AddLog("%s %s", tagInfo, msg)
	}
	return len(p), nil
}

func (w *WriteSyncer) Sync() error { return nil }

// FormattedCLIWriteSyncer writes level-tagged, styled log lines to the TUI.
type FormattedCLIWriteSyncer struct {
	mu     sync.Mutex
	level  zapcore.Level
	prefix string
}

func NewFormattedCLIWriteSyncer(level zapcore.Level, prefix string) *FormattedCLIWriteSyncer {
	return &FormattedCLIWriteSyncer{
		level:  level,
		prefix: prefix,
	}
}

func (w *FormattedCLIWriteSyncer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	msg := strings.TrimSpace(string(p))
	if msg == "" {
		return len(p), nil
	}

	var tag string
	switch w.level {
	case zapcore.DebugLevel:
		tag = tagDebug
	case zapcore.InfoLevel:
		tag = tagInfo
	case zapcore.WarnLevel:
		tag = tagWarn
	case zapcore.ErrorLevel:
		tag = tagError
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		tag = tagFatal
	default:
		tag = lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2")).Bold(true).
			Render(fmt.Sprintf("[%s]", strings.ToUpper(w.level.String())))
	}

	logMsg := fmt.Sprintf("%s %s", tag, msg)
	if w.prefix != "" {
		prefix := lipgloss.NewStyle().Foreground(lipgloss.Color("#44475A")).Render(w.prefix)
		logMsg = fmt.Sprintf("%s %s", prefix, logMsg)
	}

	AddLog("%s", logMsg)
	return len(p), nil
}

func (w *FormattedCLIWriteSyncer) Sync() error { return nil }
