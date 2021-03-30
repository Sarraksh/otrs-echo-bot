package SQLite3

import (
	"database/sql"
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
)

// Used in table validation process.
type columnInfo struct {
	CID          int64       // IN BD calls "cid"
	Name         string      // IN BD calls "name"
	Type         string      // IN BD calls "type"
	NotNULL      int64       // IN BD calls "notnull"
	DefaultValue interface{} // IN BD calls "dflt_value"
	PrimaryKey   int64       // IN BD calls  "pk"
}

// Validate all data tables and returns false if one them is invalid.
func isValidAllTables(db *sql.DB, Log logger.Logger) bool {
	var allTablesIsValid bool = true
	Log.Debug("All tables validation started")

	validTableInfo := getValidTableInfo()
	for tableName, tableInfo := range validTableInfo {
		exist, err := isTableExists(db, Log, tableName)
		if err != nil {
			allTablesIsValid = false
			continue
		}
		if !exist {
			allTablesIsValid = false
			continue
		}
		valid, err := isTableValid(db, Log, tableName, tableInfo)
		if err != nil {
			allTablesIsValid = false
			continue
		}
		if !valid {
			allTablesIsValid = false
			continue
		}
	}

	Log.Debug("All tables validation finished")
	return allTablesIsValid
}

// Select from master table for check if table exists.
// Return bool as result and error if occurred on DB engine layer.
// Write to log BD engine errors, validation errors and debug info.
func isTableExists(db *sql.DB, Log logger.Logger, tableName string) (bool, error) {
	Log.Debug(fmt.Sprintf("Start existance check for table '%v'", tableName))
	defer Log.Debug(fmt.Sprintf("Existance check for table '%v' complete", tableName))

	// Query master table for current table
	queryString := fmt.Sprintf("SELECT name FROM sqlite_master WHERE type='table' AND name='%s';", tableName)
	rows, err := db.Query(queryString)
	if err != nil {
		Log.Error(fmt.Sprintf("Can't search table '%s' in master table - '%v'", tableName, err))
		return false, err
	}

	// Check query result
	defer rows.Close()
	for rows.Next() {
		var resultTableName string
		err = rows.Scan(&resultTableName)
		if err != nil {
			Log.Error(fmt.Sprintf("Can't scan result for table '%s' - '%v'", tableName, err))
			return false, err
		}
		if resultTableName == tableName {
			Log.Debug(fmt.Sprintf("Table '%v' exists", tableName))
			return true, nil
		}
	}
	err = rows.Err()
	if err != nil {
		Log.Error(fmt.Sprintf("While iteration for table '%s' - '%v'", tableName, err))
		return false, err
	}
	return false, nil
}

// Query "pragma table_info()" and compare result with provided tableInfo.
// Return bool as result and error if occurred on DB engine layer.
// Write to log BD engine errors, validation errors and debug info.
func isTableValid(db *sql.DB, Log logger.Logger, tableName string, tableInfo []columnInfo) (bool, error) {
	Log.Debug(fmt.Sprintf("Start validation for table '%v'", tableName))
	defer Log.Debug(fmt.Sprintf("Validation check fot table '%v' complete", tableName))

	// Query "pragma table_info()" for detailed columns info
	queryString := fmt.Sprintf("pragma table_info(%s);", tableName)
	rows, err := db.Query(queryString)
	if err != nil {
		Log.Error(fmt.Sprintf("While use 'pragma table_info(%s)' - '%v'", tableName, err))
		return false, err
	}
	defer rows.Close()

	// Check query result
	currentColumnInfo := columnInfo{}
	for rows.Next() {
		err := rows.Scan(
			&currentColumnInfo.CID,
			&currentColumnInfo.Name,
			&currentColumnInfo.Type,
			&currentColumnInfo.NotNULL,
			&currentColumnInfo.DefaultValue,
			&currentColumnInfo.PrimaryKey,
		)
		if err != nil {
			Log.Error(fmt.Sprintf("While iteration for '%s' table info - '%v'", tableName, err))
			return false, err
		}
		if currentColumnInfo.CID >= int64(len(tableInfo)) {
			Log.Error(fmt.Sprintf("More columns then expected while iteration for '%s' table info", tableName))
			return false, nil
		}
		if currentColumnInfo != tableInfo[currentColumnInfo.CID] {
			Log.Error(fmt.Sprintf("Invalid column in table '%s'\nExpected '%+v'\nGet '%+v'\n",
				tableName,
				tableInfo[currentColumnInfo.CID],
				currentColumnInfo,
			))
			return false, nil
		}
	}
	err = rows.Err()
	if err != nil {
		Log.Error(fmt.Sprintf("While iteration for for '%s' table info - '%v'", tableName, err))
		return false, err
	}

	// If currentColumnInfo.Name == "" table has no columns so it's invalid.
	if currentColumnInfo.Name == "" {
		Log.Error(fmt.Sprintf("'%s' table info is empty", tableName))
		return false, nil
	}
	return true, nil
}

