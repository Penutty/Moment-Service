package moment

import (
	"database/sql"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
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
	ErrorVariableEmpty       = errors.New("Variable is empty.")
)

// NewMoment is a constructor for the MomentsRow struct.
func (mc *MomentClient) NewMoment(l *Location, uID string, p bool, h bool, c *time.Time) (m *MomentsRow) {
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
func (mc *MomentClient) NewMedia(mID int64, m string, mType uint8, d string) (mr *MediaRow) {
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
func (mc *MomentClient) NewFind(mID int64, uID string, f bool, fd *time.Time) (fr *FindsRow) {
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
		mc.err = ErrorFoundFalseFindDateNil
		return
	}

	return
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
func (mc *MomentClient) NewShare(mID int64, uID string, all bool, r string) (s *SharesRow) {
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
func (m *MediaRow) delete(db sq.BaseRunner) (rowCnt int64) {

	predicates := map[string]interface{}{
		"[MomentID]": m.momentID,
		"[Type]":     m.mType,
	}

	res, err := sq.Delete("[moment].[Media]").Where(predicates).RunWith(db).Exec()
	if err != nil {
		Error.Println(err)
		m.err = err
		return
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		Error.Println(err)
		m.err = err
		return
	}
	rowCnt = int64(cnt)

	return
}

// Media is a set of pointers to MediaRow instances.
type Media []*MediaRow

// insert inserts a set of MediaRow instances into the [Moment-Db].[moment].[Media] table.
func (mSet Media) insert(db sq.BaseRunner) (rowCnt int64, err error) {

	insert := sq.Insert("[moment].[Media]").Columns("[MomentID], [Message], [Type], [Dir]")
	for _, m := range mSet {
		insert = insert.Values(m.mID, m.message, m.mType, m.dir)
	}

	res, err := insert.RunWith(db).Exec()
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

// delete deletes a set of MediaRows from the [Moment-Db].[moment].[Media] table.
func (mSet Media) delete(db sq.BaseRunner) (rowCnt int64, err error) {

	var affCnt int64
	for _, m := range mSet {
		affCnt, err = m.delete(db)
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
func (f *FindsRow) FindPublic(db sq.BaseRunner) (cnt int64, err error) {
	if f.userID == "" || f.momentID == 0 {
		Error.Println(ErrorVariableEmpty)
		return 0, ErrorVariableEmpty
	}

	dt := time.Now().UTC()
	insert := sq.Insert("[moment].[Finds]").
		Columns("[MomentID]", "[UserID]", "[Found]", "[FindDate]").
		Values(f.momentID, f.userID, true, &dt)

	res, err := insert.RunWith(db).Exec()
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
func (f *FindsRow) FindPrivate(db sq.BaseRunner) (err error) {
	if f.userID == "" || f.momentID == 0 {
		Error.Println(ErrorVariableEmpty)
		return ErrorVariableEmpty
	}

	setM := map[string]interface{}{
		"[Found]":    1,
		"[FindDate]": time.Now().UTC(),
	}
	whereM := map[string]interface{}{
		"[UserID]":   f.userID,
		"[MomentID]": f.momentID,
	}

	update := sq.Update("[moment].[Finds]").
		SetMap(setM).
		Where(whereM)

	res, err := update.RunWith(db).Exec()
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
func (f *FindsRow) delete(db sq.BaseRunner, c *dba.Trans) (rowsAff int64, err error) {
	if c == nil {
		c = dba.OpenTx()
		defer func() { c.Close(err) }()
	}

	whereM := map[string]interface{}{
		"[MomentID]": f.momentID,
		"[UserID]":   f.userID,
	}
	delete := sq.Delete("[moment].[Finds]").Where(whereM)

	res, err := delete.RunWith(db).Exec()
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
func (fSet Finds) insert(db BaseRunner, c *dba.Trans) (rowCnt int64, err error) {
	if c == nil {
		c = dba.OpenTx()
		defer func() { c.Close(err) }()
	}

	insert := sq.Insert("[moment].[Finds]").Columns("[MomentID]", "[UserID]", "[Found]", "[FindDate]")
	for _, f := range fSet {
		insert = insert.Values(f.momentID, f.userID, f.found, f.findDate)
	}

	res, err := insert.RunWith(db).Exec()
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
func (s *SharesRow) delete(db sq.BaseRunner, c *dba.Trans) (affCnt int64, err error) {
	if c == nil {
		c = dba.OpenTx()
		defer func() { c.Close(err) }()
	}

	whereM := map[string]interface{}{
		"[MomentID]":    s.momentID,
		"[UserID]":      s.userID,
		"[RecipientID]": s.recipientID,
	}
	delete := sq.Delete("[moment].[Shares]").Where(whereM)

	res, err := delete.RunWith(db).Exec()
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
func (sSlice Shares) insert(db sq.BaseRunner) (rowCnt int64, err error) {

	insert := sq.Insert("[moment].[Shares]").Columns("[MomentID], [UserID], [All], [RecipientID]")
	for _, s := range sSlice {
		insert = insert.Values(s.momentID, s.userID, s.all, s.recipientID)
	}

	res, err := insert.RunWith(db).Exec()
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
func (mc *MomentClient) CreatePublic(db sq.BaseRunner, m *MomentsRow, media *Media) {
	switch {
	case media == nil:
		Error.Println(ErrorMediaPointerNil)
		mc.err = ErrorParameterEmpty
		return
	case m == nil:
		Error.Println(ErrorParameterEmpty)
		mc.err = ErrorParameterEmpty
		return
	}

	mID, err := m.insert(db, c)
	if err != nil {
		Error.Println(err)
		mc.err = err
		return
	}
	m.momentID = mID

	for _, p := range *media {
		p.setMomentID(m.momentID)
		if p.err != nil {
			Error.Println(err)
			mc.err = err
			return
		}
	}
	if _, err = media.insert(db, c); err != nil {
		Error.Println(err)
		mc.err = err
		return
	}

	return
}

var ErrorFindsPointerNil = errors.New("finds *Finds pointer is empty.")

// CreatePrivate creates a MomentsRow in [Moment-Db].[moment].[Moments] where Public=true
// and creates Finds in [Moment-Db].[moment].[Finds].
func (mc *MomentClient) CreatePrivate(db sq.BaseRunner, m *MomentsRow, media *Media, finds *Finds) {
	if media == nil {
		Error.Println(ErrorParameterEmpty)
		mc.err = ErrorParameterEmpty
		return
	}
	if finds == nil {
		Error.Println(ErrorParameterEmpty)
		mc.err = ErrorParameterEmpty
		return
	}

	mID := m.insert(db)
	if err := m.Err(); err != nil {
		Error.Println(err)
		mc.err = err
		return
	}
	m.momentID = mID

	for _, p := range *media {
		p.setMomentID(m.momentID)
	}

	for _, p := range *finds {
		p.setMomentID(m.momentID)
	}

	if _, err = media.insert(db); err != nil {
		Error.Println(err)
		mc.err = err
		return
	}
	if _, err = finds.insert(db); err != nil {
		Error.Println(err)
		mc.err = err
		return
	}

	return
}

// insert inserts a MomentsRow into the [Moment-Db].[moment].[Moments] table.
func (m *MomentsRow) insert(db sq.BaseRunner) (mID int64) {
	if c == nil {
		c = dba.OpenTx()
		defer func() { c.Close(err) }()
	}

	insert := sq.Insert("[moment].[Moments]").
		Columns("[UserID]", "[Latitude]", "[Longitude]", "[Public]", "[Hidden]", "[CreateDate]").
		Values(m.userID, m.latitude, m.longitude, m.public, m.hidden, m.createDate)

	res, err := insert.RunWith(db).Exec()
	if err != nil {
		Error.Println(err)
		m.err = err
		return
	}

	mID, err = res.LastInsertId()
	if err != nil {
		Error.Println(err)
		m.err = err
		return
	}
	m.momentID = mID

	return
}

// delete deletes a MomentsRow from [Moment-Db].[moment].[Moments].
func (m *MomentsRow) delete(db sq.BaseRunner) (cnt int64) {
	if c == nil {
		c = dba.OpenTx()
		defer func() { c.Close(err) }()
	}

	delete := sq.Delete("[moment].[Moments]").Where("[ID]", m.momentID)

	res, err := delete.RunWith(db).Exec()
	if err != nil {
		Error.Println(err)
		m.err = err
		return
	}
	cnt, err = res.RowsAffected()
	if err != nil {
		Error.Println(err)
		m.err = err
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
	momentSchema = "[moment]"

	momentAlias = "m"
	mediaAlias  = "med"
	sharesAlias = "s"
	findsAlias  = "f"

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
)

var (
	mLostColumns = []string{
		momentsAlias + ".[ID]",
		momentsAlias + ".[Latitude]",
		momentsAlias + ".[Longitude]",
	}
	mTypeColumns = []string{
		momentsAlias + ".[Public]",
		momentsAlias + ".[Hidden]",
	}
	mUserID = []string{
		momentsAlias + ".[UserID]",
	}
	mCreateDate = []string{
		momentsAlias + ".[CreateDate]",
	}

	moments   = momentSchema + ".[Moments] " + momentAlias
	onMoments = momentAlias + ".[ID]"
)

func mLocationBetween(l *Location) map[string]interface{} {
	return map[string]interface{}{
		momentAlias + ".Latitude BETWEEN ? AND ?":  []interface{}{l.latitude - 1, l.latitude + 1},
		momentAlias + ".Longitude BETWEEN ? AND ?": []interface{}{l.longitude - 1, l.longitude + 1},
	}
}

func mPublicEquals(b bool) map[string]interface{} {
	return map[string]interface{}{
		momentAlias + ".[Public]": b,
	}
}

func mHiddenEquals(b bool) map[string]interface{} {
	return map[string]interface{}{
		momentAlias + ".[Hidden]": b,
	}
}

func mUserIDEquals(u string) map[string]interface{} {
	return map[string]interface{}{
		momentAlias + ".[UserID]": u,
	}
}

var (
	mdColumns = []string{
		mediaAlias + ".[Message]",
		mediaAlias + ".[Type]",
		mediaAlias + ".[Dir]",
	}

	fUserID = []string{
		findsAlias + ".[UserID]",
	}

	fFindDate = []string{
		findsAlias + ".[FindDate]",
	}
)

func finds(on string) string {
	finds = momentSchema + ".[Finds] " + findsAlias
	if on != "" {
		finds += " ON " + findsAlias + ".[MomentID] = " + on
	}
	return finds
}

func media(on string) string {
	media := momentSchema + ".[Media] " + mediaAlias
	if on != "" {
		media += " ON " + mediaAlias + ".[MomentID] = " + on
	}
	return media
}

func fFoundEquals(b bool) map[string]interface{} {
	return map[string]interface{}{
		findsAlias + ".[Found]": b,
	}
}

func fUserIDEquals(u string) map[string]interface{} {
	return map[string]interface{}{
		findsAlias + ".[UserID]": u,
	}
}

func shares(on string) string {
	shares := momentSchema + ".[Shares] " + SharesAlias
	if on != "" {
		shares += " ON " + sharesAlias + ".[MomentID] = " + on
	}
	return shares
}

func sRecipientEquals(me string) map[string]interface{} {
	return map[string]interface{}{
		"(" + sharesAlias + ".[RecipientID] = ? OR " + sharesAlias + ".[All] = 1)": me,
	}
}

func sUserIDEquals(you) map[string]interface{} {
	return map[string]interface{}{
		sharesAlias + ".[UserID]": you,
	}
}

func momentsQuery(mt momentsType) *sq.SelectBuilder {
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
		From(moments).
		Join(media(onMoments))

	return query
}

func lostQuery() *sq.SelectBuilder {
	query := sq.
		Select().
		Columns(mLostColumns...).
		From(moments)

	return query
}

type Client interface {
	LocationSelector
	UserSelector
}

type LocationSelector interface {
	LocationShared(sq.BaseRunner, *Location, string) ResultsMap
	LocationPublic(sq.BaseRunner, *Location) ResultsMap
	LocationHidden(sq.BaseRunner, *Location) ResultsMap
	LocationLost(sq.BaseRunner, *Location, string) ResultsMap
}

type MomentClient struct {
	err error
}

func (mc *MomentClient) LocationShared(db sq.BaseRunner, l *Location, me string) (r ResultsMap) {
	if l == nil || m == "" {
		Error.Println(ErrorVariableEmpty)
		mc.err = ErrorVariableEmpty
		return
	}

	query := momentsQuery(Shared)

	query = query.
		Join(shares(onMoments)).
		Where(mLocationBetween(l)).
		Where(sRecipientEquals(me))

	r, err = process(query, momentsResults)
	if err != nil {
		Error.Println(err)
		mc.err = err
	}
	return
}

func (mc *MomentClient) LocationPublic(db sq.BaseRunner, l *Location) (r ResultsMap) {
	if l == nil {
		Error.Println(ErrorVariableEmpty)
		mc.err = ErrorVariableEmpty
		return
	}

	query := momentsQuery(Public)

	query = query.Where(mLocationBetween(l)).
		Where(mPublicEquals(true)).
		Where(mHiddenEquals(false))

	r, err = process(query, publicResults)
	if err != nil {
		Error.Println(err)
		mc.err = err
	}
	return
}

func (mc *MomentClient) LocationHidden(db sq.BaseRunner, l *Location) (r ResultsMap) {
	if l == nil {
		Error.Println(ErrorVariableEmpty)
		mc.err = ErrorVariableEmpty
		return
	}

	query := lostQuery()

	query = query.Where(mLocationBetween(l)).
		Where(mPublicEquals(true)).
		Where(mHiddenEquals(true))

	r, err = process(query, lostResults)
	if err != nil {
		Error.Println(err)
		mc.err = err
	}
	return
}

func (mc *MomentClient) LocationLost(db sq.BaseRunner, l *Location, me string) (r ResultsMap) {
	if l == nil || me == "" {
		Error.Println(ErrorVariableEmpty)
		mc.err = ErrorVariableEmpty
		return
	}

	query := lostQuery()

	query = query.
		Join(finds(onMoments)).
		Where(mLocationBetween(l)).
		Where(mPublicEquals(false)).
		Where(fUserIDEquals(me))

	r, err = process(query, lostResults)
	if err != nil {
		Error.Println(err)
		mc.err = err
	}
	return
}

type UserSelector interface {
	UserShared(sq.BaseRunner, string, string) ResultsMap
	UserLeft(sq.BaseRunner, string) ResultsMap
	UserFound(sq.BaseRunner, string) ResultsMap
}

func (mc *MomentClient) UserShared(db sq.BaseRunner, me string, u string) (r ResultsMap) {
	if me == "" || u == "" {
		Error.Println(ErrorVariableEmpty)
		mc.err = ErrorVariableEmpty
		return
	}

	query := momentsQuery(Shared)
	query = query.
		Join(shares(onMoments)).
		Where(sUserIDEquals(u)).
		Where(sRecipientEquals(me))

	r, err = process(query, momentsResults)
	if err != nil {
		Error.Println(err)
	}
	return
}

func (mc *MomentClient) UserLeft(db sq.BaseRunner, me string) (r ResultsMap) {
	if me == "" {
		Error.Println(ErrorVariableEmpty)
		mc.err = ErrorVariableEmpty
		return
	}

	query := momentsQuery(Left)
	query = query.
		Columns(append(fUserID, fFindDate...)...).
		Join(finds(onMoments)).
		Where(mUserIDEquals(me))

	r, err = process(db, query, leftResults)
	if err != nil {
		Error.Println(err)
		mc.err = err
	}
	return
}

func (mc *MomentClient) UserFound(db sq.BaseRunner, me string) (r ResultsMap) {
	if me == "" {
		Error.Println(ErrorVariableEmpty)
		mc.err = ErrorVariableEmpty
		return
	}

	query := momentsQuery(Found)
	query = query.
		Columns(fFindDate...).
		From(finds(onMoments)).
		Where(fUserIDEquals(me)).
		Where(fFoundEquals(true))

	fdc := findsFindDateColumn(query)
	query.SetColumns(fdc)

	ff := findsFrom(query, momentsJoin(findsAlias))
	query.SetFroms(ff)

	r, err = process(db, query, foundResults)
	if err != nil {
		Error.Println(err)
		mc.err = err
	}
	return
}

var (
	ErrorQueryStringEmpty = errors.New("Empty string passed into queryString parameter.")
)

func process(db sq.BaseRunner, query *sq.SelectBuilder, rowHandler func(*sql.Rows) (ResultsMap, error)) (r ResultsMap, err error) {
	if util.IsEmpty(query) {
		return r, ErrorVariableEmpty
	}

	Info.Printf("\nQueryString:\n\n%v\n\n", query.ToSql())

	rows, err := query.RunWith(db).Query()
	if err != nil {
		Error.Println(err)
		return
	}
	defer rows.Close()

	r, err = rowHandler(rows)
	if err != nil {
		Error.Println(err)
		return
	}
	if err = rows.Err(); err != nil {
		Error.Println(err)
	}

	return
}

func momentsResults(rows *sql.Rows) (rm ResultsMap, err error) {

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

	rm = make(ResultsMap, 0)
	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			Error.Println(err)
			return
		}

		m.createDate, err = dba.ParseDateTime2(createDate)
		if err != nil {
			Error.Println(err)
			return
		}

		if rm.exists(m) {
			if err = rm.append(m, md); err != nil {
				Error.Println(err)
				return
			}
		} else {
			if err = rm.add(m, md); err != nil {
				Error.Println(err)
				return
			}
		}
	}
	return
}

func publicResults(rows *sql.Rows) (rm ResultsMap, err error) {

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

	rm = make(ResultsMap, 0)
	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			Error.Println(err)
			return
		}

		m.createDate, err = dba.ParseDateTime2(createDate)
		if err != nil {
			Error.Println(err)
			return
		}

		if rm.exists(m) {
			if err = rm.append(m, md); err != nil {
				Error.Println(err)
				return
			}
		} else {
			if err = rm.add(m, md); err != nil {
				Error.Println(err)
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

	rm = make(ResultsMap, 0)
	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			Error.Println(err)
			return
		}
		m.createDate, err = dba.ParseDateTime2(createDate)
		if err != nil {
			Error.Println(err)
			return
		}
		f.findDate, err = dba.ParseDateTime2(findDate)
		if err != nil {
			Error.Println(err)
			return
		}

		if rm.exists(m) {
			if err = rm.append(m, md, f); err != nil {
				Error.Println(err)
				return
			}
		} else {
			if err = rm.add(m, md, f); err != nil {
				Error.Println(err)
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

	rm = make(ResultsMap, 0)
	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			Error.Println(err)
			return
		}
		m.createDate, err = dba.ParseDateTime2(createDate)
		if err != nil {
			Error.Println(err)
			return
		}
		f.findDate, err = dba.ParseDateTime2(findDate)
		if err != nil {
			Error.Println(err)
			return
		}

		if rm.exists(m) {
			if err = rm.append(m, md, f); err != nil {
				Error.Println(err)
				return
			}
		} else {
			if err = rm.add(m, md, f); err != nil {
				Error.Println(err)
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
			Error.Println(err)
			return
		}
		newR := &ResultMap{
			moment: m,
		}
		rm[m.momentID] = newR
	}

	return
}
