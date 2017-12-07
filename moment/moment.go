package moment

import (
	"database/sql"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/minus5/gofreetds"
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

	// Datetime2 is the time.Time format this package uses to communicate DateTime2 values to Moment-Db.
	Datetime2 = "2006-01-02 15:04:05"
)

var (
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger

	connStr = os.Getenv("MomentDBConnStr")
	driver  = "mssql"
)

func init() {
	Logger := func(logType string) *log.Logger {
		file := "/home/tjp/go/log/moment.txt"
		f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		l := log.New(f, strings.ToUpper(logType)+": ", log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC|log.Lshortfile)
		return l
	}

	Info = Logger("info")
	Warn = Logger("warn")
	Error = Logger("error")
}

var (
	ErrorPrivateHiddenMoment = errors.New("m *MomentsRow cannot be both private and hidden")
	ErrorParameterEmpty      = errors.New("Parameter is empty.")
	ErrorFieldInvalid        = errors.New("A struct field is not in the state required.")
	ErrorTypeNotImplemented  = errors.New("Type switch does not handle this type.")
)

// MomentDB creates *db.Sql instance.
func DB() *sql.DB {
	db, err := sql.Open(driver, connStr)
	if err != nil {
		panic(err)
	}
	return db
}

type DbRunner interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
}

type DbRunnerTrans interface {
	DbRunner
	Begin() (*sql.Tx, error)
}

type MomentClient struct {
	err error
}

func (mc *MomentClient) Err() error {
	return mc.err
}

type Finder interface {
	FindPublic(DbRunner, *FindsRow) (int64, error)
	FindPrivate(DbRunner, *FindsRow) error
}

type Sharer interface {
	Share(DbRunnerTrans, *SharesRow, []*RecipientsRow) error
}

type Creater interface {
	CreatePublic(DbRunnerTrans, *MomentsRow, []*MediaRow) error
	CreatePrivate(DbRunnerTrans, *MomentsRow, []*MediaRow, []*FindsRow) error
}

type Modifier interface {
	Finder
	Sharer
	Creater
}

// FindPublic inserts a FindsRow into the [Moment-Db].[moment].[Finds] table with Found=true.
func (mc *MomentClient) FindPublic(db DbRunner, f *FindsRow) (cnt int64, err error) {
	if err = f.isFound(); err != nil {
		Error.Println(err)
		return
	}

	fs := []*FindsRow{
		f,
	}
	cnt, err = insert(db, fs)
	if err != nil {
		Error.Println(err)
	}
	return
}

// FindPrivate updates a FindsRow in the [Moment-Db].[moment].[Finds] by setting Found=true.
func (mc *MomentClient) FindPrivate(db DbRunner, f *FindsRow) (err error) {
	if err = f.isFound(); err != nil {
		Error.Println(err)
		return
	}

	if err = update(db, f); err != nil {
		Error.Println(err)
	}
	return
}

// Share is an exported package that allows the insertion of a
// Shares instance into the [Moment-Db].[moment].[Shares] table.
func (mc *MomentClient) Share(db DbRunnerTrans, s *SharesRow, rs []*RecipientsRow) (err error) {
	if len(rs) == 0 || s == nil {
		Error.Println(ErrorParameterEmpty)
		err = ErrorParameterEmpty
		return
	}

	tx, err := db.Begin()
	if err != nil {
		Error.Println(err)
		return
	}
	defer func() {
		if err != nil {
			if txerr := tx.Rollback(); txerr != nil {
				Error.Println(txerr)
			}
			Error.Println(err)
			return
		}
		tx.Commit()
	}()

	id, err := insert(tx, s)
	if err != nil {
		Error.Println(err)
	}
	s.sharesID = id

	for _, r := range rs {
		r.setSharesID(id)
		if err = r.err; err != nil {
			return
		}
	}
	if _, err = insert(tx, rs); err != nil {
		return
	}
	return
}

var ErrorMediaPointerNil = errors.New("md *Media is nil.")

