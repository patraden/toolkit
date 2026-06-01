package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type HelpView int

const (
	defaultView HelpView = iota
)

const (
	helpKeyStyleColor  = lipgloss.Color("39")
	helpDescStyleColor = lipgloss.Color("250")
)

type HelpBar struct {
	helps     []string
	width     int
	height    int
	keyStyle  lipgloss.Style
	descStyle lipgloss.Style
	helpStyle lipgloss.Style
	view      HelpView
}

func NewHelpBar(width, height int) *HelpBar {
	keyStyle := lipgloss.NewStyle().Foreground(helpKeyStyleColor).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(helpDescStyleColor)
	helpStyle := lipgloss.NewStyle().
		Foreground(helpDescStyleColor).
		Width(width).
		Height(height).
		Padding(0, 1)

	return &HelpBar{
		helps:     []string{},
		width:     width,
		height:    height,
		keyStyle:  keyStyle,
		descStyle: descStyle,
		helpStyle: helpStyle,
		view:      defaultView,
	}
}

func (h *HelpBar) defaultHelps() []string {
	return []string{
		h.keyStyle.Render("F6") + " " + h.descStyle.Render("Unfocus"),
		h.keyStyle.Render("↑↓") + " " + h.descStyle.Render("Scroll"),
		h.keyStyle.Render("PgUp/PgDn") + " " + h.descStyle.Render("Fast Scroll"),
		h.keyStyle.Render("Home/End") + " " + h.descStyle.Render("Top/Bottom"),
		h.keyStyle.Render("F4") + " " + h.descStyle.Render("Hide"),
		h.keyStyle.Render("Q") + " " + h.descStyle.Render("Quit"),
	}
}

func (h *HelpBar) Render() string {
	var helps []string
	helps = append(helps, h.defaultHelps()...)

	helpText := strings.Join(helps, " │ ")
	return h.helpStyle.Render(helpText)
}

func (h *HelpBar) SetWidth(width int) {
	h.width = width
}

func (h *HelpBar) SetHeight(height int) {
	h.height = height
}
