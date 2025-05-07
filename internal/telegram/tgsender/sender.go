package tgsender

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Sender struct {
	bot *tgbotapi.BotAPI
}

func NewSender(bot *tgbotapi.BotAPI) *Sender {
	return &Sender{bot: bot}
}

func (s *Sender) SendMessage(chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := s.bot.Send(msg)
	return err
}