// CreatePublic creates a row in [Moment-Db].[moment].[Moments] where Public=true.
func (mc *MomentClient) CreatePublic(db DbRunnerTrans, m *MomentsRow, ms []*MediaRow) (err error) {
	if len(ms) == 0 || m == nil {
		Error.Println(ErrorParameterEmpty)
		return ErrorParameterEmpty
	}

	tx, err := db.Begin()
	if err != nil {
		Error.Println(err)
		return
	}
	defer func() {
		if err != nil {
			if txerr := tx.Rollback(); txerr != nil {
				Error.Println(txerr)
			}
			Error.Println(err)
			return
		}
		tx.Commit()
	}()

	var mID int64
	if mID, err = insert(tx, m); err != nil {
		return
	}
	m.momentID = mID

	for _, mr := range ms {
		mr.setMomentID(m.momentID)
		if err = mr.err; err != nil {
			return
		}
	}
	if _, err = insert(tx, ms); err != nil {
		return
	}

	return
}

var ErrorFindsPointerNil = errors.New("finds *Finds pointer is empty.")

// CreatePrivate creates a MomentsRow in [Moment-Db].[moment].[Moments] where Public=true
// and creates Finds in [Moment-Db].[moment].[Finds].
func (mc *MomentClient) CreatePrivate(db DbRunnerTrans, m *MomentsRow, ms []*MediaRow, fs []*FindsRow) (err error) {
	if m == nil || len(ms) == 0 || len(fs) == 0 {
		Error.Println(ErrorParameterEmpty)
		return ErrorParameterEmpty
	}

	tx, err := db.Begin()
	if err != nil {
		Error.Println(err)
		return
	}
	defer func() {
		if err != nil {
			if txerr := tx.Rollback(); txerr != nil {
				Error.Println(txerr)
			}
			Error.Println(err)
			return
		}
		tx.Commit()
	}()

	mID, err := insert(tx, m)
	if err != nil {
		Error.Println(err)
		return
	}
	m.momentID = mID

	for _, md := range ms {
		md.setMomentID(m.momentID)
		if md.err != nil {
			Error.Println(md.err)
			return md.err
		}
	}

	for _, f := range fs {
		f.setMomentID(m.momentID)
		if f.err != nil {
			Error.Println(f.err)
			return f.err
		}
	}

	if _, err = insert(tx, ms); err != nil {
		Error.Println(err)
		return
	}
	if _, err = insert(tx, fs); err != nil {
		Error.Println(err)
		return
	}

	return
}

func insert(db DbRunner, i interface{}) (resVal int64, err error) {
	var insert sq.InsertBuilder
	switch v := i.(type) {
	case []*FindsRow:
		insert = sq.
			Insert(momentSchema+"."+finds).
			Columns(momentID, userID, found, findDate)
		for _, f := range v {
			insert = insert.Values(f.momentID, f.userID, f.found, f.findDate)
		}
	case []*RecipientsRow:
		insert = sq.
			Insert(schRecipients).
			Columns(sharesID, all, recipientID)
		for _, r := range v {
			insert = insert.Values(r.sharesID, r.all, r.recipientID)
		}
	case []*MediaRow:
		insert = sq.
			Insert(momentSchema+"."+media).
			Columns(momentID, message, mtype, dir)
		for _, md := range v {
			insert = insert.Values(md.momentID, md.message, md.mType, md.dir)
		}
	case *MomentsRow:
		insert = sq.
			Insert(momentSchema+"."+moments).
			Columns(userID, latStr, longStr, public, hidden, createDate).
			Values(v.userID, v.latitude, v.longitude, v.public, v.hidden, v.createDate)
	case *SharesRow:
		insert = sq.
			Insert(schShares).
			Columns(momentID, userID).
			Values(v.momentID, v.userID)
	default:
		return resVal, ErrorTypeNotImplemented
	}

	res, err := insert.RunWith(db).Exec()
	if err != nil {
		Error.Println(err)
		return
	}

	switch i.(type) {
	case *SharesRow:
		resVal, err = res.LastInsertId()
	case *MomentsRow:
		resVal, err = res.LastInsertId()
	default:
		resVal, err = res.RowsAffected()
	}
	if err != nil {
		Error.Println(err)
	}

	return
}

func update(db DbRunner, i interface{}) (err error) {
	var query sq.UpdateBuilder
	switch v := i.(type) {
	case *FindsRow:
		query = sq.Update(momentSchema+"."+finds).
			Set(found, v.found).
			Set(findDate, v.findDate).
			Where(sq.Eq{momentID: v.momentID}).
			Where(sq.Eq{userID: v.userID})
	default:
		return ErrorTypeNotImplemented
	}

	_, err = query.RunWith(db).Exec()
	if err != nil {
		Error.Println(err)
	}
	return
}

