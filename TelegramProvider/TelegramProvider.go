package TelegramProvider

import (
	"context"
	"github.com/Sarraksh/otrs-echo-bot/DBProvider"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
)

type TelegramProvider interface {
	Initialise(botToken string, logger logger.Logger, db *DBProvider.DBProvider)
	UpdateListener(ctx context.Context, cancel context.CancelFunc) error
	SendEventMessage(chatID int64, text string)
}
