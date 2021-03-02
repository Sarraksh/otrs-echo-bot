package SQLite3

import (
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/common/errors"
)

// Add new client with bounded team.
func (db *DB) ClientTeamBoundClientAdd(client, team string) error {
	db.Log.Debug(fmt.Sprintf("Add new client '%v' bounded to team '%+v'", client, team))

	// Check if client already exists.
	_, err := db.ClientTeamBoundGetTeamByClient(client)
	if err != errors.ErrClientNotExists {
		return err
	}

	// Create new sql transaction.
	transaction, err := db.Instance.Begin()
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't create transaction for add new client '%v' with team '%v' - '%v'", client, team, err))
		return err
	}

	// Prepare transaction for insert into table.
	statement, err := transaction.Prepare(`INSERT INTO ClientTeamBound(Client, Team) VALUES(?, ?);`)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't prepare transaction for add new client '%v' with team '%v' - '%v'", client, team, err))
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(client, team)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't execute transaction for add new client '%v' with team '%v' - '%v'", client, team, err))
		return err
	}

	// Close transaction.
	err = transaction.Commit()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While commit for add new client '%v' with team '%v' - '%v'", client, team, err))
		return err
	}

	db.Log.Debug(fmt.Sprintf("Succesfully add new client '%v' bounded to team '%+v'", client, team))
	return nil
}

// Update bounded team for client.
func (db *DB) ClientTeamBoundClientUpdate(client, team string) error {
	db.Log.Debug(fmt.Sprintf("Update client '%v' bound to team '%+v'", client, team))

	// Check if client exists.
	_, err := db.ClientTeamBoundGetTeamByClient(client)
	if err != nil {
		return err
	}

	// Create new sql transaction.
	transaction, err := db.Instance.Begin()
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't create transaction for update client '%v' with team '%v' - '%v'", client, team, err))
		return err
	}

	// Prepare transaction for update table.
	statement, err := transaction.Prepare(`UPDATE ClientTeamBound SET Team = ? WHERE Client = ?`)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't prepare transaction for update client '%v' with team '%v' - '%v'", client, team, err))
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(team, client)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't execute transaction for update client '%v' with team '%v' - '%v'", client, team, err))
		return err
	}

	// Close transaction.
	err = transaction.Commit()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While commit for update client '%v' with team '%v' - '%v'", client, team, err))
		return err
	}

	db.Log.Debug(fmt.Sprintf("Succesfully update client '%v' bound to team '%+v'", client, team))
	return nil
}

// Return team name bonded with provided client.
// If team not bounded return ErrNoTeamBounded.
// If client not exists return ErrClientNotExists.
// If more then one team fined return ErrMoreThenOneTeamBounded.
func (db *DB) ClientTeamBoundGetTeamByClient(client string) (string, error) {
	db.Log.Debug(fmt.Sprintf("Get team for client '%+v'", client))
	// Create new sql transaction.
	transaction, err := db.Instance.Begin()
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't create transaction for get team for client '%v' - '%v'", client, err))
		return "", err
	}

	// Prepare transaction for select from table.
	statement, err := transaction.Prepare(`SELECT Team FROM ClientTeamBound WHERE Client = ?;`)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't prepare transaction for get team for client '%v' - '%v'", client, err))
		return "", err
	}
	defer statement.Close()

	// Query team.
	rows, err := statement.Query(client)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't query for get team for client '%v' - '%v'", client, err))
		return "", err
	}
	defer rows.Close()

	// Check query result.
	team := ""
	rowNumber := 0
	for rows.Next() {
		err = rows.Scan(&team)
		if err != nil {
			db.Log.Error(fmt.Sprintf("Can't scan team by client '%v' - '%v'", client, err))
			return "", err
		}
		rowNumber++
	}
	err = rows.Err()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While iteration for get team for client '%v' - '%v'", client, err))
		return "", err
	}

	// Handle expected logical cases.
	switch {
	case rowNumber == 0:
		db.Log.Debug(fmt.Sprintf("Client '%+v' not exist", client))
		return "", errors.ErrClientNotExists
	case rowNumber == 1:
		if team == "" {
			db.Log.Debug(fmt.Sprintf("Client '%+v' has no team bounded", client))
			return "", errors.ErrNoTeamBounded
		}
	case rowNumber > 1:
		db.Log.Debug(fmt.Sprintf("Client '%+v' has more then one team bounded", client))
		return "", errors.ErrMoreThenOneTeamBounded
	}

	// Close transaction.
	err = transaction.Commit()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While commit for get team for client '%v' - '%v'", client, err))
		return "", err
	}

	db.Log.Debug(fmt.Sprintf("Client '%+v' bound to team '%v'", client, team))
	return team, nil
}