type Newer interface {
	NewMomentsRow(*Location, string, bool, bool, *time.Time) *MomentsRow
	NewLocation(float32, float32) *Location
	NewMediaRow(int64, string, uint8, string) *MediaRow
	NewFindsRow(int64, string, bool, *time.Time) *FindsRow
	NewSharesRow(int64, int64, string) *SharesRow
	NewRecipientsRow(int64, bool, string) *RecipientsRow
}

// NewMoment is a constructor for the MomentsRow struct.
func (mc *MomentClient) NewMomentsRow(l *Location, uID string, p bool, h bool, c *time.Time) (m *MomentsRow) {
	if mc.err != nil {
		return
	}

	m = new(MomentsRow)

	m.setLocation(l)
	m.setCreateDate(c)
	m.setUserID(uID)
	if m.err != nil {
		Error.Println(m.err)
		mc.err = m.err
		return
	}

	m.hidden = h
	m.public = p
	if m.hidden && !m.public {
		Error.Println(ErrorPrivateHiddenMoment)
		mc.err = ErrorPrivateHiddenMoment
		return
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
	return fmt.Sprintf("id: %v, userID: %v, Location: %v, public: %v, hidden: %v, creatDate: %v", m.momentID, m.userID, m.Location, m.public, m.hidden, m.createDate)
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
func (mc *MomentClient) NewMediaRow(mID int64, m string, mType uint8, d string) (mr *MediaRow) {
	if mc.err != nil {
		return
	}

	mr = new(MediaRow)

	mr.setMomentID(mID)
	mr.setMessage(m)
	mr.setmType(mType)
	if mr.err != nil {
		Error.Println(mr.err)
		mc.err = mr.err
		return
	}

	mr.dir = d

	if mr.mType == DNE && mr.dir != "" {
		Error.Println(ErrorMediaDNE)
		mc.err = ErrorMediaDNE
		return
	}
	if mr.mType != DNE && mr.dir == "" {
		Error.Println(ErrorMediaExistsDirDNE)
		mc.err = ErrorMediaExistsDirDNE
		return
	}

	return
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
	return fmt.Sprintf("momentID: %v, mType: %v, message: \"%v\", dir: \"%v\"", m.momentID, m.mType, m.message, m.dir)
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
func (mc *MomentClient) NewFindsRow(mID int64, uID string, f bool, fd *time.Time) (fr *FindsRow) {
	if mc.err != nil {
		return
	}

	fr = new(FindsRow)

	fr.setMomentID(mID)
	fr.setUserID(uID)
	fr.setFindDate(fd)
	if fr.err != nil {
		Error.Println(fr.err)
		mc.err = fr.err
		return
	}

	fr.found = f

	emptyTime := time.Time{}
	if fr.found && *fr.findDate == emptyTime {
		Error.Println(ErrorFoundEmptyFindDate)
		mc.err = ErrorFoundEmptyFindDate
		return
	}
	if !fr.found && *fr.findDate != emptyTime {
		Error.Println(ErrorNotFoundFindDateExists)
		mc.err = ErrorNotFoundFindDateExists
		return
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
	return fmt.Sprintf("momentID: %v, userID: %v, found: %v, findDate: %v",
		f.momentID,
		f.userID,
		f.found,
		f.findDate)
}

func (f *FindsRow) isFound() error {
	if f.found == true && f.findDate != new(time.Time) {
		return nil
	}
	return ErrorFieldInvalid
}
func (f *FindsRow) setFindDate(fd *time.Time) {
	if f.err != nil {
		return
	}
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

// NewShare is a constructor for the SharesRow struct.
func (mc *MomentClient) NewSharesRow(id int64, mID int64, uID string) (s *SharesRow) {
	if mc.err != nil {
		return
	}

	s = new(SharesRow)

	s.setID(id)
	s.setMomentID(mID)
	s.setUserID(uID)

	if s.err != nil {
		Error.Println(s.err)
		mc.err = s.err
		return
	}

	return
}

// SharesRow is a row in the [Moment-Db].[moment].[Shares] table.
type SharesRow struct {
	sID
	mID
	uID
	err error
}

// String returns a string representation of a SharesRow instance.
func (s SharesRow) String() string {
	return fmt.Sprintf("ID: %v, momentID: %v, userID: %v",
		s.sharesID,
		s.momentID,
		s.userID)
}

func (s *SharesRow) setID(id int64) {
	if s.err != nil {
		return
	}
	s.err = s.sID.setSharesID(id)
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

// RecipientRow is a row in the [Moment-Db].[moment].[Shares] table.
type RecipientsRow struct {
	sID
	all         bool
	recipientID string
	err         error
}

var ErrorAllRecipientExists = errors.New("s.all=true, therefore s.recipientID must be \"\"")
var ErrorNotAllRecipientDNE = errors.New("s.all=false, therefore s.recipientID must be set")

func (mc *MomentClient) NewRecipientsRow(sharesID int64, all bool, recipientID string) (r *RecipientsRow) {
	if mc.err != nil {
		return
	}
	r = new(RecipientsRow)

	r.setSharesID(sharesID)
	r.setRecipientID(recipientID)
	r.all = all

	if r.all && len(r.recipientID) > 0 {
		Error.Println(ErrorAllRecipientExists)
		mc.err = ErrorAllRecipientExists
		return
	}
	if !r.all && len(r.recipientID) == 0 {
		Error.Println(ErrorNotAllRecipientDNE)
		mc.err = ErrorNotAllRecipientDNE
		return
	}

	return r
}

func (r *RecipientsRow) setSharesID(id int64) {
	if r.err != nil {
		return
	}
	r.err = r.sID.setSharesID(id)
}

func (r *RecipientsRow) setRecipientID(u string) {
	if r.err != nil {
		return
	}
	if err := checkUserIDLong(u); err != nil {
		r.err = err
		return
	}
	r.recipientID = u
}

var ErrorLatitude = errors.New("Latitude must be between -180 and 180.")
var ErrorLongitude = errors.New("Longitude must be between -90 and 90.")

// NewLocation is a constructor for the Location struct.
func (mc *MomentClient) NewLocation(lat float32, long float32) (l *Location) {
	if mc.err != nil {
		return
	}

	l = new(Location)

	l.setLatitude(lat)
	l.setLongitude(long)
	if l.err != nil {
		Error.Println(l.err)
		mc.err = l.err
		return
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
	return fmt.Sprintf("latitude: %v, longitude: %v", l.latitude, l.longitude)
}

// setLatitude ensures that the values of l is between minLat and maxLat.
func (l *Location) setLatitude(lat float32) {
	if l.err != nil {
		return
	}
	if lat < minLat || lat > maxLat {
		l.err = ErrorLatitude
		return
	}
	l.latitude = lat
	return
}

// setLongitude ensures that the values of l is between minLong and maxLong.
func (l *Location) setLongitude(long float32) {
	if l.err != nil {
		return
	}
	if long < minLong || long > maxLong {
		l.err = ErrorLongitude
		return
	}
	l.longitude = long
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

type sID struct {
	sharesID int64
}

func (s *sID) setSharesID(id int64) (err error) {
	if err = checkSharesID(id); err != nil {
		return
	}
	s.sharesID = id
	return
}

var ErrorTimePtrNil = errors.New("t *time.Time is set to nil")

// checkTime ensures that the value of t is a valid address.
func checkTime(t *time.Time) (err error) {
	if t == nil {
		return ErrorTimePtrNil
	}
	return
}

var ErrorMomentID = errors.New("momentID invalid")

// checkMomentID ensures that id is greater 0.
func checkMomentID(id int64) (err error) {
	if id < 0 {
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

var ErrorSharesID = errors.New("sharesID invalid")

// checkSharesID ensures that the id is greater than 0, and returns ErrorSharesID on error.
func checkSharesID(id int64) (err error) {
	if id < 0 {
		return ErrorSharesID
	}
	return
}

var ErrorFoundFalseFindDateNil = errors.New("A found row must have f.found=true and f.findDate=*time.Time{}")

const (
	momentsAlias    = "m"
	mediaAlias      = "md"
	sharesAlias     = "s"
	findsAlias      = "f"
	recipientsAlias = "r"

	noJoin  = ""
	noAlias = ""
)

type momentsType int64

const (
	Public momentsType = iota
	Hidden
	Shared
	Found
	Left

	momentSchema = "[moment]"

	moments    = "[Moments]"
	finds      = "[Finds]"
	media      = "[Media]"
	shares     = "[Shares]"
	recipients = "[Recipients]"

	schMoments    = momentSchema + "." + moments
	schFinds      = momentSchema + "." + finds
	schMedia      = momentSchema + "." + media
	schShares     = momentSchema + "." + shares
	schRecipients = momentSchema + "." + recipients

	iD         = "[ID]"
	momentID   = "[MomentID]"
	userID     = "[UserID]"
	latStr     = "[Latitude]"
	longStr    = "[Longitude]"
	public     = "[Public]"
	hidden     = "[Hidden]"
	createDate = "[CreateDate]"

	miD         = momentsAlias + "." + iD
	mUserID     = momentsAlias + "." + userID
	mLat        = momentsAlias + "." + latStr
	mLong       = momentsAlias + "." + longStr
	mPublic     = momentsAlias + "." + public
	mHidden     = momentsAlias + "." + hidden
	mCreateDate = momentsAlias + "." + createDate

	findDate = "[FindDate]"
	found    = "[Found]"

	fMomentID = findsAlias + "." + momentID
	fUserID   = findsAlias + "." + userID
	fFindDate = findsAlias + "." + findDate
	fFound    = findsAlias + "." + found

	siD       = sharesAlias + "." + iD
	sMomentID = sharesAlias + "." + momentID
	sUserID   = sharesAlias + "." + userID

	sharesID    = "[SharesID]"
	recipientID = "[RecipientID]"
	all         = "[All]"

	rSharesID    = recipientsAlias + "." + sharesID
	rRecipientID = recipientsAlias + "." + recipientID
	rAll         = recipientsAlias + "." + all

	message = "[Message]"
	mtype   = "[Type]"
	dir     = "[Dir]"

	mdMomentID = mediaAlias + "." + momentID
	mdMessage  = mediaAlias + "." + message
	mdType     = mediaAlias + "." + mtype
	mdDir      = mediaAlias + "." + dir
)

type Moment struct {
	momentID int64
	userID   string
	public   bool
	hidden   bool
	Location
	createDate *time.Time
	media      []*MediaRow
	finds      []*FindsRow
	shares     []*SharesRow
}

func (m Moment) String() string {
	s := fmt.Sprintf("\nmomentID: %v\nuserID: %s\npublic: %v\nhidden: %v\nLocation: %v\ncreateDate: %v\n",
		m.momentID,
		m.userID,
		m.public,
		m.hidden,
		m.Location,
		m.createDate)

	s += "media:\n"
	for i, md := range m.media {
		s += fmt.Sprintf("\t%v: %v\n", i, md)
	}

	s += "finds:\n"
	for i, f := range m.finds {
		s += fmt.Sprintf("\t%v: %v\n", i, f)
	}

	s += "shares:\n"
	for i, sh := range m.shares {
		s += fmt.Sprintf("\t%v: %v\n", i, sh)
	}
	return s
}

type Client interface {
	LocationSelector
	UserSelector
	Modifier
	Newer
	Err() error
}

type LocationSelector interface {
	LocationShared(DbRunner, *Location, string) ([]*Moment, error)
	LocationPublic(DbRunner, *Location) ([]*Moment, error)
	LocationHidden(DbRunner, *Location) ([]*Moment, error)
	LocationLost(DbRunner, *Location, string) ([]*Moment, error)
}

func (mc *MomentClient) LocationShared(db DbRunner, l *Location, me string) ([]*Moment, error) {
	if l == nil || me == "" {
		Error.Println(ErrorParameterEmpty)
		return nil, ErrorParameterEmpty
	}

	query := sq.
		Select(
			miD,
			mLat,
			mLong,
			mdMessage,
			mdType,
			mdDir,
			mCreateDate,
			mUserID,
			mPublic,
			mHidden).
		From(schMoments+" "+momentsAlias).
		Join(schMedia+" "+mediaAlias+" ON "+mdMomentID+" = "+miD).
		Join(schShares+" "+sharesAlias+" ON "+sMomentID+" = "+miD).
		Join(schRecipients+" "+recipientsAlias+" ON "+rSharesID+" = "+siD).
		Where(mLat+" BETWEEN ? AND ?", l.latitude-1, l.latitude+1).
		Where(mLong+" BETWEEN ? AND ?", l.longitude-1, l.longitude+1).
		Where("("+rRecipientID+" = ? OR "+rAll+" = 1)", me)

	return mc.selectMoments(db, query)
}

func (mc *MomentClient) LocationPublic(db DbRunner, l *Location) ([]*Moment, error) {
	if l == nil {
		Error.Println(ErrorParameterEmpty)
		return nil, ErrorParameterEmpty
	}

	query := sq.
		Select(
			miD,
			mLat,
			mLong,
			mdMessage,
			mdType,
			mdDir,
			mCreateDate,
			mUserID).
		From(schMoments+" "+momentsAlias).
		Join(schMedia+" "+mediaAlias+" ON "+mdMomentID+" = "+miD).
		Where(mLat+" BETWEEN ? AND ?", l.latitude-1, l.latitude+1).
		Where(mLong+" BETWEEN ? AND ?", l.longitude-1, l.longitude+1).
		Where(mPublic + " = true").
		Where(mHidden + " = false")

	return mc.selectPublicMoments(db, query)
}

func (mc *MomentClient) LocationHidden(db DbRunner, l *Location) ([]*Moment, error) {
	if l == nil {
		Error.Println(ErrorParameterEmpty)
		return nil, ErrorParameterEmpty
	}

	query := sq.
		Select(
			miD,
			mLat,
			mLong).
		From(schMoments+" "+momentsAlias).
		Where(mLat+" BETWEEN ? AND ?", l.latitude-1, l.latitude+1).
		Where(mLong+" BETWEEN ? AND ?", l.longitude-1, l.longitude+1).
		Where(mPublic + " = true").
		Where(mHidden + " = true")

	return mc.selectLostMoments(db, query)
}

func (mc *MomentClient) LocationLost(db DbRunner, l *Location, me string) ([]*Moment, error) {
	if l == nil || me == "" {
		Error.Println(ErrorParameterEmpty)
		return nil, ErrorParameterEmpty
	}

	query := sq.
		Select(
			miD,
			mLat,
			mLong).
		From(schMoments+" "+momentsAlias).
		Join(schFinds+" "+findsAlias+" ON "+fMomentID+" = "+miD).
		Where(mLat+" BETWEEN ? AND ?", l.latitude-1, l.latitude+1).
		Where(mLong+" BETWEEN ? AND ?", l.longitude-1, l.longitude+1).
		Where(mPublic+" = false").
		Where(mHidden+" = false").
		Where(fUserID+" = ?", me)

	return mc.selectLostMoments(db, query)
}

type UserSelector interface {
	UserShared(DbRunner, string, string) ([]*Moment, error)
	UserLeft(DbRunner, string) ([]*Moment, error)
	UserFound(DbRunner, string) ([]*Moment, error)
}

func (mc *MomentClient) UserShared(db DbRunner, you string, me string) ([]*Moment, error) {
	if me == "" || you == "" {
		Error.Println(ErrorParameterEmpty)
		return nil, ErrorParameterEmpty
	}

	query := sq.
		Select(
			miD,
			mLat,
			mLong,
			mdMessage,
			mdType,
			mdDir,
			mCreateDate,
			mUserID,
			mPublic,
			mHidden).
		From(schMoments+" "+momentsAlias).
		Join(schMedia+" "+mediaAlias+" ON "+mdMomentID+" = "+miD).
		Join(schShares+" "+sharesAlias+" ON "+sMomentID+" = "+miD).
		Join(schRecipients+" "+recipientsAlias+" ON "+rSharesID+" = "+siD).
		Where(sUserID+" = ?", you).
		Where("("+rRecipientID+" = ? OR "+rAll+" = true)", me)

	return mc.selectMoments(db, query)
}
func (mc *MomentClient) UserLeft(db DbRunner, me string) (rs []*Moment, err error) {
	if me == "" {
		Error.Println(ErrorParameterEmpty)
		return nil, ErrorParameterEmpty
	}

	query := sq.
		Select(
			miD,
			mLat,
			mLong,
			mdMessage,
			mdType,
			mdDir,
			mCreateDate,
			mPublic,
			mHidden,
			fUserID,
			fFindDate).
		From(schMoments+" "+momentsAlias).
		Join(schMedia+" "+mediaAlias+" ON "+mdMomentID+" = "+miD).
		Join(schFinds+" "+findsAlias+" ON "+fMomentID+" = "+miD).
		Where(mUserID+" = ?", me)

	return mc.selectLeftMoments(db, query)
}

func (mc *MomentClient) UserFound(db DbRunner, me string) ([]*Moment, error) {
	if me == "" {
		Error.Println(ErrorParameterEmpty)
		return nil, ErrorParameterEmpty
	}

	query := sq.
		Select(
			miD,
			mLat,
			mLong,
			mdMessage,
			mdType,
			mdDir,
			mCreateDate,
			mUserID,
			mPublic,
			mHidden,
			fFindDate).
		From(schMoments+" "+momentsAlias).
		Join(schMedia+" "+mediaAlias+" ON "+mdMomentID+" = "+miD).
		Join(schFinds+" "+findsAlias+" ON "+fMomentID+" = "+miD).
		Where(fUserID+" = ?", me).
		Where(fFound + " = true")

	return mc.selectFoundMoments(db, query)
}

func (mc *MomentClient) selectMoments(db DbRunner, query sq.SelectBuilder) (rs []*Moment, err error) {

	rows, err := query.RunWith(db).Query()
	if err != nil {
		Error.Println(err)
		return
	}
	defer rows.Close()

	m := new(MomentsRow)
	md := new(MediaRow)
	//	var createDate string
	dest := []interface{}{
		&m.momentID,
		&m.latitude,
		&m.longitude,
		&md.message,
		&md.mType,
		&md.dir,
		&m.createDate,
		&m.userID,
		&m.public,
		&m.hidden,
	}

	rm := make(map[int64]*Moment)
	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			Error.Println(err)
			return
		}

		if r, ok := rm[m.momentID]; !ok {
			r = &Moment{
				momentID:   m.momentID,
				userID:     m.userID,
				public:     m.public,
				hidden:     m.hidden,
				Location:   Location{latitude: m.latitude, longitude: m.longitude},
				createDate: m.createDate,
				media:      []*MediaRow{&MediaRow{message: md.message, mType: md.mType, dir: md.dir}},
			}
			rm[m.momentID] = r
		} else {
			r.media = append(r.media, &MediaRow{message: md.message, mType: md.mType, dir: md.dir})
		}
	}
	if err = rows.Err(); err != nil {
		Error.Println(err)
		return
	}

	rs = momentMaptoSlice(rm)

	return

}

func (mc *MomentClient) selectPublicMoments(db DbRunner, query sq.SelectBuilder) (rs []*Moment, err error) {
	rows, err := query.RunWith(db).Query()
	if err != nil {
		Error.Println(err)
		return
	}
	defer rows.Close()

	m := new(MomentsRow)
	md := new(MediaRow)
	dest := []interface{}{
		&m.momentID,
		&m.latitude,
		&m.longitude,
		&md.message,
		&md.mType,
		&md.dir,
		&m.createDate,
		&m.userID,
	}

	rm := make(map[int64]*Moment)
	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			Error.Println(err)
			return
		}

		if r, ok := rm[m.momentID]; !ok {
			r = &Moment{
				momentID: m.momentID,
				userID:   m.userID,
				Location: Location{latitude: m.latitude, longitude: m.longitude},
				media:    []*MediaRow{&MediaRow{message: md.message, mType: md.mType, dir: md.dir}},
			}
			rm[m.momentID] = r
		} else {
			r.media = append(r.media, &MediaRow{message: md.message, mType: md.mType, dir: md.dir})
		}
	}
	if err = rows.Err(); err != nil {
		Error.Println(err)
		return
	}

	rs = momentMaptoSlice(rm)

	return
}

func (mc *MomentClient) selectLostMoments(db DbRunner, query sq.SelectBuilder) (rs []*Moment, err error) {
	rows, err := query.RunWith(db).Query()
	if err != nil {
		Error.Println(err)
		return
	}
	defer rows.Close()

	m := new(MomentsRow)
	dest := []interface{}{
		&m.momentID,
		&m.latitude,
		&m.longitude,
	}

	rs = make([]*Moment, 0)
	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			Error.Println(err)
			return
		}
		rs = append(rs,
			&Moment{
				momentID: m.momentID,
				Location: Location{latitude: m.latitude, longitude: m.longitude},
			})
	}
	if err = rows.Err(); err != nil {
		Error.Println(err)
		return
	}
	return
}

func (mc *MomentClient) selectLeftMoments(db DbRunner, query sq.SelectBuilder) (rs []*Moment, err error) {
	rows, err := query.RunWith(db).Query()
	if err != nil {
		Error.Println(err)
		return
	}
	defer rows.Close()

	m := new(MomentsRow)
	md := new(MediaRow)
	f := new(FindsRow)
	dest := []interface{}{
		&m.momentID,
		&m.latitude,
		&m.longitude,
		&md.message,
		&md.mType,
		&md.dir,
		&m.createDate,
		&m.public,
		&m.hidden,
		&f.userID,
		&f.findDate,
	}

	var mdMap, fMap map[string]bool
	rm := make(map[int64]*Moment)

	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			Error.Println(err)
			return
		}

		if r, ok := rm[m.momentID]; !ok {
			r = &Moment{
				momentID:   m.momentID,
				userID:     m.userID,
				public:     m.public,
				hidden:     m.hidden,
				Location:   Location{latitude: m.latitude, longitude: m.longitude},
				createDate: m.createDate,
				media:      []*MediaRow{&MediaRow{message: md.message, mType: md.mType, dir: md.dir}},
				finds:      []*FindsRow{&FindsRow{uID: uID{userID: f.userID}, findDate: f.findDate}},
			}
			rm[m.momentID] = r

			mdMap = make(map[string]bool)
			fMap = make(map[string]bool)

			mdMap[md.dir] = true
			fMap[f.userID] = true

		} else {
			if _, ok = mdMap[md.dir]; !ok {
				r.media = append(r.media, &MediaRow{message: md.message, mType: md.mType, dir: md.dir})
				mdMap[md.dir] = true
			}

			if _, ok = fMap[f.userID]; !ok {
				r.finds = append(r.finds, &FindsRow{uID: uID{userID: f.userID}, findDate: f.findDate})
				fMap[f.userID] = true
			}
		}

	}
	if err = rows.Err(); err != nil {
		Error.Println(err)
		return
	}

	rs = momentMaptoSlice(rm)

	return
}

func (mc *MomentClient) selectFoundMoments(db DbRunner, query sq.SelectBuilder) (rs []*Moment, err error) {
	rows, err := query.RunWith(db).Query()
	if err != nil {
		Error.Println(err)
		return
	}
	defer rows.Close()

	m := new(MomentsRow)
	md := new(MediaRow)
	f := new(FindsRow)

	dest := []interface{}{
		&m.momentID,
		&m.latitude,
		&m.longitude,
		&md.message,
		&md.mType,
		&md.dir,
		&m.createDate,
		&m.userID,
		&m.public,
		&m.hidden,
		&f.findDate,
	}

	rm := make(map[int64]*Moment)
	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			Error.Println(err)
			return
		}

		if r, ok := rm[m.momentID]; !ok {
			r = &Moment{
				momentID:   m.momentID,
				userID:     m.userID,
				public:     m.public,
				hidden:     m.hidden,
				Location:   Location{latitude: m.latitude, longitude: m.longitude},
				createDate: m.createDate,
				media: []*MediaRow{
					&MediaRow{message: md.message, mType: md.mType, dir: md.dir},
				},
				finds: []*FindsRow{
					&FindsRow{findDate: f.findDate},
				},
			}
			rm[m.momentID] = r
		} else {
			r.media = append(r.media, &MediaRow{message: md.message, mType: md.mType, dir: md.dir})
		}

	}
	if err = rows.Err(); err != nil {
		Error.Println(err)
		return
	}

	rs = momentMaptoSlice(rm)

	return
}

func momentMaptoSlice(m map[int64]*Moment) (s []*Moment) {
	for _, v := range m {
		s = append(s, v)
	}
	return
}
