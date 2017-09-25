package moment

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/minus5/gofreetds"
	"log"
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

// Location is a geographical point identified by longitude, latitude.
type Location struct {
	Latitude  float32
	Longitude float32
}

func (l *Location) String() string {
	return fmt.Sprintf("Latitude: %v\nLongitude: %v\n", l.Latitude, l.Longitude)
}

func (l *Location) balloon() (lRange []interface{}) {
	lRange = []interface{}{
		l.Latitude - 3,
		l.Latitude + 3,
		l.Longitude - 3,
		l.Longitude + 3,
	}

	return
}

// Content is a set resources that belong to a Moment.
// Content may be a message, image, and/or a video.
type Content struct {
	MomentID int
	Type     uint8
	Message  string
	MediaDir string
}

func (c *Content) String() string {
	return fmt.Sprintf("Type: %v\nMessage: %v\nMediaDir: %v\n", c.Type, c.Message, c.MediaDir)
}

func (c *Content) create(t *sql.Tx) (err error) {
	query := `INSERT INTO [moment].[Media] (MomentID, Message, Type, MediaDir)
			  VALUES (?, ?, ?, ?)`
	args := []interface{}{c.MomentID, c.Message, c.Type, c.MediaDir}

	_, err = t.Exec(query, args...)
	if err != nil {
		return
	}

	return
}

type FindsRow struct {
	MomentID int
	UserID   string
	Found    bool
	FindDate time.Time
	Shares   *[]SharesRow
}

func (f *FindsRow) String() string {
	return fmt.Sprintf(`MomentID: %v
						UserID: %v
						Found: %v
						FindDate: %v
						Shares: %v`,
		f.MomentID,
		f.UserID,
		f.Found,
		f.FindDate,
		f.Shares)
}

func (f *FindsRow) create(i interface{}) (err error) {

	insert := `INSERT [moment].[Finds] (MomentID, UserID, Found) 
			  VALUES (?, ?, ?)`
	args := []interface{}{f.MomentID, f.UserID, f.Found}

	switch v := i.(type) {
	case *sql.DB:
		_, err = v.Exec(insert, args...)
	case *sql.Tx:
		_, err = v.Exec(insert, args...)
	default:
		return errors.New("No DB or Tx Connection was passed to FindsRow.Create().")
	}

	return
}

func (f *FindsRow) find(MomentID string) (err error) {
	db := openDbConn()
	defer db.Close()

	updateFindsRow := `UPDATE [moment].[Finds]
					   SET Found = 1,
					   FindDate = ?
					   WHERE UserID = ?
					  		 AND MomentID = ?`

	args := []interface{}{time.Now().UTC(), f.UserID, f.MomentID}

	res, err := db.Exec(updateFindsRow, args...)
	if err != nil {
		return
	}
	if err = validateRowsAffected(res, 1); err != nil {
		return
	}

	return
}

func (f *FindsRow) share() (err error) {
	for _, s := range *f.Shares {
		if err = s.create(); err != nil {
			return
		}
	}

	return
}

type SharesRow struct {
	MomentID    int
	UserID      string
	All         bool
	RecipientID string
}

func (s *SharesRow) String() string {
	return fmt.Sprintf(`MomentID: %v
						UserID: %v
						All: %v
						RecipientID: %v`,
		s.MomentID,
		s.UserID,
		s.All,
		s.RecipientID)
}

func (s *SharesRow) create() (err error) {
	db := openDbConn()
	defer db.Close()

	insert := `INSERT INTO [moment].[Shares] (MomentID, UserID, All, RecipientID)
			   VALUES (?, ?, ?, ?)`
	args := []interface{}{s.MomentID, s.UserID, s.All, s.RecipientID}

	if _, err = db.Exec(insert, args...); err != nil {
		return
	}

	return
}

// Moment is the main resource of this package.
// It is a grouping of the Content and Location structs.
type MomentsRow struct {
	Location
	Content
	ID         int
	UserID     string
	Finds      *[]FindsRow
	Public     bool
	Hidden     bool
	CreateDate time.Time
}

func (m *MomentsRow) String() string {
	return fmt.Sprintf(`ID: %v
						UserID: %v
						Location: %v
						Content: %v
						Public: %v
						Hidden: %v
						CreateDate: %v
						Finds: %v`,
		m.ID,
		m.UserID,
		m.Location,
		m.Content,
		m.Public,
		m.Hidden,
		m.CreateDate,
		m.Finds)
}

// Moment.leave creates a new moment in Moment-Db.
// The moment content; type, message, and MediaDir are stored.
// The moment is stored.
// If the moment is not public, the leaves are stored.
func (m *MomentsRow) create() (err error) {
	db := openDbConn()
	defer db.Close()

	insert := `INSERT [moment].[Moments] ([UserID], [Latitude], [Longitude], [Public], [Hidden], [CreateDate])
					 VALUES (?, ?, ?, ?, ?, ?, ?)`
	args := []interface{}{m.UserID, m.Latitude, m.Longitude, m.Public, m.Hidden, m.CreateDate}

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

	res, err := t.Exec(insert, args...)
	if err != nil {
		return
	}

	if err = validateRowsAffected(res, 1); err != nil {
		return
	}

	if err = m.Content.create(t); err != nil {
		return
	}

	if !m.Public {
		for _, f := range *m.Finds {
			if err = f.create(t); err != nil {
				return
			}
		}
	}

	return
}

