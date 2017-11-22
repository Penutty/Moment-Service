package moment

import (
	"database/sql"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
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
func MomentDB() *sql.DB {
	db, err := sql.Open(driver, connStr)
	if err != nil {
		Error.Fatal(err)
	}
	return db
}

type MomentClient struct {
	err error
}

func (mc *MomentClient) Err() error {
	return mc.err
}

// FindPublic inserts a FindsRow into the [Moment-Db].[moment].[Finds] table with Found=true.
func (mc *MomentClient) FindPublic(db sq.BaseRunner, f *FindsRow) (cnt int64, err error) {
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
func (mc *MomentClient) FindPrivate(db sq.BaseRunner, f *FindsRow) (err error) {
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
func (mc *MomentClient) Share(db sq.BaseRunner, ss []*SharesRow) (cnt int64, err error) {
	if len(ss) == 0 {
		Error.Println(ErrorParameterEmpty)
		err = ErrorParameterEmpty
		return
	}

	cnt, err = insert(db, ss)
	if err != nil {
		Error.Println(err)
	}
	return
}

var ErrorMediaPointerNil = errors.New("md *Media is nil.")

// CreatePublic creates a row in [Moment-Db].[moment].[Moments] where Public=true.
func (mc *MomentClient) CreatePublic(db sq.BaseRunner, m *MomentsRow, ms []*MediaRow) (err error) {
	if len(ms) == 0 || m == nil {
		Error.Println(ErrorParameterEmpty)
		return ErrorParameterEmpty
	}

	var mID int64
	if mID, err = insert(db, m); err != nil {
		Error.Println(err)
		return
	}
	m.momentID = mID

	for _, mr := range ms {
		mr.setMomentID(m.momentID)
		if err = mr.err; err != nil {
			Error.Println(err)
			return
		}
	}
	if _, err = insert(db, ms); err != nil {
		Error.Println(err)
		return
	}

	return
}

var ErrorFindsPointerNil = errors.New("finds *Finds pointer is empty.")

// CreatePrivate creates a MomentsRow in [Moment-Db].[moment].[Moments] where Public=true
// and creates Finds in [Moment-Db].[moment].[Finds].
func (mc *MomentClient) CreatePrivate(db sq.BaseRunner, m *MomentsRow, ms []*MediaRow, fs []*FindsRow) (err error) {
	if m == nil || len(ms) == 0 || len(fs) == 0 {
		Error.Println(ErrorParameterEmpty)
		return ErrorParameterEmpty
	}

	mID, err := insert(db, m)
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

	if _, err = insert(db, ms); err != nil {
		Error.Println(err)
		return
	}
	if _, err = insert(db, fs); err != nil {
		Error.Println(err)
		return
	}

	return
}

func insert(db sq.BaseRunner, i interface{}) (resVal int64, err error) {
	var insert sq.InsertBuilder
	switch v := i.(type) {
	case []*FindsRow:
		insert = sq.
			Insert(momentSchema+"."+finds).
			Columns(momentID, userID, found, findDate)
		for _, f := range v {
			insert = insert.Values(f.momentID, f.userID, f.found, f.findDate)
		}
	case []*SharesRow:
		insert = sq.
			Insert(momentSchema+"."+shares).
			Columns(momentID, userID, all, recipientID)
		for _, s := range v {
			insert = insert.Values(s.momentID, s.userID, s.all, s.recipientID)
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
			Columns(iD, userID, latStr, longStr, public, hidden, createDate).
			Values(v.momentID, v.userID, v.latitude, v.longitude, v.public, v.hidden, v.createDate)
	default:
		return resVal, ErrorTypeNotImplemented
	}

	res, err := insert.RunWith(db).Exec()
	if err != nil {
		Error.Println(err)
		return
	}

	switch i.(type) {
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

func update(db sq.BaseRunner, i interface{}) (err error) {
	var query sq.UpdateBuilder
	switch v := i.(type) {
	case *FindsRow:
		sM := map[string]interface{}{found: v.found, findDate: v.findDate}
		wM := map[string]interface{}{momentID: v.momentID, userID: v.userID}
		query = query.Table(momentSchema + "." + finds).SetMap(sM).Where(wM)
	default:
		return ErrorTypeNotImplemented
	}

	_, err = query.RunWith(db).Exec()
	if err != nil {
		Error.Println(err)
	}
	return
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
	return fmt.Sprintf("momentID: %v\nmType: %v\nmessage: %v\ndir: %v\n", m.momentID, m.mType, m.message, m.dir)
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
	return fmt.Sprintf("momentID: %v\n"+
		"userID:   %v\n"+
		"found: 	  %v\n"+
		"findDate: %v\n",
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

var ErrorAllRecipientExists = errors.New("s.all=true, therefore s.recipientID must be \"\"")
var ErrorNotAllRecipientDNE = errors.New("s.all=false, therefore s.recipientID must be set")

// NewShare is a constructor for the SharesRow struct.
func (mc *MomentClient) NewSharesRow(mID int64, uID string, all bool, r string) (s *SharesRow) {
	if mc.err != nil {
		return
	}

	s = new(SharesRow)

	s.setMomentID(mID)
	s.setUserID(uID)
	s.setRecipientID(r)
	if s.err != nil {
		Error.Println(s.err)
		mc.err = s.err
		return
	}

	s.all = all

	if s.all && s.recipientID != "" {
		Error.Println(ErrorAllRecipientExists)
		mc.err = ErrorAllRecipientExists
		return
	}
	if !s.all && s.recipientID == "" {
		Error.Println(ErrorNotAllRecipientDNE)
		mc.err = ErrorNotAllRecipientDNE
		return
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
	return fmt.Sprintf("latitude: %v\nlongitude: %v\n", l.latitude, l.longitude)
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

var ErrorFoundFalseFindDateNil = errors.New("A found row must have f.found=true and f.findDate=*time.Time{}")

const (
	momentsAlias = "m"
	mediaAlias   = "md"
	sharesAlias  = "s"
	findsAlias   = "f"

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

	moments = "[Moments]"
	finds   = "[Finds]"
	media   = "[Media]"
	shares  = "[Shares]"

	iD         = "[ID]"
	momentID   = "[MomentID]"
	userID     = "[UserID]"
	latStr     = "[Latitude]"
	longStr    = "[Longitude]"
	public     = "[Public]"
	hidden     = "[Hidden]"
	createDate = "[CreateDate]"

	findDate = "[FindDate]"
	found    = "[Found]"

	recipientID = "[RecipientID]"
	all         = "[All]"

	message = "[Message]"
	mtype   = "[Type]"
	dir     = "[Dir]"
)

var (
	mLostColumns = []string{
		momentsAlias + "." + iD,
		momentsAlias + "." + latStr,
		momentsAlias + "." + longStr,
	}
	mTypeColumns = []string{
		momentsAlias + "." + public,
		momentsAlias + "." + hidden,
	}
	mUserID = []string{
		momentsAlias + "." + userID,
	}
	mCreateDate = []string{
		momentsAlias + "." + createDate,
	}

	onMoments = momentsAlias + "." + iD
)

func mLocationBetween(l *Location) map[string]interface{} {
	return map[string]interface{}{
		momentsAlias + "." + latStr + " BETWEEN ? AND ?":  []interface{}{l.latitude - 1, l.latitude + 1},
		momentsAlias + "." + longStr + " BETWEEN ? AND ?": []interface{}{l.longitude - 1, l.longitude + 1},
	}
}

func mPublicEquals(b bool) map[string]interface{} {
	return map[string]interface{}{
		momentsAlias + "." + public: b,
	}
}

func mHiddenEquals(b bool) map[string]interface{} {
	return map[string]interface{}{
		momentsAlias + "." + hidden: b,
	}
}

func mUserIDEquals(u string) map[string]interface{} {
	return map[string]interface{}{
		momentsAlias + "." + userID: u,
	}
}

var (
	mdColumns = []string{
		mediaAlias + "." + message,
		mediaAlias + "." + mtype,
		mediaAlias + "." + dir,
	}

	fUserID = []string{
		findsAlias + "." + userID,
	}

	fFindDate = []string{
		findsAlias + "." + findDate,
	}
)

func findsTbl(on string) string {
	f := momentSchema + "." + finds + " " + findsAlias
	if on != "" {
		f += " ON " + findsAlias + "." + momentID + " = " + on
	}
	return f
}

func mediaTbl(on string) string {
	md := momentSchema + "." + media + " " + mediaAlias
	if on != "" {
		md += " ON " + mediaAlias + "." + momentID + " = " + on
	}
	return md
}

func fFoundEquals(b bool) map[string]interface{} {
	return map[string]interface{}{
		findsAlias + "." + found: b,
	}
}

func fUserIDEquals(u string) map[string]interface{} {
	return map[string]interface{}{
		findsAlias + "." + userID: u,
	}
}

func sharesTbl(on string) string {
	s := momentSchema + "." + shares + " " + sharesAlias
	if on != "" {
		s += " ON " + sharesAlias + "." + momentID + " = " + on
	}
	return s
}

func sRecipientEquals(me string) map[string]interface{} {
	return map[string]interface{}{
		"(" + sharesAlias + "." + recipientID + " = ? OR " + sharesAlias + "." + all + " = 1)": me,
	}
}

func sUserIDEquals(you string) map[string]interface{} {
	return map[string]interface{}{
		sharesAlias + "." + userID: you,
	}
}

func momentsQuery(mt momentsType) sq.SelectBuilder {
	cSet := [][]string{
		mLostColumns,
		mdColumns,
		mCreateDate,
	}
	if mt != Left {
		cSet = append(cSet, mUserID)
	}
	if mt != Public {
		cSet = append(cSet, mTypeColumns)
	}

	var columns []string
	for _, c := range cSet {
		columns = append(columns, c...)
	}

	query := sq.
		Select().
		Columns(columns...).
		From(momentSchema + "." + moments + " " + momentsAlias).
		Join(mediaTbl(onMoments))

	return query
}

func lostQuery() sq.SelectBuilder {
	query := sq.
		Select().
		Columns(mLostColumns...).
		From(moments)

	return query
}

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

type Client interface {
	LocationSelector
	UserSelector
}

type LocationSelector interface {
	LocationShared(sq.BaseRunner, *Location, string) ([]*Moment, error)
	LocationPublic(sq.BaseRunner, *Location) ([]*Moment, error)
	LocationHidden(sq.BaseRunner, *Location) ([]*Moment, error)
	LocationLost(sq.BaseRunner, *Location, string) ([]*Moment, error)
}

func (mc *MomentClient) LocationShared(db sq.BaseRunner, l *Location, me string) ([]*Moment, error) {
	if l == nil || me == "" {
		Error.Println(ErrorParameterEmpty)
		return nil, ErrorParameterEmpty
	}

	query := momentsQuery(Shared)

	query = query.
		Join(sharesTbl(onMoments)).
		Where(mLocationBetween(l)).
		Where(sRecipientEquals(me))

	return mc.selectMoments(db, query)
}

func (mc *MomentClient) LocationPublic(db sq.BaseRunner, l *Location) ([]*Moment, error) {
	if l == nil {
		Error.Println(ErrorParameterEmpty)
		return nil, ErrorParameterEmpty
	}

	query := momentsQuery(Public)

	query = query.Where(mLocationBetween(l)).
		Where(mPublicEquals(true)).
		Where(mHiddenEquals(false))

	return mc.selectPublicMoments(db, query)
}

func (mc *MomentClient) LocationHidden(db sq.BaseRunner, l *Location) ([]*Moment, error) {
	if l == nil {
		Error.Println(ErrorParameterEmpty)
		return nil, ErrorParameterEmpty
	}

	query := lostQuery()

	query = query.Where(mLocationBetween(l)).
		Where(mPublicEquals(true)).
		Where(mHiddenEquals(true))

	return mc.selectLostMoments(db, query)
}

func (mc *MomentClient) LocationLost(db sq.BaseRunner, l *Location, me string) ([]*Moment, error) {
	if l == nil || me == "" {
		Error.Println(ErrorParameterEmpty)
		return nil, ErrorParameterEmpty
	}

	query := lostQuery()

	query = query.
		Join(findsTbl(onMoments)).
		Where(mLocationBetween(l)).
		Where(mPublicEquals(false)).
		Where(fUserIDEquals(me))

	return mc.selectLostMoments(db, query)
}

type UserSelector interface {
	UserShared(sq.BaseRunner, string, string) ([]*Moment, error)
	UserLeft(sq.BaseRunner, string) ([]*Moment, error)
	UserFound(sq.BaseRunner, string) ([]*Moment, error)
}

func (mc *MomentClient) UserShared(db sq.BaseRunner, me string, u string) ([]*Moment, error) {
	if me == "" || u == "" {
		Error.Println(ErrorParameterEmpty)
		return nil, ErrorParameterEmpty
	}

	query := momentsQuery(Shared)
	query = query.
		Join(sharesTbl(onMoments)).
		Where(sUserIDEquals(u)).
		Where(sRecipientEquals(me))

	return mc.selectMoments(db, query)
}
func (mc *MomentClient) UserLeft(db sq.BaseRunner, me string) (rs []*Moment, err error) {
	if me == "" {
		Error.Println(ErrorParameterEmpty)
		return nil, ErrorParameterEmpty
	}

	query := momentsQuery(Left)
	query = query.
		Columns(append(fUserID, fFindDate...)...).
		Join(findsTbl(onMoments)).
		Where(mUserIDEquals(me))

	return mc.selectLeftMoments(db, query)
}

func (mc *MomentClient) UserFound(db sq.BaseRunner, me string) ([]*Moment, error) {
	if me == "" {
		Error.Println(ErrorParameterEmpty)
		return nil, ErrorParameterEmpty
	}

	query := momentsQuery(Found)
	query = query.
		Columns(fFindDate...).
		From(findsTbl(onMoments)).
		Where(fUserIDEquals(me)).
		Where(fFoundEquals(true))

	return mc.selectFoundMoment(db, query)
}

func (mc *MomentClient) parseDatetime2(s string) *time.Time {
	t, err := time.Parse(Datetime2, findDate)
	if err != nil {
		Error.Println(err)
		mc.err = err
	}
	return &t
}

func (mc *MomentClient) selectMoments(db sq.BaseRunner, query sq.SelectBuilder) (rs []*Moment, err error) {
	rows, err := query.RunWith(db).Query()
	if err != nil {
		Error.Println(err)
		return
	}
	defer rows.Close()

	m := new(MomentsRow)
	md := new(MediaRow)
	var createDate string
	dest := []interface{}{
		&createDate,
		&m.userID,
		&m.public,
		&m.hidden,
		&m.momentID,
		&m.latitude,
		&m.longitude,
		&md.message,
		&md.mType,
		&md.dir,
	}

	rm := make(map[int64]*Moment)
	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			Error.Println(err)
			return
		}

		m.createDate = mc.parseDatetime2(createDate)
		if err = mc.Err(); err != nil {
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
	return

}

func (mc *MomentClient) selectPublicMoments(db sq.BaseRunner, query sq.SelectBuilder) (rs []*Moment, err error) {
	rows, err := query.RunWith(db).Query()
	if err != nil {
		Error.Println(err)
		return
	}
	defer rows.Close()

	m := new(MomentsRow)
	md := new(MediaRow)
	var createDate string
	dest := []interface{}{
		&createDate,
		&m.userID,
		&m.momentID,
		&m.latitude,
		&m.longitude,
		&md.message,
		&md.mType,
		&md.dir,
	}

	rm := make(map[int64]*Moment)
	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			Error.Println(err)
			return
		}

		m.createDate = mc.parseDatetime2(createDate)
		if err = mc.Err(); err != nil {
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
	return
}

func (mc *MomentClient) selectLostMoments(db sq.BaseRunner, query sq.SelectBuilder) (rs []*Moment, err error) {
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

func (mc *MomentClient) selectLeftMoments(db sq.BaseRunner, query sq.SelectBuilder) (rs []*Moment, err error) {
	rows, err := query.RunWith(db).Query()
	if err != nil {
		Error.Println(err)
		return
	}
	defer rows.Close()

	m := new(MomentsRow)
	md := new(MediaRow)
	f := new(FindsRow)
	var createDate, findDate string
	dest := []interface{}{
		&createDate,
		&m.public,
		&m.hidden,
		&m.momentID,
		&m.latitude,
		&m.longitude,
		&md.message,
		&md.mType,
		&md.dir,
		&f.userID,
		&findDate,
	}

	var mdMap, fMap map[string]bool
	rm := make(map[int64]*Moment)

	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			Error.Println(err)
			return
		}
		m.createDate = mc.parseDatetime2(createDate)
		f.findDate = mc.parseDatetime2(findDate)
		if err = mc.Err(); err != nil {
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
	return
}

func (mc *MomentClient) selectFoundMoment(db sq.BaseRunner, query sq.SelectBuilder) (rs []*Moment, err error) {
	rows, err := query.RunWith(db).Query()
	if err != nil {
		Error.Println(err)
		return
	}
	defer rows.Close()

	m := new(MomentsRow)
	md := new(MediaRow)
	f := new(FindsRow)
	var createDate, findDate string
	dest := []interface{}{
		&createDate,
		&m.userID,
		&m.public,
		&m.hidden,
		&m.momentID,
		&m.latitude,
		&m.longitude,
		&md.message,
		&md.mType,
		&md.dir,
		&findDate,
	}

	rm := make(map[int64]*Moment)
	for rows.Next() {
		if err = rows.Scan(dest); err != nil {
			Error.Println(err)
			return
		}

		m.createDate = mc.parseDatetime2(createDate)
		f.findDate = mc.parseDatetime2(findDate)
		if err = mc.Err(); err != nil {
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
	return
}
