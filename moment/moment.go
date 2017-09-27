package moment

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/minus5/gofreetds"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	minUserChars = 1
	maxUserChars = 64

	Datetime2 = "2006-01-02 15:04:05"

	DNE = iota
	Image
	Video
)

var InterfaceTypeNotRecognized = errors.New("The type switch does not recognize the interface type.")
var ConnStrFailed = errors.New("Connection to Moment-Db failed.")

var InvalidMomentID = errors.New("MomentID must be greater than 0.")
var InvalidUserID = errors.New("UserID must be between 1 AND 64 characters long.")
var InvalidRecipients = errors.New("RecipientIDs cannot be set when All=true.")
var InvalidLocationReference = errors.New("Location reference is nil.")
var InvalidPublicHiddenCombination = errors.New("Public=false AND Hidden=true is an invalid input combination.")
var InvalidFindFindDateCombination = errors.New("FindDate may only be set when Found=true.")
var InvalidMediaMessage = errors.New("Message must be between 0 and 256 characters.")
var InvalidMediaType = errors.New("Type must be between 0 and 3")
var InvalidMediaTypeDirCombination = errors.New("Dir must be \"\" when type=0.")

// Location is a geographical point identified by longitude, latitude.
type Location struct {
	Latitude  float32
	Longitude float32
}

