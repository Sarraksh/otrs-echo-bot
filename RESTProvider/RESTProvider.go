package RESTProvider

import (
	"context"
	"github.com/Sarraksh/otrs-echo-bot/DBProvider"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
)

type RESTProvider interface {
	Initialise(logger logger.Logger, db *DBProvider.DBProvider)
	PrepareListener()
	Listen(ctx context.Context, cancel context.CancelFunc)
}
