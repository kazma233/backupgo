package notice

import (
	"backupgo/config"
	"backupgo/utils"
)

func NewManagerFromConfig(cfg config.GlobalConfig) *NoticeManager {
	manager := NewNoticeManager()

	if cfg.Notice == nil {
		return manager
	}

	if cfg.Notice.Telegram != nil {
		telegramConfig := cfg.Notice.Telegram
		tgBot := utils.NewTgBot(telegramConfig.BotToken)
		manager.AddNotifier(NewTGNotifier(&tgBot, telegramConfig.ChatID))
	}

	if cfg.Notice.Mail != nil {
		mailConfig := cfg.Notice.Mail
		mailSender := utils.NewMailSender(mailConfig.Smtp, mailConfig.Port, mailConfig.User, mailConfig.Password)
		manager.AddNotifier(NewMailNotifier(&mailSender, mailConfig.To))
	}

	return manager
}
