package moment

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/penutty/dba"
	"github.com/penutty/util"
	"log"
	"os"
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
)

var (
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
)

func init() {
	Logger := func(logType string) *log.Logger {
		file := "/home/tjp/go/log/" + logType + ".txt"
		f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		return log.New(f, strings.ToUpper(logType)+": ", log.Lshortfile|log.LUTC)
	}

	Info = Logger("info")
	Warn = Logger("warn")
	Error = Logger("error")
}

var (
	ErrorPrivateHiddenMoment = errors.New("m *MomentsRow cannot be both private and hidden")
	ErrorVariableEmpty       = errors.New("Variable is empty.")
)

// NewMoment is a constructor for the MomentsRow struct.
func NewMoment(l *Location, uID string, p bool, h bool, c *time.Time) (m *MomentsRow, err error) {
	m = new(MomentsRow)

	m.setLocation(l)
	m.setCreateDate(c)
	m.setUserID(uID)
	if m.err != nil {
		Error.Println(m.err)
		return m, m.err
	}

	m.hidden = h
	m.public = p
	if m.hidden && !m.public {
		Error.Println(ErrorPrivateHiddenMoment)
		return m, ErrorPrivateHiddenMoment
	}

	return m, m.err
}

var ErrorLocationIsNil = errors.New("l *Location is nil")

func (m *MomentsRow) setLocation(l *Location) {
	if m.err != nil {
		return
	}

	if l == nil {
		m.err = ErrorLocationIsNil
		return
	}
	m.Location = *l
	return
}

func (m *MomentsRow) setCreateDate(t *time.Time) {
	if m.err != nil {
		return
	}

	if err := checkTime(t); err != nil {
		m.err = err
		return
	}
	m.createDate = t
	return
}

func (m *MomentsRow) setUserID(u string) {
	if m.err != nil {
		return
	}
	m.err = m.uID.setUserID(u)
	return
}

var ErrorMediaDNE = errors.New("m.mType is set to DNE, therefore m.dir must remain empty.")
var ErrorMediaExistsDirDNE = errors.New("m.mType is not DNE, therefore m.dir must be set.")
var ErrorMessageLong = errors.New("m must be >= " + strconv.Itoa(minMessage) + " AND <= " + strconv.Itoa(maxMessage) + ".")

// NewMedia is a constructor for the MediaRow struct.
func NewMedia(mID int64, m string, mType uint8, d string) (mr *MediaRow, err error) {
	mr = new(MediaRow)

	mr.setMomentID(mID)
	mr.setMessage(m)
	mr.setmType(mType)
	if mr.err != nil {
		Error.Println(mr.err)
		return mr, mr.err
	}

	mr.dir = d

	if mr.mType == DNE && mr.dir != "" {
		Error.Println(ErrorMediaDNE)
		return mr, ErrorMediaDNE
	}
	if mr.mType != DNE && mr.dir == "" {
		Error.Println(ErrorMediaExistsDirDNE)
		return mr, ErrorMediaExistsDirDNE
	}

	return
}

// setMediaType ensures that t is a value between minMediaType and maxMediaType.
func (mr *MediaRow) setmType(t uint8) {
	if mr.err != nil {
		return
	}
	if err := checkMediaType(t); err != nil {
		mr.err = err
		return
	}

	mr.mType = t
	return
}

func (mr *MediaRow) setMessage(m string) {
	if mr.err != nil {
		return
	}
	if l := len(m); l > maxMessage {
		mr.err = ErrorMessageLong
		return
	}

	mr.message = m
	return
}

func (mr *MediaRow) setMomentID(mID int64) {
	if mr.err != nil {
		return
	}
	mr.err = mr.mID.setMomentID(mID)
	return
}

var ErrorFoundEmptyFindDate = errors.New("fr.found=true, therefore fr.findDate must not be empty")
var ErrorNotFoundFindDateExists = errors.New("fr.found=false, therefore fr.findDate must be empty.")

// NewFind is a constructor for the FindsRow struct
func NewFind(mID int64, uID string, f bool, fd *time.Time) (fr *FindsRow, err error) {
	fr = new(FindsRow)

	fr.setMomentID(mID)
	fr.setUserID(uID)
	fr.setFindDate(fd)
	if fr.err != nil {
		Error.Println(fr.err)
		return fr, fr.err
	}

	fr.found = f

	emptyTime := time.Time{}
	if fr.found && *fr.findDate == emptyTime {
		Error.Println(ErrorFoundEmptyFindDate)
		return fr, ErrorFoundEmptyFindDate
	}
	if !fr.found && *fr.findDate != emptyTime {
		Error.Println(ErrorNotFoundFindDateExists)
		return fr, ErrorNotFoundFindDateExists
	}

	return
}

func (f *FindsRow) setFindDate(fd *time.Time) {
	if err := checkTime(fd); err != nil {
		f.err = err
		return
	}
	f.findDate = fd
	return
}

func (f *FindsRow) setMomentID(mID int64) {
	if f.err != nil {
		return
	}
	f.err = f.mID.setMomentID(mID)
	return
}

func (f *FindsRow) setUserID(uID string) {
	if f.err != nil {
		return
	}
	f.err = f.uID.setUserID(uID)
	return
}

var ErrorAllRecipientExists = errors.New("s.all=true, therefore s.recipientID must be \"\"")
var ErrorNotAllRecipientDNE = errors.New("s.all=false, therefore s.recipientID must be set")

