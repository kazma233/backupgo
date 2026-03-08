package config

import "testing"

func TestParseConfigWithNoticeConfig(t *testing.T) {
	configBlob := []byte(`
notice:
  mail:
    smtp: 'smtp.example.com'
    port: 465
    user: 'user'
    password: 'password'
    to:
      - 'notice@example.com'
  telegram:
    bot_token: '123456:ABCDEF'
    chat_id: '123456789'
backup:
  app:
    back_path: './export'
`)

	cfg, err := ParseConfig(configBlob)
	if err != nil {
		t.Fatalf("ParseConfig returned error: %v", err)
	}

	if cfg.Notice == nil {
		t.Fatal("expected notice config to be present")
	}
	if cfg.Notice.Mail == nil {
		t.Fatal("expected mail config to be present")
	}
	if got := len(cfg.Notice.Mail.To); got != 1 {
		t.Fatalf("expected 1 mail recipient, got %d", got)
	}
	if cfg.Notice.Telegram == nil {
		t.Fatal("expected telegram config to be present")
	}
	if cfg.Notice.Telegram.BotToken != "123456:ABCDEF" {
		t.Fatalf("unexpected telegram bot token: %s", cfg.Notice.Telegram.BotToken)
	}
	if cfg.Notice.Telegram.ChatID != "123456789" {
		t.Fatalf("unexpected telegram chat id: %s", cfg.Notice.Telegram.ChatID)
	}
}
