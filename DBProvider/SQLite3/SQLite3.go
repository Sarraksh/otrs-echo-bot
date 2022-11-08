package SQLite3

import (
	"database/sql"
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
	"github.com/Sarraksh/otrs-echo-bot/common/myErrors"
	_ "github.com/mattn/go-sqlite3"
	"path/filepath"
)

const (
	ModuleName                string = "DB Provider SQLite3"
	DBFileName                string = "sqlite3.bd"
	DefaultActivationInterval int64  = 300 // In seconds
)

// Implement DBProvider interface
type DB struct {
	Instance     *sql.DB
	FileFullPath string
	Log          logger.Logger
}

// Store last ID for each table with ID column.
type lastID struct {
	BotUserList   int64
	OTRSEventList int64
	MessageList   int64
}

// Accepts the absolute path to the folder for database file.
// Creates a new database file if the file does not exist.
// Validates the table structure and  create table if not exist.
func (db *DB) Initialise(logger logger.Logger, directory string) error {
	db.Log = logger.SetModuleName(ModuleName)
	db.Log.Debug("Initialisation started")
	db.FileFullPath = filepath.Join(directory, DBFileName)
	db.Log.Debug(fmt.Sprintf("Use DB file '%v'", db.FileFullPath))

	// Prepare DB engine.
	dbInstance, err := sql.Open("sqlite3", db.FileFullPath)
	if err != nil {
		db.Log.Error(fmt.Sprintf("DB instance initialisation failed '%v'", err))
		return err
	}
	db.Instance = dbInstance

	// Create tables if not exists.
	err = createAllTablesIfNotExist(db.Instance, db.Log)
	if err != nil {
		return err
	}

	// Validate all tables.
	if !isValidAllTables(db.Instance, db.Log) {
		return myErrors.ErrTablesValidationFailed
	}

	return nil
}

func executeStatement(db *sql.DB, statement string) error {
	transaction, err := db.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()

	preparedStatement, err := transaction.Prepare(statement)
	if err != nil {
		return err
	}
	defer preparedStatement.Close()

	_, err = preparedStatement.Exec()
	if err != nil {
		return err
	}

	// Close transaction.
	err = transaction.Commit()
	if err != nil {
		return err
	}

	return nil
}