// NewShare is a constructor for the SharesRow struct.
func NewShare(mID int64, uID string, all bool, r string) (s *SharesRow, err error) {
	s = new(SharesRow)

	s.setMomentID(mID)
	s.setUserID(uID)
	s.setRecipientID(r)
	if s.err != nil {
		Error.Println(s.err)
		return s, s.err
	}

	s.all = all

	if s.all && s.recipientID != "" {
		Error.Println(ErrorAllRecipientExists)
		return s, ErrorAllRecipientExists
	}
	if !s.all && s.recipientID == "" {
		Error.Println(ErrorNotAllRecipientDNE)
		return s, ErrorNotAllRecipientDNE
	}

	return
}

func (s *SharesRow) setRecipientID(id string) {
	if s.err != nil {
		return
	}

	if err := checkUserIDLong(id); err != nil {
		s.err = err
		return
	}
	s.recipientID = id
	return
}

func (s *SharesRow) setMomentID(mID int64) {
	if s.err != nil {
		return
	}
	s.err = s.mID.setMomentID(mID)
}

func (s *SharesRow) setUserID(uID string) {
	if s.err != nil {
		return
	}
	s.err = s.uID.setUserID(uID)
}

var ErrorLatitude = errors.New("Latitude must be between -180 and 180.")
var ErrorLongitude = errors.New("Longitude must be between -90 and 90.")

// NewLocation is a constructor for the Location struct.
func NewLocation(lat float32, long float32) (*Location, error) {
	lo := new(Location)

	lo.setLatitude(lat)
	lo.setLongitude(long)
	if lo.err != nil {
		Error.Println(lo.err)
		return lo, lo.err
	}

	return lo, lo.err
}

// setLatitude ensures that the values of l is between minLat and maxLat.
func (lo *Location) setLatitude(l float32) {
	if lo.err != nil {
		return
	}
	if l < minLat || l > maxLat {
		lo.err = ErrorLatitude
		return
	}
	lo.latitude = l
	return
}

// setLongitude ensures that the values of l is between minLong and maxLong.
func (lo *Location) setLongitude(l float32) {
	if lo.err != nil {
		return
	}
	if l < minLong || l > maxLong {
		lo.err = ErrorLongitude
		return
	}
	lo.longitude = l
	return
}

type mID struct {
	momentID int64
}

func (m *mID) setMomentID(id int64) (err error) {
	if err = checkMomentID(id); err != nil {
		return err
	}
	m.momentID = id
	return
}

type uID struct {
	userID string
}

func (u *uID) setUserID(id string) (err error) {
	if err = checkUserID(id); err != nil {
		return err
	}
	u.userID = id
	return
}

// check is a helper function that is used to ensure another function did not return an error.
func check(err error) {
	if err != nil {
		panic(err)
	}
}

var ErrorTimePtrNil = errors.New("t *time.Time is set to nil")

// checkTime ensures that the value of t is a valid address.
func checkTime(t *time.Time) (err error) {
	if t == nil {
		return ErrorTimePtrNil
	}
	return
}

var ErrorMomentID = errors.New("*id must be >= 1.")

// checkMomentID ensures that id is greater 0.
func checkMomentID(id int64) (err error) {
	if id < 1 {
		return ErrorMomentID
	}
	return
}

var ErrorMediaTypeDNE = errors.New("*t must be >= " + strconv.Itoa(minMediaType) + " AND <= " + strconv.Itoa(maxMediaType))

// checkMediaType ensures that t is less than maxMediaType.
func checkMediaType(t uint8) (err error) {
	if t > maxMediaType {
		return ErrorMediaTypeDNE
	}
	return
}

var (
	ErrorUserIDShort = errors.New("len(*id) (userID) must be >= " + strconv.Itoa(minUserChars) + ".")
	ErrorUserIDLong  = errors.New("len(*id) (userID) must be <= " + strconv.Itoa(maxUserChars) + ".")
)

// checkUserID ensures that the length of id is between minUserChars and maxUserChars.
func checkUserID(id string) (err error) {
	if err = checkUserIDLong(id); err != nil {
		return
	}
	if err = checkUserIDShort(id); err != nil {
		return
	}
	return
}

// checkUserIDLong ensures that the length of id is less than maxUserChars.
func checkUserIDLong(id string) (err error) {
	if l := len(id); l > maxUserChars {
		return ErrorUserIDLong
	}
	return
}

// checkUserIDShort ensures that the length of id is greater than minUserChars.
func checkUserIDShort(id string) (err error) {
	if l := len(id); l < minUserChars {
		return ErrorUserIDShort
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
	mID
	message string
	mType   uint8
	dir     string
	err     error
}

// String returns the string representation of a MediaRow instance.
func (m MediaRow) String() string {
	return fmt.Sprintf("momentID: %v\nmType: %v\nmessage: %v\ndir: %v\n", m.momentID, m.mType, m.message, m.dir)
}

// delete is MediaRow method that deletes a row from the [Moment-Db].[moment].[Media] table.
func (m *MediaRow) delete(c *dba.Trans) (rowCnt int64, err error) {
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
		Error.Println(err)
		return
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		Error.Println(err)
		return
	}
	rowCnt = int64(cnt)

	return
}

// Media is a set of pointers to MediaRow instances.
type Media []*MediaRow

// insert inserts a set of MediaRow instances into the [Moment-Db].[moment].[Media] table.
func (mSet Media) insert(c *dba.Trans) (rowCnt int64, err error) {
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
		Error.Println(err)
		return
	}
	cnt, err := res.RowsAffected()
	if err != nil {
		Error.Println(err)
		return
	}
	rowCnt = int64(cnt)

	return
}

