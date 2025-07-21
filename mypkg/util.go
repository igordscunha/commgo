package mypkg

import (
	"fmt"
	"strings"
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit

		case tea.KeyEnter:
			userInput := m.textarea.Value()
			if userInput == "" {
				return m, nil
			}

			m.broadcastFunc(userInput)
			mensagemEstilo := m.senderStyle.Render(time.Now().Format("15:04"), m.username+":") + " " + userInput
			m.messages = append(m.messages, mensagemEstilo)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}

	case IncomingMsg:
		mensagemRecebida := fmt.Sprintf("[%s] %s: %s", msg.Timestamp.Format("15:04"), msg.Username, msg.Texto)
		m.messages = append(m.messages, string(mensagemRecebida))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}
