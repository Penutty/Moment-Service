package moment

import (
	"errors"
	"fmt"
	"github.com/penutty/dba"
	"strconv"
	"strings"
	"time"
)

const (
	minUserChars = 6
	maxUserChars = 64

	minMediaType = 0
	maxMediaType = 3

	minMessage = 0
	maxMessage = 256

	minLat = -180
	maxLat = 180

	minLong = -90
	maxLong = 90

	Datetime2 = "2006-01-02 15:04:05"

	DNE = iota
	Image
	Video
)

var InterfaceTypeNotRecognized = errors.New("The type switch does not recognize the interface type.")
var ConnStrFailed = errors.New("Connection to Moment-Db failed.")

var ErrorNoTransactionNotNil = errors.New("Invalid type passed to function. Expected *dba.Trans or nil.")

func check(err error) {
	if err != nil {
		panic(err)
	}
}

var ErrorMomentID = errors.New("*id must be >= 1.")

func validateMomentID(id int) (err error) {
	if id < 1 {
		err = ErrorMomentID
	}
	return
}

var ErrorMediaTypeDNE = errors.New("*t must be >= " + strconv.Itoa(minMediaType) + " AND <= " + strconv.Itoa(maxMediaType))

func validateMediaType(t uint8) (err error) {

	if t > maxMediaType {
		err = ErrorMediaTypeDNE
	}
	return
}

var (
	ErrorUserIDEmpty = errors.New("*id is empty.")
	ErrorUserIDShort = errors.New("len(*id) must be >= " + strconv.Itoa(minUserChars) + ".")
	ErrorUserIDLong  = errors.New("len(*id) must be <= " + strconv.Itoa(maxUserChars) + ".")
)

func validateUserID(id string) (err error) {
	l := len(id)

	switch {
	case l == 0:
		err = ErrorUserIDEmpty
	case l < minUserChars:
		err = ErrorUserIDShort
	case l > maxUserChars:
		err = ErrorUserIDLong
	}

	return
}

var ErrorMediaMessageLong = errors.New("m must be >= " + strconv.Itoa(minMessage) + " AND <= " + strconv.Itoa(maxMessage) + ".")

func validateMediaMessage(m string) (err error) {
	if len(m) > 256 {
		err = ErrorMediaMessageLong
	}
	return
}

var (
	ErrorNoMediaTypeHasDir = errors.New("Dir must be \"\" for this media type.")
	ErrorMediaTypeNoDir    = errors.New("Dir must be set for this media type.")
)

func validateMediaDir(mType uint8, d string) (err error) {
	switch {
	case mType == 0 && d != "":
		err = ErrorNoMediaTypeHasDir
	case mType > 0 && mType < 4 && d == "":
		err = ErrorMediaTypeNoDir
	}
	return
}

var (
	ErrorFindDateWithFalseFound = errors.New("fd is a valid pointer while f=false.")
	ErrorFindDateDNEWithFound   = errors.New("fd is a nil pointer while f=true.")
)

func validateFindDate(f bool, fd *time.Time) (err error) {
	switch {
	case !f && fd != nil:
		err = ErrorFindDateWithFalseFound
	case f && fd == nil:
		err = ErrorFindDateDNEWithFound
	}
	return
}

var (
	ErrorShareAllPublicWithRecipients = errors.New("r cannot be set when All=true.")
	ErrorShareAllNoRecipients         = errors.New("r must be set when All=false.")
)

func validateShareAll(All bool, r string) (err error) {
	switch {
	case All && len(r) > 0:
		err = ErrorShareAllPublicWithRecipients
	case !All && len(r) == 0:
		err = ErrorShareAllNoRecipients
	}
	return
}

var ErrorLatitude = errors.New("Latitude must be between -180 and 180.")

func validateLatitude(l float32) (err error) {
	if l < -180.00 || l > 180.00 {
		err = ErrorLatitude
	}
	return
}

var ErrorLongitude = errors.New("Longitude must be between -90 and 90.")

func validateLongitude(l float32) (err error) {
	if l < -90.00 || l > 90.00 {
		err = ErrorLongitude
	}
	return
}