// values returns a string of parameterized values for an Media Insert query.
func (mSet Media) values() (values string) {
	vSlice := make([]string, len(mSet))
	for i := 0; i < len(vSlice); i++ {
		vSlice[i] = "(?, ?, ?, ?)"
	}
	values = strings.Join(vSlice, ", ")

	return
}

// args returns an slice of empty interfaces that hold the arguments for a parameterized query.
func (mSet Media) args() (args []interface{}) {
	fCnt := 4
	argsCnt := len(mSet) * fCnt
	args = make([]interface{}, argsCnt)
	for i, m := range mSet {
		j := i * 4
		args[j] = m.momentID
		args[j+1] = m.message
		args[j+2] = m.mType
		args[j+3] = m.dir
	}
	return
}

// delete deletes a set of MediaRows from the [Moment-Db].[moment].[Media] table.
func (mSet Media) delete() (rowCnt int64, err error) {
	c := dba.OpenTx()
	defer func() { c.Close(err) }()

	var affCnt int64
	for _, m := range mSet {
		affCnt, err = m.delete(c)
		if err != nil {
			Error.Println(err)
			rowCnt = 0
			return
		}
		rowCnt += affCnt
	}

	return
}

// FindsRow is a row in the [Moment-Db].[moment].[Finds] table.
type FindsRow struct {
	mID
	uID
	found    bool
	findDate *time.Time
	err      error
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

var ErrorFoundFalseFindDateNil = errors.New("A found row must have f.found=true and f.findDate=*time.Time{}")

// FindPublic inserts a FindsRow into the [Moment-Db].[moment].[Finds] table with Found=true.
func (f *FindsRow) FindPublic() (cnt int64, err error) {
	if f.userID == "" || f.momentID == 0 {
		Error.Println(ErrorVariableEmpty)
		return 0, ErrorVariableEmpty
	}

	c := dba.OpenConn()
	defer c.Db.Close()

	insert := `INSERT INTO [moment].[Finds]
			   VALUES (?, ?, ?, ?)`
	dt := time.Now().UTC()
	args := []interface{}{f.momentID, f.userID, true, &dt}

	res, err := c.Db.Exec(insert, args...)
	if err != nil {
		Error.Println(err)
		return
	}
	cnt, err = res.RowsAffected()
	if err != nil {
		Error.Println(err)
	}
	return
}

// FindPrivate updates a FindsRow in the [Moment-Db].[moment].[Finds] by setting Found=true.
func (f *FindsRow) FindPrivate() (err error) {
	if f.userID == "" || f.momentID == 0 {
		Error.Println(ErrorVariableEmpty)
		return ErrorVariableEmpty
	}

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
		Error.Println(err)
		return
	}
	err = dba.ValidateRowsAffected(res, 1)
	if err != nil {
		Error.Println(err)
	}
	return
}

// delete deletes a FindsRow from the [Moment-Db].[moment].[Finds] table.
func (f *FindsRow) delete(c *dba.Trans) (rowsAff int64, err error) {
	if c == nil {
		c = dba.OpenTx()
		defer func() { c.Close(err) }()
	}

	deleteFrom := `DELETE FROM [moment].[Finds]
				   WHERE MomentID = ?
				   		 AND UserID = ?`
	args := []interface{}{f.momentID, f.userID}

	res, err := c.Tx.Exec(deleteFrom, args...)
	if err != nil {
		Error.Println(err)
		return
	}

	rowsAff, err = res.RowsAffected()
	if err != nil {
		Error.Println(err)
		return
	}

	return
}

// Finds is a slice of pointers to FindsRow instances.
type Finds []*FindsRow

// insert inserts a Finds instance into the [Moment-Db].[moment].[Finds] table.
func (fSet Finds) insert(c *dba.Trans) (rowCnt int64, err error) {
	if c == nil {
		c = dba.OpenTx()
		defer func() { c.Close(err) }()
	}

	insert := `INSERT [moment].[Finds] (MomentID, UserID, Found, FindDate)
			   VALUES `
	values := fSet.values()
	insert = insert + values
	args, err := fSet.args()
	if err != nil {
		Error.Println(err)
		return
	}
	res, err := c.Tx.Exec(insert, args...)
	if err != nil {
		Error.Println(err)
		return
	}
	rowCnt, err = res.RowsAffected()
	if err != nil {
		Error.Println(err)
		return
	}

	return
}

// values returns a string of parameterized values for an Finds Insert query.
func (fSet Finds) values() (values string) {
	vSlice := make([]string, len(fSet))
	for i := 0; i < len(fSet); i++ {
		vSlice[i] = "(?, ?, ?, ?)"
	}
	values = strings.Join(vSlice, ", ")

	return
}

// args returns an slice of empty interfaces that hold the arguments for a parameterized query.
func (fSet Finds) args() (args []interface{}, err error) {
	findFieldCnt := 4
	argsCnt := len(fSet) * findFieldCnt
	args = make([]interface{}, argsCnt)

	for i, f := range fSet {
		j := 4 * i
		args[j] = f.momentID
		args[j+1] = f.userID
		args[j+2] = f.found
		args[j+3] = f.findDate
	}

	return
}

// delete deletes a Finds instance from the [Moment-Db].[moment].[Finds] table.
func (fSet Finds) delete() (rowCnt int64, err error) {
	c := dba.OpenTx()
	defer func() { c.Close(err) }()

	var rAff int64
	for _, f := range fSet {
		rAff, err = f.delete(c)
		if err != nil {
			Error.Println(err)
			rowCnt = 0
			return
		}
		rowCnt += rAff
	}

	return
}

