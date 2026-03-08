package notice

import (
	"backupgo/config"
	"backupgo/utils"
)

func NewManagerFromConfig(cfg config.GlobalConfig) *NoticeManager {
	manager := NewNoticeManager()

	if cfg.TG != nil {
		tgBot := utils.NewTgBot(cfg.TG.Key)
		manager.AddNotifier(NewTGNotifier(&tgBot, cfg.TgChatId))
	}

	if cfg.Mail != nil {
		mailConfig := cfg.Mail
		mailSender := utils.NewMailSender(mailConfig.Smtp, mailConfig.Port, mailConfig.User, mailConfig.Password)
		manager.AddNotifier(NewMailNotifier(&mailSender, cfg.NoticeMail))
	}

	return manager
}
