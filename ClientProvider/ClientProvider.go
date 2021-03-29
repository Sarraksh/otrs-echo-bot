package ClientProvider

import (
	"github.com/Sarraksh/otrs-echo-bot/DBProvider"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
)

type ClientProvider interface {
	Initialise(db *DBProvider.DBProvider, logger logger.Logger)
	GetTeamByClient(client string) (string, error)
	AddClient(client string) error
	ChangeTeamForClient(client, team string) error
}
