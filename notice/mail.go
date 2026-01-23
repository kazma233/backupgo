package notice

import (
	"backupgo/utils"
	"errors"
	"log"
)

type MailNotifier struct {
	mailSender *utils.MailSender
	tos        []string
}

func NewMailNotifier(mailSender *utils.MailSender, tos []string) *MailNotifier {
	return &MailNotifier{
		mailSender: mailSender,
		tos:        tos,
	}
}

func (m *MailNotifier) IsAvailable() bool {
	return m.mailSender != nil && len(m.tos) > 0
}

func (m *MailNotifier) GetName() string {
	return "Mail"
}

// Send 发送邮件
func (m *MailNotifier) Send(content string) error {
	errs := []error{}
	for _, to := range m.tos {
		err := m.mailSender.SendEmail("backupgo", to, "备份消息通知", content)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		log.Printf("Failed to send email: %v", errs)
		return errors.Join(errs...)
	}

	return nil
}
