package moment

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/minus5/gofreetds"
	"log"
	"time"
)

// Location is a geographical point identified by longitude, latitude, and altitude.
type Location struct {
	Latitude  float64
	Longitude float64
	Altitude  float64
}

// Content is a set resources that belong to a Moment.
// Content may be a message, image, and/or a video.
type Content struct {
	Type    uint8
	Message string
	Media   []byte
}

// Moment is the main resource of this package.
// It is a grouping of the Content and Location structures.
// It also contains data on Moment Creation date and Find date.
type Moment struct {
	ID
	Content
	Location
	CreateDate   time.Time
	FindDate     time.Time
	SenderID     string
	RecipientIDs []string
}

// SharedMoment is a moment that another user has found and shared with other users.
func Share(m *Moment) error {

}

// FoundMoment is a moment that a user has found that was left by another user.
func Find(m *Moment) error {

}

// LeftMoment is a moment that a user created and is leaving for others.
func Leave(m *Moment) error {

}

// Poll is the grouping of datatypes needed to check for available lost moments.
func Poll(m *Moment) error {

}

// LostMoment is a moment that has not been found yet.

func (m *Moment) leave() error {
	db := openDbConn()
	defer db.Close()

	insertMoment := ``

	return nil
}

func (m *Moment) find() error {
	db := openDbConn()
	defer db.Close()

	updateLeave := `UPDATE [moment].[Leaves]
					SET Found = 1, 
						FoundDate = ?
					WHERE Recipient = ?
						  AND MomentID = ?`

	updateArgs := []interface{}{time.Now().UTC(), m.Recipient, m.ID}

	if _, err := db.Exec(updateLeave, updateArgs...); err != nil {
		return err
	}

	return nil
}

func (m *Moment) share() error {
	db := openDbConn()
	defer db.Close()

	updateLeave := `UPDATE [moment].[Leaves]
			  		 SET Share = 1
			  	  	 WHERE RecipientID = ?
			  			   AND MomentID = ?`
	leaveArgs := []interface{}{m.RecipientID, m.ID}

	insertShare := `INSERT INTO [moment].[Shares] (ShareID, RecipientID)
					VALUES (`

	t, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := t.Exec(query, args...); err != nil {
		t.Rollback()
		return error
	}

	query = `INSERT INTO`
}

//
// The functions below are utility functions and are only used in this package.
// There functionality has been abstracted and from the above functions for the
// sake of simplicity and readability.
//

// openDbConn is a wrapper for sql.Open() with logging.
func openDbConn() *sql.DB {
	driver := "mssql"
	connStr := "Server=192.168.1.4:1433;Database=Moment-Db;User Id=Reader;Password=123"

	dbConn, err := sql.Open(driver, connStr)
	if err != nil {
		log.Fatal(ConnStrFailed)
	}

	return dbConn
}