var ErrorLocationReference = errors.New("Location reference is nil.")

func validateLocation(l *Location) (err error) {
	if l == nil {
		err = ErrorLocationReference
	}
	return
}

var ErrorPublicHiddenCombination = errors.New("Public=false AND Hidden=true is an Error input combination.")

func validateMomentPublicHidden(p, h bool) (err error) {
	if !p && h {
		err = ErrorPublicHiddenCombination
	}
	return
}

func NewMoment(l *Location, uID string, p bool, h bool, c time.Time) (m *MomentsRow, err error) {
	check(validateLocation(l))
	check(validateUserID(uID))
	check(validateMomentPublicHidden(p, h))

	m = &MomentsRow{
		Location: Location{
			l.latitude,
			l.longitude,
		},
		userID:     uID,
		public:     p,
		hidden:     h,
		createDate: c,
	}

	return

}

func NewMedia(mID int, m string, mType uint8, d string) (mr *MediaRow, err error) {
	check(validateMomentID(mID))
	check(validateMediaMessage(m))
	check(validateMediaType(mType))
	check(validateMediaDir(mType, d))

	mr = &MediaRow{
		momentID: mID,
		message:  m,
		mType:    mType,
		dir:      d,
	}

	return

}

func NewFind(mID int, uID string, f bool, fd *time.Time) (fr *FindsRow, err error) {
	check(validateMomentID(mID))
	check(validateUserID(uID))
	check(validateFindDate(f, fd))

	fr = &FindsRow{
		momentID: mID,
		userID:   uID,
		found:    f,
		findDate: fd,
	}

	return
}

func NewShare(mID int, uID string, All bool, r string) (s *SharesRow, err error) {
	check(validateMomentID(mID))
	check(validateUserID(uID))
	check(validateUserID(r))
	check(validateShareAll(All, r))

	s = &SharesRow{
		momentID:    mID,
		userID:      uID,
		all:         All,
		recipientID: r,
	}

	return
}

