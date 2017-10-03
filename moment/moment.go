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
	// DNE, Image, and Video represent possible values stored in the [moment].[Media].[Type] column.
	DNE = iota
	Image
	Video

	// minUserChars and maxUserChars represent the max and min lengths of userDs and recipientIDs.
	minUserChars = 6
	maxUserChars = 64

	// minMediaType and maxMediaType represent the max and min values of the [moment].[Media].[Type] column.
	minMediaType = 0
	maxMediaType = 3

	// minMessage and maxMessage represent the max and min lengths of the [moment].[Media].[Message].
	minMessage = 0
	maxMessage = 256

	// minLat and maxLat represent the max and min values of the [moment].[Moments].[Latitude] column.
	minLat = -180
	maxLat = 180

	// minLong and maxLong represents the max and min values of the [moment].[Moments].[Longitude] column.
	minLong = -90
	maxLong = 90

	// Datetime2 is the time.Time format this package uses to communicate DateTime2 values to Moment-Db.
	Datetime2 = "2006-01-02 15:04:05"
)

// check is a helper function that is used to ensure another function did not return an error.
func check(err error) {
	if err != nil {
		panic(err)
	}
}

var ErrorMomentID = errors.New("*id must be >= 1.")

// validateMomentID ensures id is not less than 1.
func validateMomentID(id int) (err error) {
	if id < 1 {
		err = ErrorMomentID
	}
	return
}

var ErrorMediaTypeDNE = errors.New("*t must be >= " + strconv.Itoa(minMediaType) + " AND <= " + strconv.Itoa(maxMediaType))

// validateMediaType ensures that t is a value between minMediaType and maxMediaType.
func validateMediaType(t uint8) (err error) {

	if t > maxMediaType {
		err = ErrorMediaTypeDNE
	}
	return
}

var (
	ErrorUserIDEmpty = errors.New("*id (userID) is empty.")
	ErrorUserIDShort = errors.New("len(*id) (userID) must be >= " + strconv.Itoa(minUserChars) + ".")
	ErrorUserIDLong  = errors.New("len(*id) (userID) must be <= " + strconv.Itoa(maxUserChars) + ".")
)

// validateUserID ensures that the length of id is not less than minUserChars and not greater than maxUserChars.
func validateUserID(id string) (err error) {
	l := len(id)

	switch {
	case l < minUserChars:
		err = ErrorUserIDShort
	case l > maxUserChars:
		err = ErrorUserIDLong
	}

	return
}

// validateRecipientID ensures that the length of id is not greater than maxUserChars.
func validateRecipientID(id string) (err error) {
	l := len(id)

	switch {
	case l > maxUserChars:
		err = ErrorUserIDLong
	}

	return
}

var ErrorMediaMessageLong = errors.New("m must be >= " + strconv.Itoa(minMessage) + " AND <= " + strconv.Itoa(maxMessage) + ".")

// validateMediaMessage ensures that the length of m is not greater than 256 characters.
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

// validateMediaDir ensures that d only exists if mType is greater than 0.
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

// validateFindDate ensures that fd only exists if f is set to true.
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

// validateShareAll ensures that r only exists if All is set to false.
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

// validateLatitude ensures that the values of l is between minLat and maxLat.
func validateLatitude(l float32) (err error) {
	if l < minLat || l > maxLat {
		err = ErrorLatitude
	}
	return
}

var ErrorLongitude = errors.New("Longitude must be between -90 and 90.")

// validateLongitude ensures that the values of l is between minLong and maxLong.
func validateLongitude(l float32) (err error) {
	if l < minLong || l > maxLong {
		err = ErrorLongitude
	}
	return
}

var ErrorLocationReference = errors.New("Location reference is nil.")

// validateLocation ensures that the value of l is a valid pointer.
func validateLocation(l *Location) (err error) {
	if l == nil {
		err = ErrorLocationReference
	}
	return
}

var ErrorPublicHiddenCombination = errors.New("Public=false AND Hidden=true is an Error input combination.")

// validateMomentPublicHidden ensures that a moment it not both private and hidden.
func validateMomentPublicHidden(p, h bool) (err error) {
	if !p && h {
		err = ErrorPublicHiddenCombination
	}
	return
}

// NewMoment is a constructor for the MomentsRow struct.
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

// NewMedia is a constructor for the MediaRow struct.
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

// NewFind is a constructor for the FindsRow struct
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

// NewShare is a constructor for the SharesRow struct.
func NewShare(mID int, uID string, All bool, r string) (s *SharesRow, err error) {
	check(validateMomentID(mID))
	check(validateUserID(uID))
	check(validateRecipientID(r))
	check(validateShareAll(All, r))

	s = &SharesRow{
		momentID:    mID,
		userID:      uID,
		all:         All,
		recipientID: r,
	}

	return
}

