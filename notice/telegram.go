package notice

import (
	"backupgo/utils"
	"log"
)

type TGNotifier struct {
	tg *utils.TGBot
	to string
}

func NewTGNotifier(tg *utils.TGBot, to string) *TGNotifier {
	return &TGNotifier{
		tg: tg,
		to: to,
	}
}

func (m *TGNotifier) SendMessageNow(content string) (string, error) {
	resp, err := m.tg.SendMessage(m.to, content)
	if err != nil {
		return resp, err
	}

	return "", nil
}

func (m *TGNotifier) IsAvailable() bool {
	return m.tg != nil && m.to != ""
}

func (m *TGNotifier) GetName() string {
	return "Telegram"
}

func (m *TGNotifier) Send(msg string) error {
	resp, err := m.tg.SendMessage(m.to, msg)
	log.Printf("Telegram response: %s", resp)
	if err != nil {
		return err
	}

	return nil
}
