package chat

import (
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gap = "\n\n"

type Mensagem struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Texto     string    `json:"texto"`
	Timestamp time.Time `json:"timestamp"`
}

type IncomingMsg Mensagem

type broadcastFunc func(string)

type model struct {
	viewport      viewport.Model
	messages      []string
	textarea      textarea.Model
	senderStyle   lipgloss.Style
	username      string
	broadcastFunc broadcastFunc
}

func InitialModel(user string, broadcaster broadcastFunc) model {
	ta := textarea.New()
	ta.Placeholder = "Escreva uma mensagem..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 320
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.KeyMap.InsertNewline.SetEnabled(false)
	ta.ShowLineNumbers = false

	vp := viewport.New(85, 15)
	vp.SetContent("Bem vindo ao Chat Offline P2P! Digite alguma coisa para começar...")

	return model{
		textarea:      ta,
		messages:      []string{},
		viewport:      vp,
		senderStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("31")),
		username:      user,
		broadcastFunc: broadcaster,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}
