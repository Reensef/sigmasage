package service

import "github.com/Reensef/sigmasage/internal/telegram/tgsender"

type NotificationConfig struct {
	NotifyTelegram bool
	OrderCreate    bool
}

type Event struct {
}

type Notifier struct {
	sender *tgsender.Sender
}

func NewNotifier(sender *tgsender.Sender) *Notifier {
	return &Notifier{sender: sender}
}

func (n *Notifier) SetNotificationConfig(userID int64, config NotificationConfig) error {
	return nil
}

func (n *Notifier) Notify(userID int64, notification Event) error {
	return nil
}
