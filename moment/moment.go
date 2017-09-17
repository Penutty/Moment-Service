package moment

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/minus5/gofreetds"
	"log"
	"strings"
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


type Moment struct {
	Location
	Content
	ID 		   string
	CreateDate time.Time
}

type PublicMoment struct {
	Moment 
	SenderID string
}

type SharedMoment struct {
	Moment
	SenderID string
}

type FoundMoment struct {
	Moment
	SenderID string
	FindDate time.Time
}

type LeftMoment struct {
	Moment
	RecipientIDs []string
}

type LostMoment struct {
	Location
	ID string
}

// Search for Public moments at certain location.
// Location = location
// returns...
// MomentID
// SenderID
// Location
// Content 
// CreateDate
func searchPublic(l Location) ([]PublicMoment, error) {

	return nil
}

// Search for another user's shared moments.
// RecipientID = them
// returns...
// MomentID
// SenderID
// Location
// Content
// CreateDate
func searchShared(u string) ([]SharedMoment, error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT mo.ID, mo.SenderID, mo.Location, mo.CreateDate, m.Type, m.Message, m.Media
			  FROM [moment].[Moments] mo
			  JOIN [moment].[Media] m
			    ON mo.MediaID = m.ID
			  JOIN [moment].[Leaves] l
			    ON mo.ID = l.MomentID
			  JOIN [moment].[Shares] s
			    ON l.ID = s.LeaveID
			  WHERE s.RecipientID = ?`

	args := []interface{}{u}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sM = new([]SharedMoment)
	for rows.Next() {
		m := new(SharedMoment)
		if err = rows.Scan(&m.ID, &m.SenderID, &m.Location, &m.CreateDate, &m.Type, &m.Message, &m.Media); err != nil {
			return nil, err
		}
		sM = append(sM, m)
	}
	if err = rows.Err; err != nil {
		return nil, err
	}

	return sM,nil
}

// Search for my found moments.
// RecipientID = me
// Found = 1
// returns...
// MomentID
// SenderID
// Location
// Content
// CreatDate
// FindDate
func searchFound(u string) ([]FoundMoment, error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT mo.ID, mo.SenderID, mo.Location, mo.CreateDate, l.FindDate m.Type, m.Message, m.Media
			  FROM [moment].[Moments] mo
			  JOIN [moment].[Media] m
			    ON mo.MediaID = m.ID
			  JOIN [moment].[Leaves] l
			    ON mo.ID = l.MomentID
			  WHERE l.RecipientID = ?
			  		AND l.Found = 1`
	args := []interface{}{u}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fM := new([]FoundMoment)
	for rows.Next() {
		m := new(FoundMoment)
		if err = rows.Scan(&m.ID, &m.SenderID, &m.Location, &m.CreateDate, &m.FindDate, &m.Type, &m.Message, &m.Media); err != nil {
			return nil, err
		}
		fM = append(fM, m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return fM, nil
}

// Search for my left moments.
// SenderID = me
// returns... 
// MomentID
// RecipientIDs
// Location
// Content
// CreateDate
func searchLeft(u string) ([]LeftMoment, error) {

	return nil
}

// Search for my lost moments.
// RecipientID = me
// Found = 0
// returns...
// MomentID
// Location
func searchLost(u string) ([]LostMoment, error) {

	return nil
}

// Moment.leave creates a new moment in Moment-Db. 
// The moment content; type, message, and Media are stored.
// The moment is stored.
// If the moment is not public, the leaves are stored.
func (m *Moment) leave() error {
	db := openDbConn()
	defer db.Close()

	insert := `INSERT [moment].[Media] (Type, Message, Media)
					VALUES (?, ?, ?)`
	args := []interface{}{m.Type, m.Message, m.Media}

	t, err := db.Begin()
	if err != nil {
		return err
	}

	res, err := t.Exec(insert, args...)
	if err != nil {
		t.Rollback()
		return err
	}
	MediaID := res.LastInsertID()

	insert = `INSERT [moment].[Moments] (SenderID, Location, MediaID, Public, CreateDate)
					 VALUES (?, ?, ?, ?, ?)`
	args = []interface{}{m.SenderID, m.Location, MediaID, m.Public, m.CreateDate}

	res, err := t.Exec(insert, args...)
	if err != nil {
		t.Rollback()
		return err
	}
	MomentID := res.LastInsertId()

	if m.Public != 1 {
		insert = `INSERT [moment].[Leaves] (MomentID, RecipientID, Found) 
				  VALUES `

		values, args := getSetValues(MomentID, m.RecipientIDs)	
		insert = insert + values

		if _, err = t.Exec(insert, args...); err != nil {
			t.Rollback()
			return err
		}
	}

	t.Commit()

	return nil
}

// Moment.find updates the corresponding moment resource in the Moment-Db.
// The moment's found flag and found date are updated.
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

// Moment.share updates the corresponding leave resource in the Moment-Db.
// The leaves' share flag is set to 1.
// A share is created for each recipient of share in the Shares table. 
func (m *Moment) share() error {
	db := openDbConn()
	defer db.Close()


	update := `UPDATE [moment].[Leaves]
   	  		   SET Share = 1
			   WHERE RecipientID = ?
			  	     AND MomentID = ?`
	args := []interface{}{m.RecipientID, m.ID}

	t, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := t.Exec(update, args...); err != nil {
		t.Rollback()
		return error
	}

	insert := `INSERT INTO [moment].[Shares] (LeaveID, RecipientID)
			   VALUES `

	values, args := getSetValues(FK, m.RecipientIDs)
	insert = insert + values


	
	if _, err := t.Exec(insert, args...); err != nil {
		t.Rollback()
		return err
	}	
	
	t.Commit()

	return nil
}

// getSetValues accepts a list of recipients and returns a insert value for each recipient.
func getRecipientValues(FK interface{}, RecipientIDs []string) (values string, args []interface{}) {
	valueSet := new([]string)
	for _, v := range RecipientIDs {
		values = append(values, "(?, ?, ?)")
		args = append(args, interface{}{FK, v, 0})
	}
	values = strings.Join(valueSet, ", ")

	return values, args
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