func (l Location) String() string {
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

func NewMoment(l *Location, uID string, p bool, h bool, c time.Time) (m *MomentsRow, err error) {
	switch {
	case l == nil:
		err = InvalidLocationReference
	case len(uID) < minUserChars || len(uID) > maxUserChars:
		err = InvalidRecipients
	case !p && h:
		err = InvalidPublicHiddenCombination
	}

	m = &MomentsRow{
		location:   l,
		userID:     uID,
		public:     p,
		hidden:     h,
		createDate: c,
	}

	return

}

func NewMedia(mID int, m string, mType uint8, d string) (mr *MediaRow, err error) {
	switch {
	case mID < 0:
		err = InvalidMomentID
	case len(m) < 0 || len(m) > 256:
		err = InvalidMediaMessage
	case mType < minMediaType || mType > maxMediaType:
		err = InvalidMediaType
	case mType == 0 && d != "":
		err = InvalidMediaTypeDirCombination
	}

	mr = &MediaRow{
		momentID: mID,
		message:  m,
		mType:    mType,
		dir:      d,
	}

	return

}

func NewFind(mID int, uID, string, f bool, fd *time.Time) (fr *FindsRow, err error) {
	switch {
	case mID < 1:
		err = InvalidMomentID
	case len(uID) < minUserChars || len(uID) > maxUserChars:
		err = InvalidUserID
	case !(f && fd != nil):
		err = InvalidFindFindDateCombination
	}

	fr = &FindsRow{
		momentID: mID,
		userID:   uID,
		found:    f,
		findDate: fd,
	}

	return
}

func NewShare(mID int, uID string, All bool, r string) (s *sharesRow, err error) {
	switch {
	case mID > 0:
		err = InvalidMomentID
	case len(uID) < minUserChars || len(uID) > maxUserChars:
		err = InvalidUserID
	case len(r) < minUserChars || len(r) > maxUserChars:
		err = InvalidUserID
	case All && len(r) > 0:
		err = InvalidRecipients
	}

	s = &sharesRow{
		momentID:    mID,
		userID:      uID,
		all:         All,
		recipientID: r,
	}

	return
}

type Set interface {
	insert()
	values()
	args()
}

// Content is a set resources that belong to a Moment.
// Content may be a message, image, and/or a video.
type MediaRow struct {
	momentID int
	message  string
	mType    uint8
	dir      string
}

func (m MediaRow) String() string {
	return fmt.Sprintf("momentID: %v\nmType: %v\nmessage: %v\ndir: %v\n", m.momentID, m.mType, m.message, m.dir)
}

func (m *MediaRow) delete() (err error) {
	db := openDbConn()
	defer db.Close()

	deleteFrom := `DELETE FROM [moment].[Media]
				   WHERE MomentID = ?
				   		 AND Type = ?`
	args := []interface{}{m.momentID, m.mType}

	_, err = db.Exec(deleteFrom, args...)

	return
}

type Media []*MediaRow

func (mSet *Media) insert() (err error) {
	db := openDbConn()
	defer db.Close()

	query := `INSERT INTO [moment].[Media] (MomentID, Message, Type, Dir)
			  VALUES `
	values := mSet.values()
	query = query + values
	args := mSet.args()

	_, err = t.Exec(query, args...)

	return
}

func (mSet *Media) values() (values string) {
	vSlice := make([]string, len(mSet))
	for i := 0; i < len(vSlice); i++ {
		vSlice[i] = "(?, ?, ?, ?)"
	}
	values = strings.Join(vSlice, ", ")

	return
}

func (mSet *Media) args() (args []interface{}) {
	fCnt = 4
	argsCnt = len(mSet) * fCnt
	args = make([]interface{}, argsCnt)
	for i := 0; i < argsCnt; i = i + 4 {
		args[i] = mSet[i].momentID
		args[i+1] = mSet[i].message
		args[i+2] = mSet[i].mType
		args[i+3] = mSet[i].dir
	}
	return
}

type FindsRow struct {
	momentID int
	userID   string
	found    bool
	findDate time.Time
}

func (f FindsRow) String() string {
	return fmt.Sprintf("momentID: %v\n"+
		"userID:   %v\n"+
		"found: 	  %v\n"+
		"findDate: %v\n",
		f.momentID,
		f.userID,
		f.Found,
		f.FindDate)
}

func (f *FindsRow) find() (err error) {
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
	err = validateRowsAffected(res, 1)

	return
}

func (f *FindsRow) delete() (err error) {
	db := openDbConn()
	defer db.Close()

	deleteFrom := `DELETE FROM [moment].[Finds]
				   WHERE MomentID = ?
				   		 AND UserID = ?`
	args := []interface{}{f.momentID, f.userID}

	_, err = db.Exec(deleteFrom, args...)

	return
}

type Finds []*FindsRow

func (fSet *Finds) insert(i interface{}) (err error) {
	db := openDbConn()
	defer db.Close()

	insert := `INSERT [moment].[Finds] (MomentID, UserID, Found, FindDate)
			   VALUES `
	values := fSet.values()
	insert = insert + values
	args := fSet.args()

	if _, err := db.Exec(insert, args...); err != nil {
		return
	}

	return
}

func (fSet *Finds) values() (values string) {
	vSlice := make([]*string, len(fSet))
	for i := 0; i < len(fSet); i++ {
		vSlice[i] = "(?, ?, ?, ?)"
	}
	values = strings.Join(vSlice, ", ")

	return
}

func (fSet *Finds) args() (args []interface{}) {
	findFieldCnt := 4
	argsCnt := len(fSet) * findFieldCnt
	args = make([]interface{}, argsCnt)
	for i := 0; i < argsCnt; i = i + 4 {
		args[i] = fSet[i].momentID
		args[i+1] = fSet[i].userID
		args[i+2] = fSet[i].found
		args[i+3] = fSet[i].findDate
	}

	return
}

type SharesRow struct {
	momentID    int
	userID      string
	all         bool
	recipientID string
}

func (s SharesRow) String() string {

	return fmt.Sprintf("momentID: %v\n"+
		"userID: %v\n"+
		"all: %v\n"+
		"recipientID: %v\n",
		s.momentID,
		s.userID,
		s.all,
		s.recipientID)
}

func (s *SharesRow) delete() (err error) {
	db := openDbConn()
	defer db.Close()

	deleteFrom := `DELETE FROM [moment].[Shares]
				   WHERE MomentID = ?
				   		 AND UserID = ?
				   		 AND RecipientID = ?`
	args := []interface{}{s.momentID, s.userID, s.recipientID}

	_, err = db.Exec(deleteFrom, args...)

	return
}

type Shares []*SharesRow

func (sSlice *Shares) insert() (err error) {

	db := openDbConn()
	defer db.Close()

	insert := `INSERT INTO [moment].[Shares] (MomentID, UserID, [All], RecipientID)
			   VALUES `
	values := sSlice.values()
	insert = insert + values
	args := sSlice.args()

	if _, err = db.Exec(insert, args...); err != nil {
		return
	}

	return
}

func (sSlice *Shares) values() (values string) {
	valuesSlice := make([]*string, len(sSlice))
	for _, v := range valuesSlice {
		v = "(?, ?, ?, ?)"
	}
	values = strings.Join(valuesSlice, ", ")
	return
}

func (sSlice *Shares) args() (args []interface{}) {
	SharesFieldCnt := 4
	argsCnt := len(sSlice) * SharesFieldCnt
	args = make([]interface{}, argsCnt)
	for i := 0; i < args; i = i + 4 {
		args[i] = sSlice[i].momentID
		args[i+1] = sSlice[i].userID
		args[i+2] = sSlice[i].all
		args[i+3] = sSlice[i].recipientID
	}

	return
}

// Moment is the main resource of this package.
// It is a grouping of the Content and Location structs.
type MomentsRow struct {
	location
	id         int
	userID     string
	public     bool
	hidden     bool
	createDate time.Time
}

func (m MomentsRow) String() string {
	return fmt.Sprintf("id: %v\n"+
		"userID: %v\n"+
		"location: %v\n"+
		"content: %v\n"+
		"public: %v\n"+
		"hidden: %v\n"+
		"createDate: %v\n"+
		"finds: %v\n",
		m.id,
		m.userID,
		m.location,
		m.content,
		m.public,
		m.hidden,
		m.createDate,
		m.finds)
}

// Moment.leave creates a new moment in Moment-Db.
// The moment content; type, message, and MediaDir are stored.
// The moment is stored.
// If the moment is not public, the leaves are stored.
func (m *MomentsRow) insert() (momentID int, err error) {
	db := openDbConn()
	defer db.Close()

	insert := `INSERT [moment].[Moments] ([UserID], [Latitude], [Longitude], [Public], [Hidden], [CreateDate])
			   VALUES `
	values := m.values()
	insert = insert + values
	args := m.args()

	res, err := db.Exec(insert, args...)
	if err != nil {
		return
	}

	if err = validateRowsAffected(res, 1); err != nil {
		return
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return
	}
	MomentID := int(lastID)

	return
}

func (m *MomentsRow) values() string {
	return "(?, ?, ?, ?, ?, ?)"
}

func (m *MomentsRow) args() []interface{} {
	return []interface{}{m.UserID, m.Latitude, m.Longitude, m.Public, m.Hidden, m.CreateDate}
}

func (m *MomentsRow) delete() (err error) {
	db := openDbConn()
	defer db.Close()

	deleteFrom := `DELETE FROM [moment].[Moments]
				   WHERE ID = ?`

	res, err := db.Exec(deleteFrom, m.ID)
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
	f := make([]*FindsRow, 1)
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
		m.Finds = f
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

		m.Finds = append(m.Finds, f)

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

const EmptyString = ""
const EmptyInt = 0
const EmptyFloat32 = 0.00

func isset(iSlice ...[]interface{}) (err error) {
	for j, i := range iSlice {
		switch v := i.(type) {
		case string:
			if v == EmptyString {
				return errors.New("Argument " + j + " passed in an empty string.")
			}
		case int:
			if v == EmptyInt {
				return errors.New("Argument " + j + " passed in an empty int.")
			}
		case float32:
			if v == EmptyFloat32 {
				return errors.New("Argument " + j + " passed in an empty int.")
			}
		}
	}

	return
}
