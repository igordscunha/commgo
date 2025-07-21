package chat

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

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
			mensagemEnviada := fmt.Sprintf("[%s] %s: %s", time.Now().Format("15:04"), m.username, userInput)
			mensagemEstilo := m.senderStyle.Render(mensagemEnviada)
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