// Moment.find inserts a row.
func (m *MomentsRow) find(u string) (err error) {
	db := openDbConn()
	defer db.Close()

	f := &FindsRow{
		MomentID: m.ID,
		UserID:   u,
		Found:    true,
	}

	if err = f.create(db); err != nil {
		return
	}

	return
}

// searchPublic queries Moment-Db for moments that are public and not hidden.
func searchPublic(l Location) (ms []MomentsRow, err error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT mo.ID,
					 mo.UserID,
					 mo.Latitude, 
					 mo.Longitude,
					 m.Type,
					 m.Message,
					 m.MediaDir,
					 mo.CreateDate
			  FROM [moment].[Moments] mo
			  JOIN [moment].[Media] m
			    ON mo.ID = m.MomentID
			  WHERE mo.Hidden = 0
			  		AND mo.[Public] = 1
			  		AND mo.Latitude BETWEEN ? AND ?
			  		AND mo.Longitude BETWEEN ? AND ?`

	lRange := l.balloon()
	rows, err := db.Query(query, lRange...)
	if err != nil {
		return
	}
	defer rows.Close()

	m := new(MomentsRow)
	var createDate string
	fieldAddrs := []interface{}{
		&m.ID,
		&m.UserID,
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
func searchShared(u string, me string) (ms []MomentsRow, err error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT mo.ID, 
					 mo.UserID, 
					 mo.Latitude,
					 mo.Longitude,
					 mo.CreateDate,
					 m.Type,
					 m.Message,
					 m.MediaDir
			  FROM [moment].[Moments] mo
			  JOIN [moment].[Media] m
			    ON mo.ID = m.MomentID
			  JOIN [moment].[Finds] f
			    ON mo.ID = f.MomentID
			  JOIN [moment].[Shares] s
			  	ON mo.ID = s.MomentID
			  WHERE f.UserID = ?
			  		AND (s.All = 1 OR s.RecipientID = ?)`

	rows, err := db.Query(query, u, me)
	if err != nil {
		return
	}
	defer rows.Close()

	m := new(MomentsRow)
	var createDate string
	fieldAddrs := []interface{}{
		&m.ID,
		&m.UserID,
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
func searchFound(u string) (ms []MomentsRow, err error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT mo.ID,
					 mo.UserID,
					 mo.Latitude,
					 mo.Longitude,
					 mo.CreateDate,
					 f.FindDate,
					 m.Type,
					 m.Message,
					 m.MediaDir
			  FROM [moment].[Moments] mo
			  JOIN [moment].[Media] m
			    ON mo.ID = m.MomentID
			  JOIN [moment].[Finds] f
			    ON mo.ID = f.MomentID
			  WHERE f.UserID = ?
			  		AND f.Found = 1`

	rows, err := db.Query(query, u)
	if err != nil {
		return
	}
	defer rows.Close()

	m := new(MomentsRow)
	f := make([]FindsRow, 1)
	var createDate string
	var findDate string
	fieldAddrs := []interface{}{
		&m.ID,
		&m.UserID,
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

		f[0].FindDate, err = time.Parse(Datetime2, findDate)
		if err != nil {
			return
		}
		m.Finds = &f
		ms = append(ms, *m)
	}
	if err = rows.Err(); err != nil {
		return
	}

	return
}

// searchLeft queries Moment-Db for moments a user has left for others to find.
func searchLeft(u string) (ms []MomentsRow, err error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT mo.ID,
					 mo.Latitude,
					 mo.Longitude,
					 mo.CreateDate,
					 m.Type,
					 m.Message,
					 m.MediaDir,
					 f.UserID, 
					 f.Found,
					 f.FindDate
			  FROM [moment].[Moments] mo
			  JOIN [moment].[Media] m
			    ON mo.ID = m.MomentID
			  JOIN [moment].[Finds] f
			    ON mo.ID = f.MomentID
			  WHERE mo.UserID = ?`

	rows, err := db.Query(query, u)
	if err != nil {
		return
	}
	defer rows.Close()

	m := new(MomentsRow)
	f := new(FindsRow)
	var createDate string
	addrs := []interface{}{
		&m.ID,
		&m.Latitude,
		&m.Longitude,
		&createDate,
		&m.Type,
		&m.Message,
		&m.MediaDir,
		&f.UserID,
		&f.Found,
		&f.FindDate,
	}

	var prevID int
	for rows.Next() {
		if err = rows.Scan(addrs...); err != nil {
			return
		}

		*m.Finds = append(*m.Finds, *f)

		if m.ID != prevID {
			m.CreateDate, err = time.Parse(Datetime2, createDate)
			if err != nil {
				return
			}

			ms = append(ms, *m)
		}
	}
	if err = rows.Err(); err != nil {
		return
	}

	return
}

// searchLost queries Moment-Db for moments others have left for the specified user to find.
func searchLost(u string, l Location) (ms []MomentsRow, err error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT mo.ID, mo.Latitude, mo.Longitude
			  FROM [moment].[Moments] mo
			  JOIN [moment].[Finds] f
			    ON mo.ID = f.MomentID
			  WHERE ((f.UserID = ? AND f.Found = 0)
			  		OR (mo.Hidden = 1))
			  		AND mo.Latitude BETWEEN ? AND ?
			  		AND mo.Longitude BETWEEN ? AND ?`

	lRange := l.balloon()
	if err != nil {
		return
	}

	args := []interface{}{u}
	args = append(args, lRange...)

	rows, err := db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	m := new(MomentsRow)
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
