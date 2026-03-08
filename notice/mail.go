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

func (m *MailNotifier) GetFormatType() FormatType {
	return FormatTypeHTML
}

// Send 发送邮件 (使用HTML格式)
func (m *MailNotifier) Send(content string) error {
	errs := []error{}
	contentType := "text/html; charset=UTF-8"
	for _, to := range m.tos {
		err := m.mailSender.SendEmailWithContentType("backupgo", to, "备份消息通知", content, contentType)
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
