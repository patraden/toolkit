// Copyright 2025 The Toolkit Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LogInfo LogLevel = iota
	LogSuccess
	LogWarning
	LogError
	LogDebug
)

const (
	logInfoStyleColor  = lipgloss.Color("39")
	logSuccStyleColor  = lipgloss.Color("42")
	logWarnStyleColor  = lipgloss.Color("220")
	logErrorStyleColor = lipgloss.Color("196")
	logDebugStyleColor = lipgloss.Color("244")
	logTimestampStyle  = lipgloss.Color("241")
)

// LogEntry represents a single log message
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
}

// LogPanel displays operation logs
type LogPanel struct {
	logs    []LogEntry
	width   int
	height  int
	visible bool
	maxLogs int
	offset  int  // Scroll offset from bottom
	active  bool // Whether the log panel has focus
}

// NewLogPanel creates a new log panel
func NewLogPanel() *LogPanel {
	return &LogPanel{
		logs:    []LogEntry{},
		width:   100,
		height:  8,
		visible: false,
		maxLogs: 100,
	}
}

// AddLog adds a new log entry
func (lp *LogPanel) AddLog(level LogLevel, message string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
	}

	lp.logs = append(lp.logs, entry)

	// Keep only last maxLogs entries
	if len(lp.logs) > lp.maxLogs {
		lp.logs = lp.logs[len(lp.logs)-lp.maxLogs:]
	}

	// Reset scroll to bottom when new log arrives (unless user is scrolling)
	if lp.offset == 0 {
		lp.offset = 0 // Stay at bottom
	}
}

// Clear clears all logs
func (lp *LogPanel) Clear() {
	lp.logs = []LogEntry{}
	lp.offset = 0
}

// Toggle toggles visibility
func (lp *LogPanel) Toggle() {
	lp.visible = !lp.visible
}

// Show makes the log panel visible
func (lp *LogPanel) Show() {
	lp.visible = true
}

// Hide hides the log panel
func (lp *LogPanel) Hide() {
	lp.visible = false
}

// ScrollUp scrolls up in the log history
func (lp *LogPanel) ScrollUp() {
	visibleHeight := lp.height - 2 // Account for border and title
	maxOffset := len(lp.logs) - visibleHeight
	if maxOffset < 0 {
		maxOffset = 0
	}

	if lp.offset < maxOffset {
		lp.offset++
	}
}

// ScrollDown scrolls down in the log history (towards latest)
func (lp *LogPanel) ScrollDown() {
	if lp.offset > 0 {
		lp.offset--
	}
}

// ScrollToBottom scrolls to the latest logs
func (lp *LogPanel) ScrollToBottom() {
	lp.offset = 0
}

// ScrollToTop scrolls to the oldest logs
func (lp *LogPanel) ScrollToTop() {
	visibleHeight := lp.height - 2
	maxOffset := len(lp.logs) - visibleHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	lp.offset = maxOffset
}

// Render renders the log panel
func (lp *LogPanel) Render() string {
	if !lp.visible {
		return ""
	}

	borderColor := lipgloss.Color("240")
	titleColor := lipgloss.Color("39")
	if lp.active {
		borderColor = lipgloss.Color("42") // Green when active
		titleColor = lipgloss.Color("42")
	}

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(borderColor).
		Width(lp.width).
		Height(lp.height)

	titleStyle := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(true).
		Padding(0, 1)

	timestampStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39"))

	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("42"))

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("220"))

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	debugStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	// Title with scroll indicator
	scrollIndicator := ""
	if len(lp.logs) > 0 {
		visibleLogs := lp.height - 2
		if lp.offset > 0 {
			scrollIndicator = fmt.Sprintf(" [↑ %d older]", lp.offset)
		}
		totalHidden := len(lp.logs) - visibleLogs
		if totalHidden > lp.offset && totalHidden > 0 {
			hiddenAtBottom := totalHidden - lp.offset
			if scrollIndicator != "" {
				scrollIndicator += fmt.Sprintf(" [↓ %d newer]", hiddenAtBottom)
			} else {
				scrollIndicator = fmt.Sprintf(" [↓ %d newer]", hiddenAtBottom)
			}
		}
	}

	titleText := "Operation Log"
	if lp.active {
		titleText += " (active)"
	}
	title := titleStyle.Render(titleText + scrollIndicator)

	// Build log lines
	var logLines []string
	visibleLogs := lp.height - 2 // Account for border and title

	// Calculate which logs to display based on offset
	// offset=0 means show latest logs (bottom)
	// offset>0 means scroll up (show older logs)
	endIdx := len(lp.logs) - lp.offset
	startIdx := endIdx - visibleLogs
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx > len(lp.logs) {
		endIdx = len(lp.logs)
	}

	for i := startIdx; i < endIdx; i++ {
		entry := lp.logs[i]

		// Format timestamp
		ts := timestampStyle.Render(entry.Timestamp.Format("15:04:05"))

		// Choose style based on level
		var msgStyle lipgloss.Style
		var prefix string

		switch entry.Level {
		case LogInfo:
			msgStyle = infoStyle
			prefix = "ℹ"
		case LogSuccess:
			msgStyle = successStyle
			prefix = "✓"
		case LogWarning:
			msgStyle = warningStyle
			prefix = "⚠"
		case LogError:
			msgStyle = errorStyle
			prefix = "✗"
		case LogDebug:
			msgStyle = debugStyle
			prefix = "●"
		}

		// Truncate message if too long
		maxMsgLen := lp.width - 15 // Space for timestamp and prefix
		msg := entry.Message
		if len(msg) > maxMsgLen {
			msg = msg[:maxMsgLen-3] + "..."
		}

		line := fmt.Sprintf("%s %s %s",
			ts,
			msgStyle.Render(prefix),
			msgStyle.Render(msg),
		)

		logLines = append(logLines, line)
	}

	// Fill empty space if not enough logs
	for len(logLines) < visibleLogs {
		logLines = append(logLines, "")
	}

	content := title + "\n" + strings.Join(logLines, "\n")

	return borderStyle.Render(content)
}

// GetLevelIcon returns an icon for a log level
func GetLevelIcon(level LogLevel) string {
	switch level {
	case LogInfo:
		return "ℹ"
	case LogSuccess:
		return "✓"
	case LogWarning:
		return "⚠"
	case LogError:
		return "✗"
	case LogDebug:
		return "●"
	default:
		return "•"
	}
}
