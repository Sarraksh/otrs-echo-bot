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
	db.Log.Info(fmt.Sprintf("Write ne OTRS event with type '%+v' and ticket ID '%+v'", Type, TicketID))

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

	// Prepare insert string.
	Insert := "insert into OTRSEventList(ID, Status, Channel, Type, TicketID, Created, ActivationInterval, NextActivation) values"
	sqlStatement := fmt.Sprintf(
		"%s(%d, '%s', '%s', '%s',%d, %d, %d, %d)",
		Insert,
		ID,
		Status,
		Channel,
		Type,
		TicketIDint,
		Created,
		DefaultActivationInterval,
		NextActivation,
	)
	db.Log.Debug(fmt.Sprintf("Insert string - '%+v'", sqlStatement))

	// Write data into DB.
	_, err = db.Instance.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

// Return OTRS event ticket ID with status "New" or with now activation time and  it's  DB ID.
// If there is no event that meets the conditions, return ErrNoActiveEvents error.
func (db *DB) OTRSEventGetActive() (int64, string, error) {
	// Query provided table for last ID.
	currentTimestamp := time.Now().Unix()
	queryString := fmt.Sprintf("SELECT ID, TicketID FROM OTRSEventList where status in ('New', 'Processing', 'Suspended') and NextActivation < %d LIMIT 1;", currentTimestamp)
	db.Log.Debug(fmt.Sprintf("Query string '%v'", queryString))
	rows, err := db.Instance.Query(queryString)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't get active event - '%+v'", err))
		return 0, "", err
	}
	defer rows.Close()

	// Check query result
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
	sqlStatement := fmt.Sprintf("UPDATE OTRSEventList SET Status = 'Suspended', NextActivation = %d WHERE ID = %d;",
		nextActivation, id)
	db.Log.Debug(fmt.Sprintf("Query string '%v'", sqlStatement))

	_, err := db.Instance.Exec(sqlStatement)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't get update OTRSEventList - '%+v'", err))
		return err
	}

	return nil
}

// Mark event as "Ended" and add current timestamp into "Finished" column.
func (db *DB) OTRSEventEnded(id int64) error {
	Finished := time.Now().Unix()
	sqlStatement := fmt.Sprintf("UPDATE OTRSEventList SET Status = 'Ended', Finished = %d WHERE ID = %d;",
		Finished, id)
	db.Log.Debug(fmt.Sprintf("Query string '%v'", sqlStatement))

	_, err := db.Instance.Exec(sqlStatement)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't get update OTRSEventList - '%+v'", err))
		return err
	}

	return nil
}