// Return
func getValidTableInfo() map[string][]columnInfo {
	result := make(map[string][]columnInfo, 8)

	// BotUserList
	tmpTableInfo := make([]columnInfo, 0, 16)
	tmpTableInfo = append(tmpTableInfo,
		columnInfo{CID: 0, Name: "ID", Type: "integer", NotNULL: 1, DefaultValue: nil, PrimaryKey: 1},
		columnInfo{CID: 1, Name: "Token", Type: "text", NotNULL: 1, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 2, Name: "Active", Type: "integer", NotNULL: 1, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 3, Name: "FirstName", Type: "text", NotNULL: 0, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 4, Name: "LastName", Type: "text", NotNULL: 0, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 5, Name: "Phone", Type: "integer", NotNULL: 0, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 6, Name: "Email", Type: "text", NotNULL: 0, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 7, Name: "Created", Type: "integer", NotNULL: 1, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 8, Name: "TelegramID", Type: "integer", NotNULL: 0, DefaultValue: nil, PrimaryKey: 0},
	)
	result["BotUserList"] = tmpTableInfo

	//OTRSEventList
	tmpTableInfo = make([]columnInfo, 0, 16)
	tmpTableInfo = append(tmpTableInfo,
		columnInfo{CID: 0, Name: "ID", Type: "integer", NotNULL: 1, DefaultValue: nil, PrimaryKey: 1},
		columnInfo{CID: 1, Name: "Status", Type: "text", NotNULL: 1, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 2, Name: "Channel", Type: "text", NotNULL: 1, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 3, Name: "Type", Type: "text", NotNULL: 1, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 4, Name: "TicketID", Type: "integer", NotNULL: 1, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 5, Name: "Created", Type: "integer", NotNULL: 1, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 6, Name: "ActivationInterval", Type: "integer", NotNULL: 0, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 7, Name: "NextActivation", Type: "integer", NotNULL: 0, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 8, Name: "Finished", Type: "integer", NotNULL: 0, DefaultValue: nil, PrimaryKey: 0},
	)
	result["OTRSEventList"] = tmpTableInfo

	//SubscriptionList
	tmpTableInfo = make([]columnInfo, 0, 16)
	tmpTableInfo = append(tmpTableInfo,
		columnInfo{CID: 0, Name: "Active", Type: "integer", NotNULL: 1, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 1, Name: "Subscription", Type: "text", NotNULL: 1, DefaultValue: nil, PrimaryKey: 2},
		columnInfo{CID: 2, Name: "UserID", Type: "integer", NotNULL: 1, DefaultValue: nil, PrimaryKey: 1},
		columnInfo{CID: 3, Name: "Created", Type: "integer", NotNULL: 1, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 4, Name: "Finished", Type: "integer", NotNULL: 0, DefaultValue: nil, PrimaryKey: 0},
	)
	result["SubscriptionList"] = tmpTableInfo

	//SubscriptionScheduler
	tmpTableInfo = make([]columnInfo, 0, 16)
	tmpTableInfo = append(tmpTableInfo,
		columnInfo{CID: 0, Name: "Active", Type: "integer", NotNULL: 1, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 1, Name: "Subscription", Type: "text", NotNULL: 1, DefaultValue: nil, PrimaryKey: 2},
		columnInfo{CID: 2, Name: "UserID", Type: "integer", NotNULL: 1, DefaultValue: nil, PrimaryKey: 1},
		columnInfo{CID: 3, Name: "CreateIn", Type: "integer", NotNULL: 1, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 4, Name: "DeleteIn", Type: "integer", NotNULL: 0, DefaultValue: nil, PrimaryKey: 0},
	)
	result["SubscriptionScheduler"] = tmpTableInfo

	//MessageList
	tmpTableInfo = make([]columnInfo, 0, 16)
	tmpTableInfo = append(tmpTableInfo,
		columnInfo{CID: 0, Name: "ID", Type: "integer", NotNULL: 1, DefaultValue: nil, PrimaryKey: 1},
		columnInfo{CID: 1, Name: "SocialMedia", Type: "text", NotNULL: 1, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 2, Name: "ChatID", Type: "text", NotNULL: 1, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 3, Name: "MessageText", Type: "text", NotNULL: 1, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 4, Name: "Created", Type: "integer", NotNULL: 1, DefaultValue: nil, PrimaryKey: 0},
		columnInfo{CID: 5, Name: "Sent", Type: "integer", NotNULL: 0, DefaultValue: nil, PrimaryKey: 0},
	)
	result["MessageList"] = tmpTableInfo

	//ClientTeamBound
	tmpTableInfo = make([]columnInfo, 0, 16)
	tmpTableInfo = append(tmpTableInfo,
		columnInfo{CID: 0, Name: "Client", Type: "text", NotNULL: 1, DefaultValue: nil, PrimaryKey: 1},
		columnInfo{CID: 1, Name: "Team", Type: "text", NotNULL: 1, DefaultValue: nil, PrimaryKey: 0},
	)
	result["ClientTeamBound"] = tmpTableInfo

	return result
}
