package notice

import (
	"log"
	"time"
)

type Notifier interface {
	// Send 发送消息
	Send(msg string) error

	// IsAvailable 检查通知渠道是否可用
	IsAvailable() bool

	// GetName 获取通知渠道名称
	GetName() string

	// GetFormatType 获取首选的消息格式类型
	GetFormatType() FormatType
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

// NoticeEntries 根据日志条目发送格式化的消息
func (m *NoticeManager) NoticeEntries(taskID string, startTime time.Time, entries []LogEntry) {
	summary := buildTaskSummary(taskID, startTime, entries)
	messages := make(map[FormatType]string)

	for _, n := range m.notifiers {
		if !n.IsAvailable() {
			continue
		}

		formatType := n.GetFormatType()
		msg, ok := messages[formatType]
		if !ok {
			msg = newFormatter(formatType).FormatSummary(summary)
			messages[formatType] = msg
		}

		if err := n.Send(msg); err != nil {
			log.Printf("Failed to send messages via %s: %v", n.GetName(), err)
		}
	}
}