// SharesRow is a row in the [Moment-Db].[moment].[Shares] table.
type SharesRow struct {
	mID
	uID
	all         bool
	recipientID string
	err         error
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
func (s *SharesRow) delete(c *dba.Trans) (affCnt int64, err error) {
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
		Error.Println(err)
		return
	}

	affCnt, err = res.RowsAffected()
	if err != nil {
		Error.Println(err)
		return
	}

	return
}

// Shares is a slice of pointers to SharesRow instances.
type Shares []*SharesRow

// Share is an exported package that allows the insertion of a
// Shares instance into the [Moment-Db].[moment].[Shares] table.
func (sSlice Shares) Share() (rowCnt int64, err error) {
	rowCnt, err = sSlice.insert()
	if err != nil {
		Error.Println(err)
		return
	}
	return
}

// insert inserts a Shares instance into [Moment-Db].[moment].[Shares] table.
func (sSlice Shares) insert() (rowCnt int64, err error) {
	c := dba.OpenConn()
	defer c.Db.Close()

	insert := `INSERT INTO [moment].[Shares] (MomentID, UserID, [All], RecipientID)
			   VALUES `
	values := sSlice.values()
	insert = insert + values
	args := sSlice.args()

	res, err := c.Db.Exec(insert, args...)
	if err != nil {
		Error.Println(err)
		return
	}
	rowCnt, err = res.RowsAffected()
	if err != nil {
		Error.Println(err)
		return
	}

	return
}

// values returns a string of parameterized values for a Shares insert query.
func (sSlice Shares) values() (values string) {
	valuesSlice := make([]string, len(sSlice))
	for i := 0; i < len(valuesSlice); i++ {
		valuesSlice[i] = "(?, ?, ?, ?)"
	}
	values = strings.Join(valuesSlice, ", ")
	return
}

// args returns an slice of empty interfaces that hold the arguments for a parameterized query.
func (sSlice Shares) args() (args []interface{}) {
	SharesFieldCnt := 4
	argsCnt := len(sSlice) * SharesFieldCnt
	args = make([]interface{}, argsCnt)

	for i, s := range sSlice {
		j := i * 4
		args[j] = s.momentID
		args[j+1] = s.userID
		args[j+2] = s.all
		args[j+3] = s.recipientID
	}

	return
}

