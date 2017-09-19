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

const (
	Datetime2 = "2006-01-02 15:04:05"

	DNE = iota
	Image
	Video
)

var InterfaceTypeNotRecognized = errors.New("The type switch does not recognize the interface type.")
var ConnStrFailed = errors.New("Connection to Moment-Db failed.")

// Location is a geographical point identified by longitude, latitude, and altitude.
type Location struct {
	Latitude  float32
	Longitude float32
}

func (l *Location) String() string {
	return fmt.Sprintf("Latitude: %v\nLongitude: %v\n", l.Latitude, l.Longitude)
}

// Content is a set resources that belong to a Moment.
// Content may be a message, image, and/or a video.
type Content struct {
	Type     uint8
	Message  string
	MediaDir string
}

func (c *Content) String() string {
	return fmt.Sprintf("Type: %v\nMessage: %v\nMediaDir: %v\n", c.Type, c.Message, c.MediaDir)
}

// Moment is the main resource of this package.
// It is a grouping of the Content and Location structs.
type Moment struct {
	ID           int
	SenderID     string
	RecipientID  string
	RecipientIDs []string
	Location
	Content
	Found      bool
	FindDate   time.Time
	Shared     bool
	Public     bool
	CreateDate time.Time
}

func (m *Moment) String() string {
	return fmt.Sprintf(`ID: %v
						SenderID: %v
						RecipientID: %v
						RecipientIDs: %v
						Location: %v
						Content: %v
						Found: %v
						FindDate: %v
						Shared: %v
						Public: %v
						CreateDate: %v`,
		m.ID,
		m.SenderID,
		m.RecipientID,
		m.RecipientIDs,
		m.Location,
		m.Content,
		m.Found,
		m.FindDate,
		m.Shared,
		m.Public,
		m.CreateDate)
}

