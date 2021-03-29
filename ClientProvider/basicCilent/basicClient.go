package basicCilent

import (
	"github.com/Sarraksh/otrs-echo-bot/DBProvider"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
)

type BasicClient struct {
	DB  *DBProvider.DBProvider
	Log logger.Logger
}

// Initialise module.
func (bc *BasicClient) Initialise(db *DBProvider.DBProvider, logger logger.Logger) {
	logger.SetModuleName("ClientProvider")
	bc.Log = logger
	bc.DB = db
}

// Get data from DB.
func (bc *BasicClient) GetTeamByClient(client string) (string, error) {
	db := *bc.DB
	return db.ClientTeamBoundGetTeamByClient(client)
}

// Add new client.
func (bc *BasicClient) AddClient(client string) error {
	db := *bc.DB
	return db.ClientTeamBoundClientAdd(client, "")
}

// Change team for client client.
func (bc *BasicClient) ChangeTeamForClient(client, team string) error {
	db := *bc.DB
	return db.ClientTeamBoundClientUpdate(client, team)
}
