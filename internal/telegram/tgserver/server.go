package tgserver

import (
	"context"

	"github.com/Reensef/sigmasage/internal/telegram/tgsender"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Server struct {
	bot      *tgbotapi.BotAPI
	sender   *tgsender.Sender
	stopFunc context.CancelFunc
}

func NewServer(bot *tgbotapi.BotAPI, sender *tgsender.Sender) *Server {
	return &Server{
		bot:    bot,
		sender: sender,
	}
}

func (s *Server) Run() error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 1

	updatesChan := s.bot.GetUpdatesChan(updateConfig)

	var ctx context.Context
	ctx, s.stopFunc = context.WithCancel(context.Background())

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-updatesChan:
				if update.Message != nil {
					s.handleUpdate(update)
				}
			}
		}
	}(ctx)

	return nil
}

func (s *Server) Stop() {
	s.stopFunc()
}

func (s *Server) handleUpdate(update tgbotapi.Update) {
	s.sender.SendMessage(update.Message.Chat.ID, "Hello, mthfk!")
}
