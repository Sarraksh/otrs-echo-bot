package SQLite3

import (
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/common/errors"
	"strconv"
	"time"
)

// Create new OTRS event in database.
// Increment LastID.OTRSEventList even if error occurred,
func (db *DB) OTRSEventCreateNew(Channel, Type, TicketID string) error {
	db.Log.Info(fmt.Sprintf("Write new OTRS event with type '%+v' and ticket ID '%+v'", Type, TicketID))

	// Prepare data for insert.
	db.LastIDmx.Lock()
	ID := db.LastID.OTRSEventList + 1
	db.LastID.OTRSEventList = ID
	db.LastIDmx.Unlock()
	Status := "New"
	TicketIDint, err := strconv.Atoi(TicketID)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't create new OTRS event - '%v'", err))
		return err
	}
	Created := time.Now().Unix()
	NextActivation := Created + DefaultActivationInterval

	// Create new sql transaction.
	transaction, err := db.Instance.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()

	// Prepare and execute transaction for update row.
	statement, err := transaction.Prepare(`INSERT INTO OTRSEventList(ID, Status, Channel, Type, TicketID, Created, ActivationInterval, NextActivation)
values(?, ?, ?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return err
	}
	defer statement.Close()

	// Update data into DB.
	_, err = statement.Exec(
		ID,
		Status,
		Channel,
		Type,
		TicketIDint,
		Created,
		DefaultActivationInterval,
		NextActivation,
	)
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

// Return OTRS event ticket ID with status "New", 'Processing' or 'Suspended' and it's DB ID.
// If there is no event that meets the conditions, return ErrNoActiveEvents error.
func (db *DB) OTRSEventGetActive() (int64, string, error) {
	// Query provided table for last ID.
	currentTimestamp := time.Now().Unix()

	// Create new sql transaction.
	transaction, err := db.Instance.Begin()
	if err != nil {
		return 0, "", err
	}
	defer transaction.Rollback()

	// Prepare and execute transaction for update row.
	statement, err := transaction.Prepare(
		`SELECT ID, TicketID FROM OTRSEventList where status in ('New', 'Processing', 'Suspended') and NextActivation < ? LIMIT 1;`,
	)
	if err != nil {
		return 0, "", err
	}
	defer statement.Close()

	// Query provided table for last ID.
	rows, err := statement.Query(currentTimestamp)
	if err != nil {
		return 0, "", err
	}

	// Check query result.
	var ID int64 = 0
	var ticketIDint int64
	for rows.Next() {
		err = rows.Scan(&ID, &ticketIDint)
		if err != nil {
			db.Log.Error(fmt.Sprintf("Can't scan for active event - '%+v'", err))
			return 0, "", err
		}
	}
	err = rows.Err()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While iteration for active event '%v'", err))
		return 0, "", err
	}
	defer rows.Close()

	// Close transaction.
	err = transaction.Commit()
	if err != nil {
		return 0, "", err
	}

	// Check if no one row received.
	if ID == 0 {
		return 0, "", errors.ErrNoActiveEvents
	}

	ticketID := fmt.Sprint(ticketIDint)
	return ID, ticketID, nil
}

// Return earliest event activation timestamp for active events.
func (db *DB) OTRSEventGetEarliestActivationTimestamp() (int64, error) {
	// Query provided table for last ID.
	queryString := "SELECT Created FROM OTRSEventList where Status in ('New', 'Processing', 'Suspended') ORDER BY NextActivation LIMIT 1;"
	db.Log.Debug(fmt.Sprintf("Query string '%v'", queryString))
	rows, err := db.Instance.Query(queryString)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't get earliest activation timestamp - '%+v'", err))
		return 0, err
	}
	defer rows.Close()

	// Check query result
	var earliestTimestamp int64 = 0
	for rows.Next() {
		err = rows.Scan(&earliestTimestamp)
		if err != nil {
			db.Log.Error(fmt.Sprintf("Can't scan for earliest activation timestamp - '%+v'", err))
			return 0, err
		}
	}
	err = rows.Err()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While iteration for earliest activation timestamp '%v'", err))
		return 0, err
	}

	// Check if no one row received. In this case return minimal activation interval.
	if earliestTimestamp == 0 {
		return DefaultActivationInterval, nil
	} else {
		return earliestTimestamp, nil
	}
}

// Mark event as "Suspended" and renew "NextActivation" time.
func (db *DB) OTRSEventSuspend(id int64) error {
	nextActivation := time.Now().Unix() + DefaultActivationInterval

	// Create new sql transaction.
	transaction, err := db.Instance.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()

	// Prepare and execute transaction for update row.
	statement, err := transaction.Prepare(`UPDATE OTRSEventList SET Status = 'Suspended', NextActivation = ? WHERE ID = ?;`)
	if err != nil {
		return err
	}
	defer statement.Close()

	// Update data into DB.
	_, err = statement.Exec(nextActivation, id)
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

// Mark event as "Ended" and add current timestamp into "Finished" column.
func (db *DB) OTRSEventEnded(id int64) error {
	Finished := time.Now().Unix()

	// Create new sql transaction.
	transaction, err := db.Instance.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()

	// Prepare and execute transaction for update row.
	statement, err := transaction.Prepare(`UPDATE OTRSEventList SET Status = 'Ended', Finished = ? WHERE ID = ?;`)
	if err != nil {
		return err
	}
	defer statement.Close()

	// Update data into DB.
	_, err = statement.Exec(Finished, id)
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