func NewLocation(lat float32, long float32) (l *Location, err error) {
	check(validateLatitude(lat))
	check(validateLongitude(long))

	l = &Location{
		latitude:  lat,
		longitude: long,
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

func (m *MediaRow) delete(i interface{}) (rowCnt int, err error) {

	c := new(dba.Trans)
	switch v := i.(type) {
	case nil:
		c = dba.OpenTx()
		defer func() { c.Close(err) }()
	case *dba.Trans:
		c = v
	default:
		err = ErrorNoTransactionNotNil
	}

	deleteFrom := `DELETE FROM [moment].[Media]
				   WHERE MomentID = ?
				   		 AND Type = ?`
	args := []interface{}{m.momentID, m.mType}

	res, err := c.Tx.Exec(deleteFrom, args...)
	if err != nil {
		return
	}

	cnt, err := res.RowsAffected()
	rowCnt = int(cnt)

	return
}

type Media []*MediaRow

func (mSet *Media) insert() (rowCnt int, err error) {
	c := dba.OpenConn()
	defer c.Db.Close()

	query := `INSERT INTO [moment].[Media] (MomentID, Message, Type, Dir)
			  VALUES `
	values := mSet.values()
	query = query + values
	args := mSet.args()

	res, err := c.Db.Exec(query, args...)
	if err != nil {
		return
	}
	cnt, err := res.RowsAffected()
	rowCnt = int(cnt)

	return
}

func (mSet *Media) values() (values string) {
	vSlice := make([]string, len(*mSet))
	for i := 0; i < len(vSlice); i++ {
		vSlice[i] = "(?, ?, ?, ?)"
	}
	values = strings.Join(vSlice, ", ")

	return
}

func (mSet *Media) args() (args []interface{}) {
	fCnt := 4
	argsCnt := len(*mSet) * fCnt
	args = make([]interface{}, argsCnt)
	for i, m := range *mSet {
		j := i * 4
		args[j] = m.momentID
		args[j+1] = m.message
		args[j+2] = m.mType
		args[j+3] = m.dir
	}
	return
}

func (mSet *Media) delete() (rowCnt int, err error) {
	c := dba.OpenTx()
	defer func() { c.Close(err) }()

	var affCnt int
	for _, m := range *mSet {
		affCnt, err = m.delete(c)
		if err != nil {
			rowCnt = 0
			return
		}
		rowCnt += affCnt
	}

	return
}

type FindsRow struct {
	momentID int
	userID   string
	found    bool
	findDate *time.Time
}

func (f FindsRow) String() string {
	return fmt.Sprintf("momentID: %v\n"+
		"userID:   %v\n"+
		"found: 	  %v\n"+
		"findDate: %v\n",
		f.momentID,
		f.userID,
		f.found,
		f.findDate)
}

func (f *FindsRow) find() (err error) {
	c := dba.OpenConn()
	defer c.Db.Close()

	updateFindsRow := `UPDATE [moment].[Finds]
					   SET Found = 1,
					   FindDate = ?
					   WHERE UserID = ?
					  		 AND MomentID = ?`

	args := []interface{}{time.Now().UTC(), f.userID, f.momentID}

	res, err := c.Db.Exec(updateFindsRow, args...)
	if err != nil {
		return
	}
	err = dba.ValidateRowsAffected(res, 1)

	return
}

func (f *FindsRow) delete(c *dba.Trans) (rowsAff int, err error) {

	deleteFrom := `DELETE FROM [moment].[Finds]
				   WHERE MomentID = ?
				   		 AND UserID = ?`
	args := []interface{}{f.momentID, f.userID}

	res, err := c.Tx.Exec(deleteFrom, args...)
	aff, err := res.RowsAffected()
	rowsAff = int(aff)

	return
}

type Finds []*FindsRow

func (fSet *Finds) insert() (rowCnt int, err error) {
	c := dba.OpenConn()
	defer c.Db.Close()

	insert := `INSERT [moment].[Finds] (MomentID, UserID, Found, FindDate)
			   VALUES `
	values := fSet.values()
	insert = insert + values
	args := fSet.args()

	res, err := c.Db.Exec(insert, args...)
	if err != nil {
		return
	}
	cnt, err := res.RowsAffected()
	rowCnt = int(cnt)

	return
}

func (fSet *Finds) values() (values string) {
	vSlice := make([]string, len(*fSet))
	for i := 0; i < len(*fSet); i++ {
		vSlice[i] = "(?, ?, ?, ?)"
	}
	values = strings.Join(vSlice, ", ")

	return
}

func (fSet *Finds) args() (args []interface{}) {
	findFieldCnt := 4
	argsCnt := len(*fSet) * findFieldCnt
	args = make([]interface{}, argsCnt)

	for i, f := range *fSet {
		j := 4 * i
		args[j] = f.momentID
		args[j+1] = f.userID
		args[j+2] = f.found
		args[j+3] = f.findDate
	}

	return
}

func (fSet *Finds) delete() (rowCnt int, err error) {
	c := dba.OpenTx()
	defer func() { c.Close(err) }()

	var rAff int
	for _, f := range *fSet {
		rAff, err = f.delete(c)
		if err != nil {
			rowCnt = 0
			return
		}
		rowCnt += rAff
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

func (s *SharesRow) delete(i interface{}) (affCnt int, err error) {

	c := new(dba.Trans)
	switch v := i.(type) {
	case nil:
		c = dba.OpenTx()
		defer func() { c.Close(err) }()
	case *dba.Trans:
		c = v
	default:
		err = ErrorNoTransactionNotNil
	}

	deleteFrom := `DELETE FROM [moment].[Shares]
				   WHERE MomentID = ?
				   		 AND UserID = ?
				   		 AND RecipientID = ?`
	args := []interface{}{s.momentID, s.userID, s.recipientID}

	res, err := c.Tx.Exec(deleteFrom, args...)
	if err != nil {
		return
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return
	}
	affCnt = int(cnt)

	return
}

type Shares []*SharesRow

func (sSlice *Shares) insert() (rowCnt int, err error) {
	c := dba.OpenConn()
	defer c.Db.Close()

	insert := `INSERT INTO [moment].[Shares] (MomentID, UserID, [All], RecipientID)
			   VALUES `
	values := sSlice.values()
	insert = insert + values
	args := sSlice.args()

	res, err := c.Db.Exec(insert, args...)
	if err != nil {
		return
	}
	cnt, err := res.RowsAffected()
	if err != nil {
		return
	}
	rowCnt = int(cnt)

	return
}

func (sSlice *Shares) values() (values string) {
	valuesSlice := make([]string, len(*sSlice))
	for i := 0; i < len(valuesSlice); i++ {
		valuesSlice[i] = "(?, ?, ?, ?)"
	}
	values = strings.Join(valuesSlice, ", ")
	return
}

func (sSlice *Shares) args() (args []interface{}) {
	SharesFieldCnt := 4
	argsCnt := len(*sSlice) * SharesFieldCnt
	args = make([]interface{}, argsCnt)

	for i, s := range *sSlice {
		j := i * 4
		args[j] = s.momentID
		args[j+1] = s.userID
		args[j+2] = s.all
		args[j+3] = s.recipientID
	}

	return
}

func (sSlice *Shares) delete() (rowCnt int, err error) {
	c := dba.OpenTx()
	defer func() { c.Close(err) }()

	var affCnt int
	for _, s := range *sSlice {
		affCnt, err = s.delete(c)
		if err != nil {
			rowCnt = 0
			return
		}
		rowCnt += affCnt
	}
	return
}

// Location is a geographical point identified by longitude, latitude.
type Location struct {
	latitude  float32
	longitude float32
}

func (l Location) String() string {
	return fmt.Sprintf("latitude: %v\nlongitude: %v\n", l.latitude, l.longitude)
}

func (l *Location) balloon() (lRange []interface{}) {
	lRange = []interface{}{
		l.latitude - 3,
		l.latitude + 3,
		l.longitude - 3,
		l.longitude + 3,
	}

	return
}

// Moment is the main resource of this package.
// It is a grouping of the Content and Location structs.
type MomentsRow struct {
	Location
	id         int
	userID     string
	public     bool
	hidden     bool
	createDate time.Time
}

func (m MomentsRow) String() string {
	return fmt.Sprintf("id: %v\n"+
		"userID: %v\n"+
		"Location: %v\n"+
		"public: %v\n"+
		"hidden: %v\n"+
		"createDate: %v\n",
		m.id,
		m.userID,
		m.Location,
		m.public,
		m.hidden,
		m.createDate)
}

// Moment.leave creates a new moment in Moment-Db.
// The moment content; type, message, and MediaDir are stored.
// The moment is stored.
// If the moment is not public, the leaves are stored.
func (m *MomentsRow) insert() (momentID int, err error) {
	c := dba.OpenConn()
	defer c.Db.Close()

	insert := `INSERT [moment].[Moments] ([UserID], [Latitude], [Longitude], [Public], [Hidden], [CreateDate])
			   VALUES `
	values := m.values()
	insert = insert + values
	args := m.args()

	res, err := c.Db.Exec(insert, args...)
	if err != nil {
		return
	}

	if err = dba.ValidateRowsAffected(res, 1); err != nil {
		return
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return
	}
	momentID = int(lastID)

	return
}

func (m *MomentsRow) values() string {
	return "(?, ?, ?, ?, ?, ?)"
}

func (m *MomentsRow) args() []interface{} {
	return []interface{}{m.userID, m.latitude, m.longitude, m.public, m.hidden, m.createDate}
}

func (m *MomentsRow) delete() (err error) {
	c := dba.OpenConn()
	defer c.Db.Close()

	deleteFrom := `DELETE FROM [moment].[Moments]
				   WHERE ID = ?`

	_, err = c.Db.Exec(deleteFrom, m.id)

	return
}

// // searchPublic queries Moment-Db for moments that are public and not hidden.
// func searchPublic(l Location) (ms []MomentsRow, err error) {
// 	db := openDbConn()
// 	defer db.Close()

// 	query := `SELECT mo.ID,
// 					 mo.UserID,
// 					 mo.Latitude,
// 					 mo.Longitude,
// 					 m.Type,
// 					 m.Message,
// 					 m.MediaDir,
// 					 mo.CreateDate
// 			  FROM [moment].[Moments] mo
// 			  JOIN [moment].[Media] m
// 			    ON mo.ID = m.MomentID
// 			  WHERE mo.Hidden = 0
// 			  		AND mo.[Public] = 1
// 			  		AND mo.Latitude BETWEEN ? AND ?
// 			  		AND mo.Longitude BETWEEN ? AND ?`

// 	lRange := l.balloon()
// 	rows, err := db.Query(query, lRange...)
// 	if err != nil {
// 		return
// 	}
// 	defer rows.Close()

// 	m := new(MomentsRow)
// 	var createDate string
// 	fieldAddrs := []interface{}{
// 		&m.id,
// 		&m.userID,
// 		&m.latitude,
// 		&m.longitude,
// 		&m.Type,
// 		&m.Message,
// 		&m.MediaDir,
// 		&createDate,
// 	}

// 	for rows.Next() {
// 		if err = rows.Scan(fieldAddrs...); err != nil {
// 			return
// 		}
// 		m.CreateDate, err = time.Parse(Datetime2, createDate)
// 		if err != nil {
// 			return
// 		}
// 		ms = append(ms, *m)
// 	}
// 	if err = rows.Err(); err != nil {
// 		return
// 	}

// 	return
// }

// // searchShared queries Moment-Db for moments that a user has found, and shared with others.
// func searchShared(u string, me string) (ms []MomentsRow, err error) {
// 	db := openDbConn()
// 	defer db.Close()

// 	query := `SELECT mo.ID,
// 					 mo.UserID,
// 					 mo.Latitude,
// 					 mo.Longitude,
// 					 mo.CreateDate,
// 					 m.Type,
// 					 m.Message,
// 					 m.MediaDir
// 			  FROM [moment].[Moments] mo
// 			  JOIN [moment].[Media] m
// 			    ON mo.ID = m.MomentID
// 			  JOIN [moment].[Finds] f
// 			    ON mo.ID = f.MomentID
// 			  JOIN [moment].[Shares] s
// 			  	ON mo.ID = s.MomentID
// 			  WHERE f.UserID = ?
// 			  		AND (s.All = 1 OR s.RecipientID = ?)`

// 	rows, err := db.Query(query, u, me)
// 	if err != nil {
// 		return
// 	}
// 	defer rows.Close()

// 	m := new(MomentsRow)
// 	var createDate string
// 	fieldAddrs := []interface{}{
// 		&m.ID,
// 		&m.UserID,
// 		&m.latitude,
// 		&m.longitude,
// 		&createDate,
// 		&m.Type,
// 		&m.Message,
// 		&m.MediaDir,
// 	}

// 	for rows.Next() {
// 		if err = rows.Scan(fieldAddrs...); err != nil {
// 			return
// 		}
// 		m.CreateDate, err = time.Parse(Datetime2, createDate)
// 		if err != nil {
// 			return
// 		}
// 		ms = append(ms, *m)
// 	}
// 	if err = rows.Err(); err != nil {
// 		return
// 	}

// 	return
// }

// // searchFound queries Moment-Db for moments that a user has been left and has found.
// func searchFound(u string) (ms []MomentsRow, err error) {
// 	db := openDbConn()
// 	defer db.Close()

// 	query := `SELECT mo.ID,
// 					 mo.UserID,
// 					 mo.Latitude,
// 					 mo.Longitude,
// 					 mo.CreateDate,
// 					 f.FindDate,
// 					 m.Type,
// 					 m.Message,
// 					 m.MediaDir
// 			  FROM [moment].[Moments] mo
// 			  JOIN [moment].[Media] m
// 			    ON mo.ID = m.MomentID
// 			  JOIN [moment].[Finds] f
// 			    ON mo.ID = f.MomentID
// 			  WHERE f.UserID = ?
// 			  		AND f.Found = 1`

// 	rows, err := db.Query(query, u)
// 	if err != nil {
// 		return
// 	}
// 	defer rows.Close()

// 	m := new(MomentsRow)
// 	f := make([]*FindsRow, 1)
// 	var createDate string
// 	var findDate string
// 	fieldAddrs := []interface{}{
// 		&m.ID,
// 		&m.UserID,
// 		&m.latitude,
// 		&m.longitude,
// 		&createDate,
// 		&findDate,
// 		&m.Type,
// 		&m.Message,
// 		&m.MediaDir,
// 	}
// 	for rows.Next() {
// 		if err = rows.Scan(fieldAddrs...); err != nil {
// 			return
// 		}
// 		m.CreateDate, err = time.Parse(Datetime2, createDate)
// 		if err != nil {
// 			return
// 		}

// 		f[0].FindDate, err = time.Parse(Datetime2, findDate)
// 		if err != nil {
// 			return
// 		}
// 		m.Finds = f
// 		ms = append(ms, *m)
// 	}
// 	if err = rows.Err(); err != nil {
// 		return
// 	}

// 	return
// }

// // searchLeft queries Moment-Db for moments a user has left for others to find.
// func searchLeft(u string) (ms []MomentsRow, err error) {
// 	db := openDbConn()
// 	defer db.Close()

// 	query := `SELECT mo.ID,
// 					 mo.Latitude,
// 					 mo.longitude,
// 					 mo.CreateDate,
// 					 m.Type,
// 					 m.Message,
// 					 m.MediaDir,
// 					 f.UserID,
// 					 f.Found,
// 					 f.FindDate
// 			  FROM [moment].[Moments] mo
// 			  JOIN [moment].[Media] m
// 			    ON mo.ID = m.MomentID
// 			  JOIN [moment].[Finds] f
// 			    ON mo.ID = f.MomentID
// 			  WHERE mo.UserID = ?`

// 	rows, err := db.Query(query, u)
// 	if err != nil {
// 		return
// 	}
// 	defer rows.Close()

// 	m := new(MomentsRow)
// 	f := new(FindsRow)
// 	var createDate string
// 	addrs := []interface{}{
// 		&m.ID,
// 		&m.latitude,
// 		&m.longitude,
// 		&createDate,
// 		&m.Type,
// 		&m.Message,
// 		&m.MediaDir,
// 		&f.UserID,
// 		&f.Found,
// 		&f.FindDate,
// 	}

// 	var prevID int
// 	for rows.Next() {
// 		if err = rows.Scan(addrs...); err != nil {
// 			return
// 		}

// 		m.Finds = append(m.Finds, f)

// 		if m.ID != prevID {
// 			m.CreateDate, err = time.Parse(Datetime2, createDate)
// 			if err != nil {
// 				return
// 			}

// 			ms = append(ms, *m)
// 		}
// 	}
// 	if err = rows.Err(); err != nil {
// 		return
// 	}

// 	return
// }

// // searchLost queries Moment-Db for moments others have left for the specified user to find.
// func searchLost(u string, l Location) (ms []MomentsRow, err error) {
// 	db := openDbConn()
// 	defer db.Close()

// 	query := `SELECT mo.ID, mo.latitude, mo.longitude
// 			  FROM [moment].[Moments] mo
// 			  JOIN [moment].[Finds] f
// 			    ON mo.ID = f.MomentID
// 			  WHERE ((f.UserID = ? AND f.Found = 0)
// 			  		OR (mo.Hidden = 1))
// 			  		AND mo.Latitude BETWEEN ? AND ?
// 			  		AND mo.longitude BETWEEN ? AND ?`

// 	lRange := l.balloon()
// 	if err != nil {
// 		return
// 	}

// 	args := []interface{}{u}
// 	args = append(args, lRange...)

// 	rows, err := db.Query(query, args...)
// 	if err != nil {
// 		return
// 	}
// 	defer rows.Close()

// 	m := new(MomentsRow)
// 	for rows.Next() {
// 		if err = rows.Scan(&m.ID, &m.latitude, &m.Longitude); err != nil {
// 			return
// 		}
// 		ms = append(ms, *m)
// 	}
// 	if err = rows.Err(); err != nil {
// 		return
// 	}

// 	return
// }

//
// The functions below are utility functions and are only used in this package.
// There functionality has been abstracted and from the above functions for the
// sake of simplicity and readability.
//

// validateDbExecResult compares the number of records modified by the Exec
// to the expected number of records expected to have been modified.
// Function errors on actual and expected not being equal.
