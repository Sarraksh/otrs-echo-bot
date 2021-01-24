package SQLite3

import (
	"database/sql"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
)

const ModuleName string = "DBProviderSQLite3"

type DB struct {
	Instance *sql.DB
	FileName string
	Log      logger.Logger
}

// TODO
// Accepts the absolute path to the database file.
// Creates a new database file if the file does not exist.
// Validates the table structure if the file exists.
// Returns an error if the specified file is not a table file, is not read / write,
// or contains an invalid table structure.
/*func (db *DB)Initialise(logger logger.Logger, path string) error {
	logger.SetModuleName(ModuleName)
	db.Log = logger
	db.Log.Debug("Initialisation start")
	if isDBFileExists(path) {
		if !isDBFileRW(path) {
			return errors.ErrInsufficientReadOrWritePermissions
		}
		db.Instance, err := open(path)
		validateAllTables()
		return nil
	}
	createNewDB(path)
}*/
