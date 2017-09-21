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
	FinderID     string
	RecipientIDs []string
	Location
	Content
	Found      bool
	FindDate   time.Time
	Shared     bool
	Public     bool
	Hidden     bool
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

// searchHidden looks up public moments that a user has found that are flagged as hidden.
func searchFoundPublic(u string) (ms []Moment, err error) {
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
			  JOIN [moment].[Finds] f
			    ON mo.ID = f.MomentID
			  WHERE FinderID = ?`

	rows, err := db.Query(query, u)
	if err != nil {
		return
	}
	defer rows.Close()

	m := new(Moment)
	var createDate string
	destAddrs := []interface{}{
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
		if err = rows.Scan(destAddrs...); err != nil {
			return
		}
		m.CreateDate, err = time.Parse(Datetime2, createDate)
		if err != nil {
			return
		}
		ms = append(ms, *m)
	}
	if err = rows.Err(); err != nil {
		return
	}

	return

}

// searchPublic queries Moment-Db for moments that are public and not hidden.
func searchPublic(l Location) (ms []Moment, err error) {
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
			  		AND mo.Longitude = ?
			  		AND mo.Hidden = 0
			  		AND mo.[Public] = 1`

	rows, err := db.Query(query, l.Latitude, l.Longitude)
	if err != nil {
		return
	}
	defer rows.Close()

	m := new(Moment)
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
			return
		}
		m.CreateDate, err = time.Parse(Datetime2, createDate)
		if err != nil {
			return
		}
		ms = append(ms, *m)
	}
	if err = rows.Err(); err != nil {
		return
	}

	return
}

// searchShared queries Moment-Db for moments that a user has found, and shared with others.
func searchShared(u string) (ms []Moment, err error) {
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
		return
	}
	defer rows.Close()

	m := new(Moment)
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
			return
		}
		m.CreateDate, err = time.Parse(Datetime2, createDate)
		if err != nil {
			return
		}
		ms = append(ms, *m)
	}
	if err = rows.Err(); err != nil {
		return
	}

	return
}

// searchFound queries Moment-Db for moments that a user has been left and has found.
func searchFound(u string) (ms []Moment, err error) {
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
		return
	}
	defer rows.Close()

	m := new(Moment)
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
			return
		}
		m.CreateDate, err = time.Parse(Datetime2, createDate)
		if err != nil {
			return
		}

		m.FindDate, err = time.Parse(Datetime2, findDate)
		if err != nil {
			return
		}
		ms = append(ms, *m)
	}
	if err = rows.Err(); err != nil {
		return
	}

	return
}

// searchLeft queries Moment-Db for moments a user has left for others to find.
func searchLeft(u string) (ms []Moment, err error) {
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
		return
	}
	defer rows.Close()

	m := new(Moment)
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
			return
		}

		m.CreateDate, err = time.Parse("2006-01-02 15:04:05", createDate)
		if err != nil {
			return
		}
		ms = append(ms, *m)
	}
	if err = rows.Err(); err != nil {
		return
	}

	query = `SELECT RecipientID
			  FROM [moment].[Leaves]
			  WHERE MomentID = ?`

	for _, v := range ms {
		rows, err = db.Query(query, v.ID)
		if err != nil {
			return
		}

		var recipientID string
		for rows.Next() {
			if err = rows.Scan(&recipientID); err != nil {
				return
			}

			v.RecipientIDs = append(v.RecipientIDs, recipientID)
		}
		if err = rows.Err(); err != nil {
			return
		}
	}

	return
}

// searchLost queries Moment-Db for moments others have left for the specified user to find.
func searchLost(u string) (ms []Moment, err error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT mo.ID, mo.Latitude, mo.Longitude
			  FROM [moment].[Moments] mo
			  JOIN [moment].[Leaves] l
			    ON mo.ID = l.MomentID
			  WHERE l.RecipientID = ?`

	rows, err := db.Query(query, u)
	if err != nil {
		return
	}
	defer rows.Close()

	m := new(Moment)
	for rows.Next() {
		if err = rows.Scan(&m.ID, &m.Latitude, &m.Longitude); err != nil {
			return
		}
		ms = append(ms, *m)
	}
	if err = rows.Err(); err != nil {
		return
	}

	return
}

// searchHiddenPublic queries for near-by hidden public moments.
func searchHiddenPublic(l Location) (ms []Moment, err error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT ID,
					 Latitude,
					 Longitude
			  FROM [moment].[Moments]
			  WHERE [Public] = 1
			  		AND Hidden = 1
			  		AND Latitude = ?
			  		AND Longitude = ?`
	args := []interface{}{l.Latitude, l.Longitude}

	rows, err := db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	m := new(Moment)
	destAddrs := []interface{}{
		&m.ID,
		&m.Latitude,
		&m.Longitude,
	}

	for rows.Next() {
		if err = rows.Scan(destAddrs...); err != nil {
			return
		}
		ms = append(ms, *m)
	}
	if err = rows.Err(); err != nil {
		return
	}

	return
}