// delete deletes an instance of Shares from the [Moment-Db].[moment].[Shares] table.
func (sSlice Shares) delete() (rowCnt int64, err error) {
	c := dba.OpenTx()
	defer func() { c.Close(err) }()

	var affCnt int64
	for _, s := range sSlice {
		affCnt, err = s.delete(c)
		if err != nil {
			Error.Println(err)
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
	err       error
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
	mID
	uID
	public     bool
	hidden     bool
	createDate *time.Time
	err        error
}

// String returns a string representation of a MomentsRow instance.
func (m MomentsRow) String() string {
	return fmt.Sprintf("id: %v\n"+
		"userID: %v\n"+
		"Location: %v\n"+
		"public: %v\n"+
		"hidden: %v\n"+
		"createDate: %v\n",
		m.momentID,
		m.userID,
		m.Location,
		m.public,
		m.hidden,
		m.createDate)
}

var ErrorMediaPointerNil = errors.New("md *Media is nil.")

// CreatePublic creates a row in [Moment-Db].[moment].[Moments] where Public=true.
func (m *MomentsRow) CreatePublic(media *Media) (err error) {
	if media == nil {
		return ErrorMediaPointerNil
	}

	c := dba.OpenTx()
	defer func() { c.Close(err) }()

	mID, err := m.insert(c)
	if err != nil {
		Error.Println(err)
		return
	}
	m.momentID = mID

	for _, p := range *media {
		p.setMomentID(m.momentID)
		if p.err != nil {
			Error.Println(err)
			return
		}
	}
	if _, err = media.insert(c); err != nil {
		Error.Println(err)
		return
	}

	return
}

var ErrorFindsPointerNil = errors.New("finds *Finds pointer is empty.")

// CreatePrivate creates a MomentsRow in [Moment-Db].[moment].[Moments] where Public=true
// and creates Finds in [Moment-Db].[moment].[Finds].
func (m *MomentsRow) CreatePrivate(media *Media, finds *Finds) (err error) {
	if media == nil {
		Error.Println(ErrorMediaPointerNil)
		return ErrorMediaPointerNil
	}
	if finds == nil {
		Error.Println(ErrorFindsPointerNil)
		return ErrorFindsPointerNil
	}

	c := dba.OpenTx()
	defer func() { c.Close(err) }()

	mID, err := m.insert(c)
	if err != nil {
		Error.Println(err)
		return
	}
	m.momentID = mID

	for _, p := range *media {
		p.setMomentID(m.momentID)
		if p.err != nil {
			Error.Println(p.err)
			return
		}
	}

	for _, p := range *finds {
		p.setMomentID(m.momentID)
		if p.err != nil {
			Error.Println(p.err)
			return
		}
	}

	if _, err = media.insert(c); err != nil {
		Error.Println(err)
		return
	}
	if _, err = finds.insert(c); err != nil {
		Error.Println(err)
		return
	}

	return
}

// insert inserts a MomentsRow into the [Moment-Db].[moment].[Moments] table.
func (m *MomentsRow) insert(c *dba.Trans) (mID int64, err error) {
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
		Error.Println(err)
		return
	}

	if err = dba.ValidateRowsAffected(res, 1); err != nil {
		Error.Println(err)
		return
	}

	mID, err = res.LastInsertId()
	if err != nil {
		Error.Println(err)
		return
	}
	m.momentID = mID

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

	res, err := c.Tx.Exec(deleteFrom, m.momentID)
	if err != nil {
		Error.Println(err)
		return
	}
	cnt, err = res.RowsAffected()
	if err != nil {
		Error.Println(err)
	}
	return
}

// Result is grouping of a *MomentsRow with its related *FindsRow(s), *MediaRow(s), and *SharesRow(s).
type Result struct {
	moment *MomentsRow
	finds  Finds
	media  Media
	shares Shares
}

// Results is a slice of *Result.
type Results []*Result

// ResultMap is a grouping of a *MomentsRow and mappingts to its related *FindsRow(s), *MediaRow(s), and *SharesRow(s).
type ResultMap struct {
	moment *MomentsRow
	finds  FindsMap
	media  MediaMap
	shares SharesMap
}

// ResultsMap is a mapping of *MomentsRow.momentId to a *Result
type ResultsMap map[int64]*ResultMap

var ErrorTypeNotImplemented = errors.New("Type switch does not handle specified type.")

func (rm ResultsMap) add(mRow *MomentsRow, rs ...interface{}) (err error) {
	r := new(ResultMap)
	r.moment = mRow
	for _, i := range rs {
		switch v := i.(type) {
		case *MediaRow:
			r.media = make(MediaMap)
			if err = r.media.add(v); err != nil {
				Error.Println(err)
				return
			}
		case *FindsRow:
			r.finds = make(FindsMap)
			if err = r.finds.add(v); err != nil {
				Error.Println(err)
				return
			}
		case *SharesRow:
			r.shares = make(SharesMap)
			if err = r.shares.add(v); err != nil {
				Error.Println(err)
				return
			}
		default:
			Error.Println(err)
			return ErrorTypeNotImplemented
		}
	}
	rm[mRow.momentID] = r
	return
}
func (rs ResultsMap) exists(mRow *MomentsRow) bool {
	_, ok := rs[mRow.momentID]
	return ok
}
func (rs ResultsMap) append(mRow *MomentsRow, is ...interface{}) (err error) {
	mID := mRow.momentID
	for _, i := range is {
		switch v := i.(type) {
		case *MediaRow:
			if !rs[mID].media.exists(v) {
				if err = rs[mID].media.add(v); err != nil {
					Error.Println(err)
					return
				}
			}
		case *FindsRow:
			if !rs[mID].finds.exists(v) {
				if err = rs[mID].finds.add(v); err != nil {
					Error.Println(err)
					return
				}
			}
		case *SharesRow:
			if !rs[mID].shares.exists(v) {
				if err = rs[mID].shares.add(v); err != nil {
					Error.Println(err)
					return
				}
			}
		default:
			Error.Println(ErrorTypeNotImplemented)
			return ErrorTypeNotImplemented
		}
	}
	return
}

func (rs ResultsMap) mapToSlice() (r Results) {
	r = make(Results, len(rs))
	i := 0
	for _, v := range rs {
		newR := &Result{
			moment: v.moment,
			media:  v.media.mapToSlice(),
			finds:  v.finds.mapToSlice(),
			shares: v.shares.mapToSlice(),
		}
		r[i] = newR
		i++
	}
	return
}

// SharesMap is a map of pointers to SharesRow instances.
type SharesMap map[[2]string]*SharesRow

// add inserts a pointer to a SharesRow instance into the SharesMap receiver.
func (sm SharesMap) add(s *SharesRow) (err error) {
	if util.IsEmpty(s) {
		return ErrorVariableEmpty
	}
	index := [2]string{s.userID, s.recipientID}
	sm[index] = s
	return
}

// exists checks if the specified SharesRow instance is already in the SharesMap receiver.
func (sm SharesMap) exists(s *SharesRow) bool {
	index := [2]string{s.userID, s.recipientID}
	_, ok := sm[index]
	return ok
}

// mapToSlice converts a SharesMap instance into a Shares instance.
func (sm SharesMap) mapToSlice() (s Shares) {
	s = make(Shares, len(sm))
	var i int
	for _, v := range sm {
		s[i] = v
		i++
	}
	return
}

// sliceToMap converts a Shares instance to a SharesMap instance
func (s Shares) sliceToMap() (sm SharesMap) {
	sm = make(SharesMap)
	for _, v := range s {
		index := [2]string{v.userID, v.recipientID}
		sm[index] = v
	}
	return
}

// FindsMap is a map of pointers to FindsRow instances.
type FindsMap map[string]*FindsRow

// add inserts a pointer to a FindsRow instance into the Findsmap receiver.
func (fm FindsMap) add(f *FindsRow) (err error) {
	if util.IsEmpty(f) {
		Error.Println(ErrorVariableEmpty)
		return ErrorVariableEmpty
	}
	fm[f.userID] = f
	return
}

// exists checks if the f *FindsRow exists in the FindsMap receiver.
func (fm FindsMap) exists(f *FindsRow) bool {
	_, ok := fm[f.userID]
	return ok
}

// mapToSlice converts a FindsMap instance into a Finds instance.
func (fm FindsMap) mapToSlice() (f Finds) {
	f = make(Finds, len(fm))
	var i int
	for _, v := range fm {
		f[i] = v
		i++
	}
	return
}

// sliceToMap converts a Shares instance into a SharesMap instance.
func (f Finds) sliceToMap() (fm FindsMap) {
	fm = make(FindsMap)
	for _, v := range f {
		fm[v.userID] = v
	}
	return
}

// MediaMap is a map of pointers to MediaRow instances.
type MediaMap map[uint8]*MediaRow

// add inserts a pointer to a MediaRow instance into the MediaMap receiver.
func (mp MediaMap) add(md *MediaRow) (err error) {
	if util.IsEmpty(md) {
		Error.Println(ErrorVariableEmpty)
		return ErrorVariableEmpty
	}
	mp[md.mType] = md
	return
}

// exists checks if the specified MediaRow instance is already in the MediaMap receiver.
func (mp MediaMap) exists(md *MediaRow) bool {
	_, ok := mp[md.mType]
	return ok
}

// mapToSlice converts a MediaMap instance into a Media instance.
func (mm MediaMap) mapToSlice() (m Media) {
	m = make(Media, len(mm))
	var i int
	for _, v := range mm {
		m[i] = v
		i++
	}
	return
}

// sliceToMap converts a Media instance into MediaMap instance.
func (m Media) sliceToMap() (mm MediaMap) {
	mm = make(MediaMap)
	for _, v := range m {
		mm[v.mType] = v
	}
	return
}

const (
	momentSchema = "moment"

	momentAlias = "m"
	mediaAlias  = "med"
	sharesAlias = "s"
	findsAlias  = "f"

	noJoin  = ""
	noAlias = ""
)

func momentsLostColumns() (c []*dba.Column, err error) {
	c = make([]*dba.Column, 3)
	c[0], err = dba.NewColumn(momentAlias, "ID", noAlias)
	c[1], err = dba.NewColumn(momentAlias, "Latitude", noAlias)
	c[2], err = dba.NewColumn(momentAlias, "Longitude", noAlias)
	return
}

func momentsColumns() (c []*dba.Column, err error) {
	c = make([]*dba.Column, 4)
	c[0], err = dba.NewColumn(momentAlias, "UserID", noAlias)
	c[1], err = dba.NewColumn(momentAlias, "[Public]", noAlias)
	c[2], err = dba.NewColumn(momentAlias, "[Hidden]", noAlias)
	c[3], err = dba.NewColumn(momentAlias, "CreateDate", noAlias)
	lc, err := momentsLostColumns()
	if err != nil {
		Error.Println(err)
		return
	}
	c = append(c, lc...)
	return
}

func momentsFrom(join string) (t *dba.Table, err error) {
	t, err = dba.NewTable(momentSchema, "Moments", momentAlias, join)
	return
}

func momentsJoin(alias string) string {
	return momentAlias + ".ID = " + alias + ".MomentID"
}

func momentsWhereLocation(l *Location) (w []*dba.Where, err error) {
	w = make([]*dba.Where, 2)
	w[0], err = dba.NewWhere("", momentAlias+".Latitude BETWEEN ? AND ?", []interface{}{l.latitude - 1, l.latitude + 1})
	w[1], err = dba.NewWhere("AND", momentAlias+".Longitude BETWEEN ? AND ?", []interface{}{l.longitude - 1, l.longitude + 1})
	return
}

func momentsWherePublic() (w []*dba.Where, err error) {
	w = make([]*dba.Where, 2)
	w[0], err = dba.NewWhere("AND", momentAlias+".[Public] = 1", nil)
	w[1], err = dba.NewWhere("AND", momentAlias+".[Hidden] = 0", nil)
	return
}

func momentsWhereHidden() (w []*dba.Where, err error) {
	w = make([]*dba.Where, 2)
	w[0], err = dba.NewWhere("AND", momentAlias+".[Public] = 1", nil)
	w[1], err = dba.NewWhere("AND", momentAlias+".[Hidden] = 1", nil)
	return
}

func momentsWherePrivate() (w *dba.Where, err error) {
	w, err = dba.NewWhere("AND", momentAlias+".[Public] = 0", nil)
	return
}

func momentsWhereUserID(u string) (w *dba.Where, err error) {
	w, err = dba.NewWhere("", momentAlias+".UserID = ?", []interface{}{u})
	return
}

func mediaColumns() (c []*dba.Column, err error) {
	c = make([]*dba.Column, 3)
	c[0], err = dba.NewColumn(mediaAlias, "Message", noAlias)
	c[1], err = dba.NewColumn(mediaAlias, "Type", noAlias)
	c[2], err = dba.NewColumn(mediaAlias, "Dir", noAlias)
	return
}

func mediaFrom(join string) (t *dba.Table, err error) {
	t, err = dba.NewTable(momentSchema, "Media", mediaAlias, join)
	return
}

func findsColumns() (c []*dba.Column, err error) {
	c = make([]*dba.Column, 2)
	c[0], err = dba.NewColumn(findsAlias, "UserID", noAlias)
	c[1], err = findsFindDateColumn()
	return
}

func findsFindDateColumn() (c *dba.Column, err error) {
	c, err = dba.NewColumn(findsAlias, "FindDate", noAlias)
	return
}

func findsFrom(join string) (t *dba.Table, err error) {
	t, err = dba.NewTable(momentSchema, "Finds", findsAlias, join)
	return
}

func findsWhereLostUserId(u string) (w []*dba.Where, err error) {
	w = make([]*dba.Where, 2)
	w[0], err = dba.NewWhere("AND", findsAlias+".Found = 0", nil)
	w[1], err = dba.NewWhere("AND", findsAlias+".UserID = ?", []interface{}{u})
	return
}

func findsWhereUserIDFound(u string) (w []*dba.Where, err error) {
	w = make([]*dba.Where, 2)
	w[0], err = dba.NewWhere("", findsAlias+".Found = 1", nil)
	w[1], err = dba.NewWhere("AND", findsAlias+".UserID = ?", []interface{}{u})
	return
}

func sharesFrom(join string) (t *dba.Table, err error) {
	t, err = dba.NewTable(momentSchema, "Shares", sharesAlias, join)
	return
}

func sharesWhereRecipient(me string) (w *dba.Where, err error) {
	w, err = dba.NewWhere("AND", "("+sharesAlias+".RecipientID = ? OR s.[All] = 1)", []interface{}{me})
	return
}

func sharesWhereUserIDRecipient(you, me string) (w []*dba.Where, err error) {
	w = make([]*dba.Where, 2)
	w[0], err = dba.NewWhere("", "s.UserID = ?", []interface{}{you})
	w[1], err = dba.NewWhere("AND", "(s.RecipientID = ? OR s.[All] = 1)", []interface{}{me})
	return
}

func momentsQuery() (q *dba.Query, err error) {
	q = dba.NewQuery("Standard Moments Select")

	mc, err := momentsColumns()
	medc, err := mediaColumns()
	if err != nil {
		Error.Println(err)
		return
	}

	columns := append(mc, medc...)
	if err = q.SetColumns(columns...); err != nil {
		Error.Println(err)
		return
	}

	mf, err := momentsFrom(noJoin)
	medf, err := mediaFrom(momentsJoin(mediaAlias))
	if err != nil {
		Error.Println(err)
		return
	}

	froms := []*dba.Table{
		mf,
		medf,
	}
	if err = q.SetFroms(froms...); err != nil {
		Error.Println(err)
		return
	}

	return
}

func lostQuery() (q *dba.Query, err error) {
	q = dba.NewQuery("Lost Moments Query")

	mlc, err := momentsLostColumns()
	if err != nil {
		Error.Println(err)
		return
	}
	if err = q.SetColumns(mlc...); err != nil {
		Error.Println(err)
		return
	}

	mf, err := momentsFrom(noJoin)
	if err != nil {
		Error.Println(err)
		return
	}
	if err = q.SetFroms(mf); err != nil {
		Error.Println(err)
		return
	}
	return
}

func LocationShared(l *Location, me string) (r ResultsMap, err error) {
	if util.IsEmpty(l) || util.IsEmpty(me) {
		Error.Println(ErrorVariableEmpty)
		return r, ErrorVariableEmpty
	}

	query, err := momentsQuery()
	if err != nil {
		Error.Println(err)
		return
	}

	sf, err := sharesFrom(momentsJoin(sharesAlias))
	if err != nil {
		Error.Println(err)
		return
	}
	if err = query.SetFroms(sf); err != nil {
		Error.Println(err)
		return
	}

	mwl, err := momentsWhereLocation(l)
	if err != nil {
		Error.Println(err)
		return
	}
	swr, err := sharesWhereRecipient(me)
	if err != nil {
		Error.Println(err)
		return
	}
	wheres := append(mwl, swr)
	if err = query.SetWheres(wheres...); err != nil {
		Error.Println(err)
		return
	}

	r, err = process(query, momentsResults)

	return
}

func LocationPublic(l *Location) (r ResultsMap, err error) {
	if util.IsEmpty(l) {
		Error.Println(err)
		return r, ErrorVariableEmpty
	}

	query, err := momentsQuery()
	if err != nil {
		Error.Println(err)
		return
	}

	mwl, err := momentsWhereLocation(l)
	if err != nil {
		Error.Println(err)
		return
	}
	mwp, err := momentsWherePublic()
	if err != nil {
		Error.Println(err)
		return
	}

	wheres := append(mwl, mwp...)
	if err = query.SetWheres(wheres...); err != nil {
		Error.Println(err)
		return
	}

	r, err = process(query, momentsResults)
	if err != nil {
		Error.Println(err)
	}
	return
}

func LocationHidden(l *Location) (r ResultsMap, err error) {
	if util.IsEmpty(l) {
		return r, ErrorLocationIsNil
	}

	query, err := lostQuery()
	if err != nil {
		Error.Println(err)
		return
	}

	mwl, err := momentsWhereLocation(l)
	if err != nil {
		Error.Println(err)
		return
	}

	mwh, err := momentsWhereHidden()
	if err != nil {
		Error.Println(err)
		return
	}

	wheres := append(mwl, mwh...)
	if err = query.SetWheres(wheres...); err != nil {
		Error.Println(err)
		return
	}

	r, err = process(query, lostResults)
	if err != nil {
		Error.Println(err)
	}
	return
}

func LocationLost(l *Location, me string) (r ResultsMap, err error) {
	if util.IsEmpty(l) {
		return r, ErrorLocationIsNil
	}
	if util.IsEmpty(me) {
		return r, ErrorVariableEmpty
	}

	query, err := lostQuery()
	if err != nil {
		return
	}

	ff, err := findsFrom(momentsJoin(findsAlias))
	if err != nil {
		return
	}
	if err = query.SetFroms(ff); err != nil {
		return
	}

	mwl, err := momentsWhereLocation(l)
	if err != nil {
		return
	}
	mwp, err := momentsWherePrivate()
	if err != nil {
		return
	}
	fwlu, err := findsWhereLostUserId(me)
	if err != nil {
		return
	}

	wheres := append(mwl, mwp)
	wheres = append(wheres, fwlu...)
	if err = query.SetWheres(wheres...); err != nil {
		return
	}

	r, err = process(query, lostResults)
	return
}

func UserShared(me string, u string) (r ResultsMap, err error) {
	if util.IsEmpty(me) || util.IsEmpty(u) {
		return r, ErrorVariableEmpty
	}

	query, err := momentsQuery()
	if err != nil {
		return
	}

	sf, err := sharesFrom(momentsJoin(sharesAlias))
	if err != nil {
		return
	}

	if err = query.SetFroms(sf); err != nil {
		return
	}

	swur, err := sharesWhereUserIDRecipient(u, me)
	if err != nil {
		return
	}
	if err = query.SetWheres(swur...); err != nil {
		return
	}

	r, err = process(query, momentsResults)
	return
}

func UserLeft(me string) (r ResultsMap, err error) {
	if util.IsEmpty(me) {
		return r, ErrorVariableEmpty
	}

	query, err := momentsQuery()
	if err != nil {
		return
	}

	fc, err := findsColumns()
	if err != nil {
		return
	}
	if err = query.SetColumns(fc...); err != nil {
		return
	}

	ff, err := findsFrom(momentsJoin(findsAlias))
	if err != nil {
		return
	}
	if err = query.SetFroms(ff); err != nil {
		return
	}

	mwu, err := momentsWhereUserID(me)
	if err != nil {
		return
	}
	if err = query.SetWheres(mwu); err != nil {
		return
	}

	r, err = process(query, leftResults)
	return
}

func UserFound(me string) (r ResultsMap, err error) {
	if util.IsEmpty(me) {
		return r, ErrorVariableEmpty
	}

	query, err := momentsQuery()
	if err != nil {
		return
	}

	fdc, err := findsFindDateColumn()
	if err != nil {
		return
	}
	if err = query.SetColumns(fdc); err != nil {
		return
	}

	ff, err := findsFrom(momentsJoin(findsAlias))
	if err != nil {
		return
	}
	if err = query.SetFroms(ff); err != nil {
		return
	}

	fwu, err := findsWhereUserIDFound(me)
	if err != nil {
		return
	}
	if err = query.SetWheres(fwu...); err != nil {
		return
	}

	r, err = process(query, foundResults)
	return
}

var (
	ErrorQueryStringEmpty = errors.New("Empty string passed into queryString parameter.")
)

func process(query *dba.Query, rowHandler func(*sql.Rows) (ResultsMap, error)) (r ResultsMap, err error) {
	if util.IsEmpty(query) {
		return r, ErrorVariableEmpty
	}

	queryString, err := query.Build()
	if err != nil {
		return
	}

	c := dba.OpenConn()
	defer c.Db.Close()

	args, err := query.Args()
	if err != nil {
		return
	}
	rows, err := c.Db.Query(queryString, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	r, err = rowHandler(rows)
	if err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		return
	}

	return
}

func momentsResults(rows *sql.Rows) (rm ResultsMap, err error) {

	m := new(MomentsRow)
	md := new(MediaRow)
	var createDate string
	dest := []interface{}{
		&m.userID,
		&m.public,
		&m.hidden,
		&createDate,
		&m.momentID,
		&m.latitude,
		&m.longitude,
		&md.message,
		&md.mType,
		&md.dir,
	}

	rm = make(ResultsMap, 0)
	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			return
		}

		m.createDate, err = dba.ParseDateTime2(createDate)
		if err != nil {
			return
		}

		if rm.exists(m) {
			if err = rm.append(m, md); err != nil {
				return
			}
		} else {
			if err = rm.add(m, md); err != nil {
				return
			}
		}
	}
	return
}

