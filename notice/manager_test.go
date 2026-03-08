package notice

import (
	"strings"
	"testing"
	"time"
)

type stubNotifier struct {
	name       string
	formatType FormatType
	available  bool
	sent       []string
	err        error
}

func (n *stubNotifier) Send(msg string) error {
	n.sent = append(n.sent, msg)
	return n.err
}

func (n *stubNotifier) IsAvailable() bool {
	return n.available
}

func (n *stubNotifier) GetName() string {
	return n.name
}

func (n *stubNotifier) GetFormatType() FormatType {
	return n.formatType
}

func TestNoticeManagerUsesNotifierFormats(t *testing.T) {
	manager := NewNoticeManager()
	plainNotifier := &stubNotifier{name: "tg", formatType: FormatTypePlain, available: true}
	htmlNotifier := &stubNotifier{name: "mail", formatType: FormatTypeHTML, available: true}

	manager.AddNotifier(plainNotifier)
	manager.AddNotifier(htmlNotifier)

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []LogEntry{
		{EntryType: EntryTypeStep, Timestamp: start.Add(time.Second), StepName: "备份", StepStatus: StepStatusSuccess, Message: "完成: 备份"},
	}

	manager.NoticeEntries("task-1", start, entries)

	if len(plainNotifier.sent) != 1 {
		t.Fatalf("expected plain notifier to receive 1 message, got %d", len(plainNotifier.sent))
	}
	if len(htmlNotifier.sent) != 1 {
		t.Fatalf("expected html notifier to receive 1 message, got %d", len(htmlNotifier.sent))
	}
	if !strings.Contains(plainNotifier.sent[0], "📦 备份任务: task-1") {
		t.Fatalf("unexpected plain message: %s", plainNotifier.sent[0])
	}
	if !strings.Contains(htmlNotifier.sent[0], "<b>📦 备份任务:</b> <code>task-1</code>") {
		t.Fatalf("unexpected html message: %s", htmlNotifier.sent[0])
	}
}

func TestNotifierFormatTypes(t *testing.T) {
	if (&TGNotifier{}).GetFormatType() != FormatTypePlain {
		t.Fatalf("telegram notifier should use plain format")
	}
	if (&MailNotifier{}).GetFormatType() != FormatTypeHTML {
		t.Fatalf("mail notifier should use html format")
	}
}
