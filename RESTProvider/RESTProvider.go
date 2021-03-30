package RESTProvider

import (
	"context"
	"github.com/Sarraksh/otrs-echo-bot/DBProvider"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
	"github.com/Sarraksh/otrs-echo-bot/event"
)

type RESTProvider interface {
	Initialise(logger logger.Logger, db *DBProvider.DBProvider)
	PrepareListener(eventProcessor *event.Processor)
	Listen(ctx context.Context, cancel context.CancelFunc) error
}
