package SQLite3

import (
	"database/sql"
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/common/errors"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
	"path/filepath"
	"sync"
)

const (
	ModuleName                string = "DBProviderSQLite3"
	DBFileName                string = "sqlite3.bd"
	DefaultActivationInterval int64  = 300 // In seconds
)

// Implement DBProvider interface
type DB struct {
	Instance     *sql.DB
	FileFullPath string
	Log          logger.Logger
	LastID       lastID
	LastIDmx     sync.Mutex
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
		return errors.ErrTablesValidationFailed
	}

	// Collect last ID for tables
	LID, err := getLastIDAllTables(db.Instance, db.Log)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't collect last ID for tables - '%v'", err))
		return err
	}
	db.LastIDmx.Lock()
	db.LastID = LID
	db.LastIDmx.Unlock()

	return nil
}

// Get last ID for each table with ID column.
func getLastIDAllTables(db *sql.DB, Log logger.Logger) (lastID, error) {
	var LID lastID

	currentResult, err := getLastIDFromTable(db, Log, "BotUserList", "Created", "ID")
	if err != nil {
		return lastID{}, err
	}
	LID.BotUserList = currentResult

	currentResult, err = getLastIDFromTable(db, Log, "OTRSEventList", "Created", "ID")
	if err != nil {
		return lastID{}, err
	}
	LID.OTRSEventList = currentResult

	currentResult, err = getLastIDFromTable(db, Log, "MessageList", "Created", "ID")
	if err != nil {
		return lastID{}, err
	}
	LID.MessageList = currentResult

	return LID, nil
}

// Get last ID from table by timestamp column.
func getLastIDFromTable(db *sql.DB, Log logger.Logger, tableName, timestampColumn, idColumn string) (int64, error) {
	Log.Debug(fmt.Sprintf("Get last ID from table '%v' by timestamp. Timestamp column name '%v'. ID column name '%v'.",
		tableName, timestampColumn, idColumn))

	// Query provided table for last ID.
	queryString := fmt.Sprintf("SELECT %s FROM %s ORDER BY %s DESC LIMIT 1;",
		idColumn, tableName, timestampColumn)
	Log.Debug(fmt.Sprintf("Query string '%v'", queryString))
	rows, err := db.Query(queryString)
	if err != nil {
		return 0, err
	}

	// Check query result
	var resultID int64 = 0
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&resultID)
		if err != nil {
			Log.Error(fmt.Sprintf("Can't scan last timestamp from table '%s' - '%v'", tableName, err))
			return 0, err
		}
	}
	err = rows.Err()
	if err != nil {
		Log.Error(fmt.Sprintf("While iteration for table '%s' - '%v'", tableName, err))
		return 0, err
	}
	return resultID, nil
}
