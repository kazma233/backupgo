package utils

import (
	"backupgo/config"
	"testing"
)

func TestSendMessage(t *testing.T) {
	config.InitConfig()

	tgBot := NewTgBot(config.Config.TG.Key)
	resp, err := tgBot.SendMessage(config.Config.TgChatId, "test")
	t.Errorf("resp %v, err %v", resp, err)
}