// NewLocation is a constructor for the Location struct.
func NewLocation(lat float32, long float32) (l *Location, err error) {
	check(validateLatitude(lat))
	check(validateLongitude(long))

	l = &Location{
		latitude:  lat,
		longitude: long,
	}

	return
}

// Set is an instance that has insert, delete, values, and args methods.
type Set interface {
	insert()
	delete()
	values()
	args()
}

// MediaRow is a row in the [Moment-Db].[moment].[Media] table.
type MediaRow struct {
	momentID int
	message  string
	mType    uint8
	dir      string
}

// String returns the string representation of a MediaRow instance.
func (m MediaRow) String() string {
	return fmt.Sprintf("momentID: %v\nmType: %v\nmessage: %v\ndir: %v\n", m.momentID, m.mType, m.message, m.dir)
}

// delete is MediaRow method that deletes a row from the [Moment-Db].[moment].[Media] table.
func (m *MediaRow) delete(c *dba.Trans) (rowCnt int, err error) {
	if c == nil {
		c = dba.OpenTx()
		defer func() { c.Close(err) }()
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

// Media is a set of pointers to MediaRow instances.
type Media []*MediaRow

// insert inserts a set of MediaRow instances into the [Moment-Db].[moment].[Media] table.
func (mSet *Media) insert(c *dba.Trans) (rowCnt int, err error) {
	if c == nil {
		c = dba.OpenTx()
		defer func() { c.Close(err) }()
	}

	query := `INSERT INTO [moment].[Media] (MomentID, Message, Type, Dir)
			  VALUES `
	values := mSet.values()
	query = query + values
	args := mSet.args()

	res, err := c.Tx.Exec(query, args...)
	if err != nil {
		return
	}
	cnt, err := res.RowsAffected()
	rowCnt = int(cnt)

	return
}

// values returns a string of parameterized values for an Media Insert query.
func (mSet *Media) values() (values string) {
	vSlice := make([]string, len(*mSet))
	for i := 0; i < len(vSlice); i++ {
		vSlice[i] = "(?, ?, ?, ?)"
	}
	values = strings.Join(vSlice, ", ")

	return
}

// args returns an slice of empty interfaces that hold the arguments for a parameterized query.
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

// delete deletes a set of MediaRows from the [Moment-Db].[moment].[Media] table.
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

// FindsRow is a row in the [Moment-Db].[moment].[Finds] table.
type FindsRow struct {
	momentID int
	userID   string
	found    bool
	findDate *time.Time
}

// String returns the string representation of FindsRow
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

// FindPublic inserts a FindsRow into the [Moment-Db].[moment].[Finds] table with Found=true.
func (f *FindsRow) FindPublic() (id int64, err error) {
	c := dba.OpenConn()
	defer c.Db.Close()

	insert := `INSERT INTO [moment].[Finds]
			   VALUES (?, ?, ?, ?)`
	dt := time.Now().UTC()
	args := []interface{}{f.momentID, f.userID, true, &dt}

	res, err := c.Db.Exec(insert, args...)
	if err != nil {
		return
	}
	id, err = res.LastInsertId()
	return
}

// FindPrivate updates a FindsRow in the [Moment-Db].[moment].[Finds] by setting Found=true.
func (f *FindsRow) FindPrivate() (err error) {
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

// delete deletes a FindsRow from the [Moment-Db].[moment].[Finds] table.
func (f *FindsRow) delete(c *dba.Trans) (rowsAff int, err error) {
	if c == nil {
		c = dba.OpenTx()
		defer func() { c.Close(err) }()
	}

	deleteFrom := `DELETE FROM [moment].[Finds]
				   WHERE MomentID = ?
				   		 AND UserID = ?`
	args := []interface{}{f.momentID, f.userID}

	res, err := c.Tx.Exec(deleteFrom, args...)
	aff, err := res.RowsAffected()
	rowsAff = int(aff)

	return
}

// Finds is a slice of pointers to FindsRow instances.
type Finds []*FindsRow

// insert inserts a Finds instance into the [Moment-Db].[moment].[Finds] table.
func (fSet *Finds) insert(c *dba.Trans) (rowCnt int, err error) {
	if c == nil {
		c = dba.OpenTx()
		defer func() { c.Close(err) }()
	}

	insert := `INSERT [moment].[Finds] (MomentID, UserID, Found, FindDate)
			   VALUES `
	values := fSet.values()
	insert = insert + values
	args := fSet.args()

	res, err := c.Tx.Exec(insert, args...)
	if err != nil {
		return
	}
	cnt, err := res.RowsAffected()
	rowCnt = int(cnt)

	return
}

// values returns a string of parameterized values for an Finds Insert query.
func (fSet *Finds) values() (values string) {
	vSlice := make([]string, len(*fSet))
	for i := 0; i < len(*fSet); i++ {
		vSlice[i] = "(?, ?, ?, ?)"
	}
	values = strings.Join(vSlice, ", ")

	return
}

// args returns an slice of empty interfaces that hold the arguments for a parameterized query.
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

// delete deletes a Finds instance from the [Moment-Db].[moment].[Finds] table.
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

// SharesRow is a row in the [Moment-Db].[moment].[Shares] table.
type SharesRow struct {
	momentID    int
	userID      string
	all         bool
	recipientID string
}

// String returns a string representation of a SharesRow instance.
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

// delete deletes a SharesRow instance from [Moment-Db].[moment].[Shares] table.
func (s *SharesRow) delete(c *dba.Trans) (affCnt int, err error) {
	if c == nil {
		c = dba.OpenTx()
		defer func() { c.Close(err) }()
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

// Shares is a slice of pointers to SharesRow instances.
type Shares []*SharesRow

// Share is an exported package that allows the insertion of a
// Shares instance into the [Moment-Db].[moment].[Shares] table.
func (sSlice *Shares) Share() (rowCnt int, err error) {
	cnt, err := sSlice.insert(nil)
	if err != nil {
		return
	}
}

// insert inserts a Shares instance into [Moment-Db].[moment].[Shares] table.
func (sSlice *Shares) insert(c *dba.Trans) (rowCnt int, err error) {
	if c == nil {
		c = dba.OpenTx()
		defer func() { c.Close(err) }()
	}

	insert := `INSERT INTO [moment].[Shares] (MomentID, UserID, [All], RecipientID)
			   VALUES `
	values := sSlice.values()
	insert = insert + values
	args := sSlice.args()

	res, err := c.Tx.Exec(insert, args...)
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

// values returns a string of parameterized values for a Shares insert query.
func (sSlice *Shares) values() (values string) {
	valuesSlice := make([]string, len(*sSlice))
	for i := 0; i < len(valuesSlice); i++ {
		valuesSlice[i] = "(?, ?, ?, ?)"
	}
	values = strings.Join(valuesSlice, ", ")
	return
}

// args returns an slice of empty interfaces that hold the arguments for a parameterized query.
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

// delete deletes an instance of Shares from the [Moment-Db].[moment].[Shares] table.
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

// String returns the string representation of a Location instance.
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

// MomentsRow is a row in the [Moment-Db].[moment].[Moments] table.
// It is composed of various fields and a Location instance.
type MomentsRow struct {
	Location
	id         int64
	userID     string
	public     bool
	hidden     bool
	createDate time.Time
}

// String returns a string representation of a MomentsRow instance.
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

// CreatePublic creates a row in [Moment-Db].[moment].[Moments] where Public=true.
func (m *MomentsRow) CreatePublic(md *Media) (err error) {

	return
}

// CreatePrivate creates a MomentsRow in [Moment-Db].[moment].[Moments] where Public=true
// and creates Finds in [Moment-Db].[moment].[Finds].
func (m *MomentsRow) CreatePrivate(md *Media, f *Finds) (err error) {

	return
}

// insert inserts a MomentsRow into the [Moment-Db].[moment].[Moments] table.
func (m *MomentsRow) insert(c *dba.Trans) (momentID int64, err error) {
	if c == nil {
		c = dba.OpenTx()
		defer func() { c.Close(err) }()
	}

	insert := `INSERT [moment].[Moments] ([UserID], [Latitude], [Longitude], [Public], [Hidden], [CreateDate])
			   VALUES `
	values := m.values()
	insert = insert + values
	args := m.args()

	res, err := c.Tx.Exec(insert, args...)
	if err != nil {
		return
	}

	if err = dba.ValidateRowsAffected(res, 1); err != nil {
		return
	}

	momentID, err = res.LastInsertId()
	if err != nil {
		return
	}
	m.id = momentID

	return
}

// values returns a string of parameterized values for a MomentsRow Insert query.
func (m *MomentsRow) values() string {
	return "(?, ?, ?, ?, ?, ?)"
}

// args returns an slice of empty interfaces that hold the arguments for a parameterized query.
func (m *MomentsRow) args() []interface{} {
	return []interface{}{m.userID, m.latitude, m.longitude, m.public, m.hidden, m.createDate}
}

// delete deletes a MomentsRow from [Moment-Db].[moment].[Moments].
func (m *MomentsRow) delete(c *dba.Trans) (cnt int64, err error) {
	if c == nil {
		c = dba.OpenTx()
		defer func() { c.Close(err) }()
	}

	deleteFrom := `DELETE FROM [moment].[Moments]
				   WHERE ID = ?`

	res, err := c.Tx.Exec(deleteFrom, m.id)
	if err != nil {
		return
	}
	cnt, err = res.RowsAffected()

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
