package notice

import (
	"log"
)

type Notifier interface {
	// Send 发送消息
	Send(msg string) error

	// IsAvailable 检查通知渠道是否可用
	IsAvailable() bool

	// GetName 获取通知渠道名称
	GetName() string
}

type NoticeManager struct {
	notifiers []Notifier
}

func NewNoticeManager() *NoticeManager {
	return &NoticeManager{
		notifiers: make([]Notifier, 0),
	}
}

func (m *NoticeManager) AddNotifier(n Notifier) {
	m.notifiers = append(m.notifiers, n)
}

// Notice 直接发送拼接好的字符串消息给所有可用的通知器
func (m *NoticeManager) Notice(message string) {
	for _, n := range m.notifiers {
		if !n.IsAvailable() {
			continue
		}

		if err := n.Send(message); err != nil {
			log.Printf("Failed to send messages via %s: %v", n.GetName(), err)
		}
	}
}