func leftResults(rows *sql.Rows) (rm ResultsMap, err error) {
	m := new(MomentsRow)
	md := new(MediaRow)
	f := new(FindsRow)
	var createDate, findDate string
	dest := []interface{}{
		&m.userID,
		&m.public,
		&m.hidden,
		&createDate,
		&m.momentID,
		&m.latitude,
		&m.longitude,
		&md.message,
		&md.mType,
		&md.dir,
		&f.userID,
		&findDate,
	}

	rm = make(ResultsMap, 0)
	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			return
		}
		m.createDate, err = dba.ParseDateTime2(createDate)
		if err != nil {
			return
		}
		f.findDate, err = dba.ParseDateTime2(findDate)
		if err != nil {
			return
		}

		if rm.exists(m) {
			if err = rm.append(m, md, f); err != nil {
				return
			}
		} else {
			if err = rm.add(m, md, f); err != nil {
				return
			}
		}
	}
	return
}

func foundResults(rows *sql.Rows) (rm ResultsMap, err error) {
	m := new(MomentsRow)
	md := new(MediaRow)
	f := new(FindsRow)
	var createDate, findDate string
	dest := []interface{}{
		&m.userID,
		&m.public,
		&m.hidden,
		&createDate,
		&m.momentID,
		&m.latitude,
		&m.longitude,
		&md.message,
		&md.mType,
		&md.dir,
		&findDate,
	}

	rm = make(ResultsMap, 0)
	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			return
		}
		m.createDate, err = dba.ParseDateTime2(createDate)
		if err != nil {
			return
		}
		f.findDate, err = dba.ParseDateTime2(findDate)
		if err != nil {
			return
		}

		if rm.exists(m) {
			if err = rm.append(m, md, f); err != nil {
				return
			}
		} else {
			if err = rm.add(m, md, f); err != nil {
				return
			}
		}
	}
	return

}

func lostResults(rows *sql.Rows) (rm ResultsMap, err error) {
	m := new(MomentsRow)
	dest := []interface{}{
		&m.momentID,
		&m.latitude,
		&m.longitude,
	}

	rm = make(ResultsMap, 0)
	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			return
		}
		newR := &ResultMap{
			moment: m,
		}
		rm[m.momentID] = newR
	}

	return
}