// Moment.leave creates a new moment in Moment-Db.
// The moment content; type, message, and MediaDir are stored.
// The moment is stored.
// If the moment is not public, the leaves are stored.
func (m *Moment) leave() (err error) {
	db := openDbConn()
	defer db.Close()

	insert := `INSERT [moment].[Media] (Type, Message, MediaDir)
					VALUES (?, ?, ?)`
	args := []interface{}{m.Type, m.Message, m.MediaDir}

	t, err := db.Begin()
	if err != nil {
		return
	}

	res, err := t.Exec(insert, args...)
	if err != nil {
		t.Rollback()
		return
	}

	MediaID, err := res.LastInsertId()
	if err != nil {
		return
	}

	insert = `INSERT [moment].[Moments] ([SenderID], [Latitude], [Longitude], [MediaID], [Public], [Hidden], [CreateDate])
					 VALUES (?, ?, ?, ?, ?, ?, ?)`
	args = []interface{}{m.SenderID, m.Latitude, m.Longitude, MediaID, m.Public, m.Hidden, m.CreateDate}

	res, err = t.Exec(insert, args...)
	if err != nil {
		t.Rollback()
		return
	}

	MomentID, err := res.LastInsertId()
	if err != nil {
		return
	}

	if !m.Public {
		insert = `INSERT [moment].[Leaves] (MomentID, RecipientID) 
				  VALUES `

		values, args := getRecipientValues(MomentID, m.RecipientIDs)
		insert = insert + values

		if _, err = t.Exec(insert, args...); err != nil {
			t.Rollback()
			return
		}
	}

	t.Commit()

	return nil
}

// Moment.findLeave updates the corresponding moment resource in the Moment-Db.
// The moment's found flag and found date are updated.
func (m *Moment) findLeave() (err error) {
	db := openDbConn()
	defer db.Close()

	updateLeave := `UPDATE [moment].[Leaves]
					SET Found = 1,
						FindDate = ?
					WHERE RecipientID = ?
						  AND MomentID = ?`

	updateArgs := []interface{}{time.Now().UTC(), m.RecipientID, m.ID}

	res, err := db.Exec(updateLeave, updateArgs...)
	if err != nil {
		return
	}
	if err = validateRowsAffected(res, 1); err != nil {
		return
	}

	return
}

// Moment.findPublic inserts a new record into the Finds table.
// This row consists of the user's UserID and the found MomentID.
func (m *Moment) findPublic() (err error) {
	db := openDbConn()
	defer db.Close()

	insertFind := `INSERT INTO [moment].[Finds]
				   VALUES (?, ?)`
	args := []interface{}{m.ID, m.FinderID}

	res, err := db.Exec(insertFind, args...)
	if err != nil {
		return
	}
	if err = validateRowsAffected(res, 1); err != nil {
		return
	}

	return
}

// // Moment.share updates the corresponding leave resource in the Moment-Db.
// // The leaves' share flag is set to 1.
// // A share is created for each recipient of share in the Shares table.
func (m *Moment) share() (err error) {
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
		return
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
		return
	}

	if err != nil {
		return
	}

	insert := `INSERT INTO [moment].[Shares] (LeaveID, RecipientID)
			   VALUES `

	values, args := getRecipientValues(LeaveID, m.RecipientIDs)
	insert = insert + values

	if _, err = t.Exec(insert, args...); err != nil {
		return
	}

	return
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

// validateDbExecResult compares the number of records modified by the Exec
// to the expected number of records expected to have been modified.
// Function errors on actual and expected not being equal.
func validateRowsAffected(res sql.Result, expected int) (err error) {
	rows, err := res.RowsAffected()
	if err != nil {
		return
	}

	if rows != int64(expected) {
		err = errors.New("db.Exec affected an unpredicted number of rows.")
		return
	}

	return
}