// Search for Public moments at certain location.
func searchPublic(l Location) ([]Moment, error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT mo.ID,
					 mo.SenderID,
					 mo.Latitude, 
					 mo.Longitude,
					 m.Type,
					 m.Message,
					 m.MediaDir,
					 mo.CreateDate
			  FROM [moment].[Moments] mo
			  JOIN [moment].[Media] m
			    ON mo.MediaID = m.ID
			  WHERE mo.Latitude = ?
			  		AND mo.Longitude = ?`

	rows, err := db.Query(query, l.Latitude, l.Longitude)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := new(Moment)
	ms := make([]Moment, 0)
	var createDate string
	fieldAddrs := []interface{}{
		&m.ID,
		&m.SenderID,
		&m.Latitude,
		&m.Longitude,
		&m.Type,
		&m.Message,
		&m.MediaDir,
		&createDate,
	}

	for rows.Next() {
		if err = rows.Scan(fieldAddrs...); err != nil {
			return nil, err
		}
		m.CreateDate, err = time.Parse(Datetime2, createDate)
		ms = append(ms, *m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ms, nil
}

// Search for another user's shared moments.
func searchShared(u string) ([]Moment, error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT mo.ID, 
					 mo.SenderID, 
					 mo.Latitude,
					 mo.Longitude,
					 mo.CreateDate,
					 m.Type,
					 m.Message,
					 m.MediaDir
			  FROM [moment].[Moments] mo
			  JOIN [moment].[Media] m
			    ON mo.MediaID = m.ID
			  JOIN [moment].[Leaves] l
			    ON mo.ID = l.MomentID
			  WHERE l.RecipientID = ?
			  		AND l.Shared = 1`

	rows, err := db.Query(query, u)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := new(Moment)
	ms := make([]Moment, 0, 100)
	var createDate string
	fieldAddrs := []interface{}{
		&m.ID,
		&m.SenderID,
		&m.Latitude,
		&m.Longitude,
		&createDate,
		&m.Type,
		&m.Message,
		&m.MediaDir,
	}

	for rows.Next() {
		if err = rows.Scan(fieldAddrs...); err != nil {
			return nil, err
		}
		m.CreateDate, err = time.Parse(Datetime2, createDate)
		if err != nil {
			return nil, err
		}
		ms = append(ms, *m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ms, nil
}

// Search for my found moments.
func searchFound(u string) ([]Moment, error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT mo.ID,
					 mo.SenderID,
					 mo.Latitude,
					 mo.Longitude,
					 mo.CreateDate,
					 l.FindDate,
					 m.Type,
					 m.Message,
					 m.MediaDir
			  FROM [moment].[Moments] mo
			  JOIN [moment].[Media] m
			    ON mo.MediaID = m.ID
			  JOIN [moment].[Leaves] l
			    ON mo.ID = l.MomentID
			  WHERE l.RecipientID = ?
			  		AND l.Found = 1`

	rows, err := db.Query(query, u)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := new(Moment)
	ms := make([]Moment, 0)
	var createDate string
	var findDate string
	fieldAddrs := []interface{}{
		&m.ID,
		&m.SenderID,
		&m.Latitude,
		&m.Longitude,
		&createDate,
		&findDate,
		&m.Type,
		&m.Message,
		&m.MediaDir,
	}
	for rows.Next() {
		if err = rows.Scan(fieldAddrs...); err != nil {
			return nil, err
		}
		m.CreateDate, err = time.Parse(Datetime2, createDate)
		if err != nil {
			return nil, err
		}

		m.FindDate, err = time.Parse(Datetime2, findDate)
		if err != nil {
			return nil, err
		}
		ms = append(ms, *m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ms, nil
}

// // Search for my left moments.
func searchLeft(u string) ([]Moment, error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT mo.ID,
					 mo.Latitude,
					 mo.Longitude,
					 mo.CreateDate,
					 m.Type,
					 m.Message,
					 m.MediaDir
			  FROM [moment].[Moments] mo
			  JOIN [moment].[Media] m
			    ON mo.MediaID = m.ID
			  WHERE mo.SenderID = ?`

	rows, err := db.Query(query, u)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := new(Moment)
	ms := make([]Moment, 0)
	var createDate string
	fieldAddrs := []interface{}{
		&m.ID,
		&m.Latitude,
		&m.Longitude,
		&createDate,
		&m.Type,
		&m.Message,
		&m.MediaDir,
	}

	for rows.Next() {

		if err = rows.Scan(fieldAddrs...); err != nil {
			return nil, err
		}

		m.CreateDate, err = time.Parse("2006-01-02 15:04:05", createDate)
		if err != nil {
			return nil, err
		}
		ms = append(ms, *m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	query = `SELECT RecipientID
			  FROM [moment].[Leaves]
			  WHERE MomentID = ?`

	for _, v := range ms {
		rows, err = db.Query(query, v.ID)
		if err != nil {
			return nil, err
		}

		var recipientID string
		for rows.Next() {
			if err = rows.Scan(&recipientID); err != nil {
				return nil, err
			}

			v.RecipientIDs = append(v.RecipientIDs, recipientID)
		}
		if err = rows.Err(); err != nil {
			return nil, err
		}
	}

	return ms, nil
}

// Search for my lost moments.
func searchLost(u string) ([]Moment, error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT mo.ID, mo.Latitude, mo.Longitude
			  FROM [moment].[Moments] mo
			  JOIN [moment].[Leaves] l
			    ON mo.ID = l.MomentID
			  WHERE l.RecipientID = ?`

	rows, err := db.Query(query, u)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := new(Moment)
	ms := make([]Moment, 0)
	for rows.Next() {
		if err = rows.Scan(&m.ID, &m.Latitude, &m.Longitude); err != nil {
			return nil, err
		}
		ms = append(ms, *m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ms, nil
}

// Moment.leave creates a new moment in Moment-Db.
// The moment content; type, message, and MediaDir are stored.
// The moment is stored.
// If the moment is not public, the leaves are stored.
func (m *Moment) leave() error {
	db := openDbConn()
	defer db.Close()

	insert := `INSERT [moment].[Media] (Type, Message, MediaDir)
					VALUES (?, ?, ?)`
	args := []interface{}{m.Type, m.Message, m.MediaDir}

	t, err := db.Begin()
	if err != nil {
		return err
	}

	res, err := t.Exec(insert, args...)
	if err != nil {
		t.Rollback()
		return err
	}

	MediaID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	insert = `INSERT [moment].[Moments] ([SenderID], [Latitude], [Longitude], [MediaID], [Public], [CreateDate])
					 VALUES (?, ?, ?, ?, ?, ?)`
	args = []interface{}{m.SenderID, m.Latitude, m.Longitude, MediaID, m.Public, m.CreateDate}

	res, err = t.Exec(insert, args...)
	if err != nil {
		t.Rollback()
		return err
	}

	MomentID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	if !m.Public {
		insert = `INSERT [moment].[Leaves] (MomentID, RecipientID) 
				  VALUES `

		values, args := getRecipientValues(MomentID, m.RecipientIDs)
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
						FindDate = ?
					WHERE RecipientID = ?
						  AND MomentID = ?`

	updateArgs := []interface{}{time.Now().UTC(), m.RecipientID, m.ID}

	if _, err := db.Exec(updateLeave, updateArgs...); err != nil {
		return err
	}

	return nil
}

// // Moment.share updates the corresponding leave resource in the Moment-Db.
// // The leaves' share flag is set to 1.
// // A share is created for each recipient of share in the Shares table.
func (m *Moment) share() error {
	db := openDbConn()
	defer db.Close()

	update := `UPDATE [moment].[Leaves]
   	  		   SET Shared = 1
   	  		   OUTPUT inserted.ID
			   WHERE RecipientID = ?
			  	     AND MomentID = ?`
	args := []interface{}{m.RecipientID, m.ID}

	t, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			t.Rollback()
			return
		}
		t.Commit()
	}()

	var LeaveID int

	if err = t.QueryRow(update, args...).Scan(&LeaveID); err != nil {
		return err
	}

	if err != nil {
		return err
	}

	insert := `INSERT INTO [moment].[Shares] (LeaveID, RecipientID)
			   VALUES `

	values, args := getRecipientValues(LeaveID, m.RecipientIDs)
	insert = insert + values

	if _, err := t.Exec(insert, args...); err != nil {
		return err
	}

	return nil
}

// getSetValues accepts a list of recipients and returns a insert value for each recipient.
func getRecipientValues(FK interface{}, RecipientIDs []string) (values string, args []interface{}) {
	valueSet := make([]string, 0)
	for _, v := range RecipientIDs {
		valueSet = append(valueSet, "(?, ?)")
		args = append(args, []interface{}{FK, v}...)
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
