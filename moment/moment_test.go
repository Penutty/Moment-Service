package moment

import (
	// "errors"
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	// "math/rand"
	"fmt"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	tUser      string = "user00"
	tUser2     string = "user02"
	tUser3     string = "user03"
	tEmptyUser string = ""

	lat  float32 = 0.00
	long float32 = 0.00

	base int = 1
)

func TestMain(m *testing.M) {
	call := m.Run()

	os.Exit(call)
}

func TestMomentDB(t *testing.T) {
	db := DB()
	assert.IsType(t, new(sql.DB), db)
}

func TestCheckTime(t *testing.T) {
	type test struct {
		td       *time.Time
		expected error
	}
	td := time.Now().UTC()
	tests := []test{
		test{nil, ErrorTimePtrNil},
		test{&td, nil},
	}

	for _, v := range tests {
		err := checkTime(v.td)
		assert.Exactly(t, v.expected, err)
	}
}

func TestCheckMomentID(t *testing.T) {
	type test struct {
		momentID int64
		expected error
	}
	tests := []test{
		test{1, nil},
		test{0, nil},
		test{-1, ErrorMomentID},
		test{100, nil},
		test{1000000, nil},
	}

	for _, v := range tests {
		err := checkMomentID(v.momentID)
		assert.Exactly(t, v.expected, err)
	}
}

func TestCheckMediaType(t *testing.T) {
	type test struct {
		mType    uint8
		expected error
	}
	tests := []test{
		test{0, nil},
		test{1, nil},
		test{2, nil},
		test{3, nil},
		test{4, ErrorMediaTypeDNE},
		test{200, ErrorMediaTypeDNE},
	}
	for _, v := range tests {
		err := checkMediaType(v.mType)
		assert.Exactly(t, v.expected, err)
	}

}

func TestCheckUserID(t *testing.T) {
	type test struct {
		userID   string
		expected error
	}
	tests := []test{
		test{tUser, nil},
		test{tEmptyUser, ErrorUserIDShort},
		test{"user", ErrorUserIDShort},
		test{strings.Repeat("c", maxUserChars), nil},
		test{strings.Repeat("c", maxUserChars+1), ErrorUserIDLong},
	}
	for _, v := range tests {
		err := checkUserID(v.userID)
		assert.Exactly(t, v.expected, err)
	}
}

func TestNewLocation(t *testing.T) {
	type test struct {
		lat      float32
		long     float32
		expected error
	}
	tests := []test{
		test{0, 0, nil},
		test{minLat, 0, nil},
		test{minLat - 1, 0, ErrorLatitude},
		test{maxLat, 0, nil},
		test{maxLat + 1, 0, ErrorLatitude},
		test{0, minLong, nil},
		test{0, minLong - 1, ErrorLongitude},
		test{0, maxLong, nil},
		test{0, maxLong + 1, ErrorLongitude},
	}

	for _, v := range tests {
		mc := new(MomentClient)
		_ = mc.NewLocation(v.lat, v.long)
		assert.Exactly(t, v.expected, mc.Err())
	}
}

func TestLocationString(t *testing.T) {
	mc := new(MomentClient)
	l := mc.NewLocation(lat, long)

	expected := fmt.Sprintf("latitude: %v, longitude: %v", l.latitude, l.longitude)
	actual := l.String()
	assert.Equal(t, expected, actual)
}

func TestNewMediaRow(t *testing.T) {
	type test struct {
		momentID int64
		message  string
		mType    uint8
		dir      string
		expected error
	}
	tests := []test{
		test{1, "message", DNE, "", nil},
		test{1, strings.Repeat("c", maxMessage), DNE, "", nil},
		test{1, strings.Repeat("c", maxMessage+1), DNE, "", ErrorMessageLong},
		test{1, "message", DNE, "D:/dir/", ErrorMediaDNE},
		test{1, "message", Image, "D:/dir/", nil},
		test{1, "message", Image, "", ErrorMediaExistsDirDNE},
	}

	for _, v := range tests {
		mc := new(MomentClient)
		_ = mc.NewMediaRow(v.momentID, v.message, v.mType, v.dir)
		assert.Exactly(t, v.expected, mc.Err())
	}
}

func TestMediaRowString(t *testing.T) {
	mc := new(MomentClient)
	md := mc.NewMediaRow(1, "msg", DNE, "")
	expected := fmt.Sprintf("momentID: %v, mType: %v, message: \"%v\", dir: \"%v\"", md.momentID, md.mType, md.message, md.dir)
	actual := md.String()
	assert.Equal(t, expected, actual)
}

func TestNewFindsRow(t *testing.T) {
	type test struct {
		momentID int64
		userID   string
		found    bool
		findDate *time.Time
		expected error
	}
	fd := time.Now().UTC()
	tests := []test{
		test{1, tUser, true, &fd, nil},
		test{1, tUser, true, &time.Time{}, ErrorFoundEmptyFindDate},
		test{1, tUser, false, &time.Time{}, nil},
		test{1, tUser, false, &fd, ErrorNotFoundFindDateExists},
	}

	for _, v := range tests {
		mc := new(MomentClient)
		_ = mc.NewFindsRow(v.momentID, v.userID, v.found, v.findDate)
		assert.Exactly(t, v.expected, mc.Err())
	}
}

func TestFindsRowString(t *testing.T) {
	mc := new(MomentClient)
	f := mc.NewFindsRow(1, tUser, false, &time.Time{})
	expected := fmt.Sprintf("momentID: %v, userID: %v, found: %v, findDate: %v", f.momentID, f.userID, f.found, f.findDate)
	actual := f.String()
	assert.Equal(t, expected, actual)
}

func TestNewSharesRow(t *testing.T) {
	type test struct {
		sharesID int64
		momentID int64
		userID   string
		expected error
	}
	tests := []test{
		test{1, 1, tUser, nil},
		test{1, 1, strings.Repeat(tUser, 100), ErrorUserIDLong},
	}

	for _, v := range tests {
		mc := new(MomentClient)
		_ = mc.NewSharesRow(v.sharesID, v.momentID, v.userID)
		assert.Exactly(t, v.expected, mc.Err())
	}
}

func TestSharesRowString(t *testing.T) {
	mc := new(MomentClient)
	s := mc.NewSharesRow(1, 1, tUser)
	expected := fmt.Sprintf("ID: %v, momentID: %v, userID: %v", s.sharesID, s.momentID, s.userID)
	actual := s.String()
	assert.Equal(t, expected, actual)
}

func TestNewMomentsRow(t *testing.T) {
	type test struct {
		location   *Location
		userID     string
		public     bool
		hidden     bool
		createDate *time.Time
		expected   error
	}

	mc := new(MomentClient)
	lo := mc.NewLocation(lat, long)
	cd := time.Now().UTC()
	tests := []test{
		test{lo, tUser, true, false, &cd, nil},
		test{nil, tUser, true, false, &cd, ErrorLocationIsNil},
		test{lo, tUser, true, false, &cd, nil},
		test{lo, tUser, false, false, &cd, nil},
		test{lo, tUser, false, true, &cd, ErrorPrivateHiddenMoment},
	}

	for _, v := range tests {
		mc := new(MomentClient)
		_ = mc.NewMomentsRow(v.location, v.userID, v.public, v.hidden, v.createDate)
		assert.Exactly(t, v.expected, mc.Err())
	}
}

func TestMomentsRowString(t *testing.T) {
	mc := new(MomentClient)
	dt := time.Now().UTC()
	m := mc.NewMomentsRow(mc.NewLocation(lat, long), tUser, false, false, &dt)
	expected := fmt.Sprintf("id: %v, userID: %v, Location: %v, public: %v, hidden: %v, creatDate: %v", m.momentID, m.userID, m.Location, m.public, m.hidden, m.createDate)
	actual := m.String()
	assert.Equal(t, expected, actual)
}

func TestNewRecipientsRow(t *testing.T) {
	type test struct {
		id        int64
		all       bool
		recipient string
		expected  error
	}

	tests := []test{
		test{0, false, tUser, nil},
		test{0, true, tEmptyUser, nil},
		test{0, false, tEmptyUser, ErrorNotAllRecipientDNE},
		test{0, true, tUser, ErrorAllRecipientExists},
	}

	for _, v := range tests {
		mc := new(MomentClient)
		_ = mc.NewRecipientsRow(v.id, v.all, v.recipient)
		assert.Exactly(t, v.expected, mc.Err())
	}
}

func TestMomentString(t *testing.T) {
	dt := time.Now().UTC()
	m := Moment{
		momentID:   1,
		userID:     tUser,
		public:     false,
		hidden:     false,
		Location:   Location{latitude: lat, longitude: long},
		createDate: &dt,
	}
	expected := fmt.Sprintf("\nmomentID: %v\nuserID: %v\npublic: %v\nhidden: %v\nLocation: %v\ncreateDate: %v\nmedia:\nfinds:\nshares:\n", m.momentID, m.userID, m.public, m.hidden, m.Location, m.createDate)
	actual := m.String()
	assert.Equal(t, expected, actual)
}

var (
	MomentsRowRegexpStr = fmt.Sprintf(`^INSERT INTO \%s\.\%s \(\%s,\%s,\%s,\%s,\%s,\%s\) VALUES \(\?,\?,\?,\?,\?,\?\)$`,
		momentSchema,
		moments,
		userID,
		latStr,
		longStr,
		public,
		hidden,
		createDate)

	FindsRowRegexpStr = fmt.Sprintf(`^INSERT INTO \%s\.\%s \(\%s,\%s,\%s,\%s\) VALUES (\(\?,\?,\?,\?\)(,|$))+`,
		momentSchema,
		finds,
		momentID,
		userID,
		found,
		findDate)

	SharesRowRegexpStr = fmt.Sprintf(`INSERT INTO \%s\.\%s \(\%s,\%s\) VALUES \(\?,\?\)$`,
		momentSchema,
		shares,
		momentID,
		userID)

	RecipientsRowRegexpStr = fmt.Sprintf(`INSERT INTO \%s\.\%s \(\%s,\%s,\%s\) VALUES \(\?,\?,\?\)$`,
		momentSchema,
		recipients,
		sharesID,
		all,
		recipientID)

	MediaRowRegexpStr = fmt.Sprintf(`^INSERT INTO \%s\.\%s \(\%s,\%s,\%s,\%s\) VALUES \(\?,\?,\?,\?\)$`,
		momentSchema,
		media,
		momentID,
		message,
		mtype,
		dir)
)

func TestFindPublic(t *testing.T) {
	t.Run("Parameter Checks", func(t *testing.T) {

		db, _, err := sqlmock.New()
		assert.Nil(t, err)

		mc := new(MomentClient)
		f := mc.NewFindsRow(1, tUser, false, &time.Time{})
		_, err = mc.FindPublic(db, f)
		assert.Equal(t, ErrorFieldInvalid, err)
	})

	t.Run("1", func(t *testing.T) {
		db, mock, err := sqlmock.New()

		mc := new(MomentClient)
		dt := time.Now().UTC()
		f := mc.NewFindsRow(1, tUser, true, &dt)

		mock.ExpectExec(FindsRowRegexpStr).
			WithArgs(f.momentID, f.userID, f.found, f.findDate).
			WillReturnResult(sqlmock.NewResult(f.momentID, 1))

		cnt, err := mc.FindPublic(db, f)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), cnt)

		assert.Nil(t, mock.ExpectationsWereMet())
	})

}

func TestFindPrivate(t *testing.T) {
	t.Run("Parameter Checks", func(t *testing.T) {
		db, _, err := sqlmock.New()
		assert.Nil(t, err)

		mc := new(MomentClient)
		f := mc.NewFindsRow(1, tUser, false, &time.Time{})
		err = mc.FindPrivate(db, f)
		assert.Equal(t, ErrorFieldInvalid, err)
	})

	t.Run("1", func(t *testing.T) {
		db, mock, err := sqlmock.New()

		dt := time.Now().UTC()
		mc := new(MomentClient)
		f := mc.NewFindsRow(1, tUser, true, &dt)

		s := fmt.Sprintf(`^UPDATE \%s\.\%s SET \%s = \?, \%s = \? WHERE \%s = \? AND \%s = \?$`,
			momentSchema,
			finds,
			found,
			findDate,
			momentID,
			userID)
		mock.ExpectExec(s).
			WithArgs(f.found, f.findDate, f.momentID, f.userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = mc.FindPrivate(db, f)
		assert.Nil(t, err)

		assert.Nil(t, mock.ExpectationsWereMet())
	})
}

func TestCreatePrivate(t *testing.T) {
	t.Run("Parameter Checks", func(t *testing.T) {
		db, _, err := sqlmock.New()
		assert.Nil(t, err)

		mc := new(MomentClient)
		err = mc.CreatePrivate(db, nil, nil, nil)
		assert.Equal(t, ErrorParameterEmpty, err)
	})

	t.Run("1", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.Nil(t, err)

		mock.ExpectBegin()

		dt := time.Now().UTC()
		mock.ExpectExec(MomentsRowRegexpStr).
			WithArgs(tUser, lat, long, false, false, &dt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(MediaRowRegexpStr).
			WithArgs(1, "Helloworld.", DNE, "").
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(FindsRowRegexpStr).
			WithArgs(1, tUser2, false, &time.Time{}, 1, tUser3, false, &time.Time{}).
			WillReturnResult(sqlmock.NewResult(0, 2))

		mock.ExpectCommit()

		mc := new(MomentClient)
		m := mc.NewMomentsRow(mc.NewLocation(lat, long), tUser, false, false, &dt)
		md := mc.NewMediaRow(0, "Helloworld.", DNE, "")
		f1 := mc.NewFindsRow(0, tUser2, false, &time.Time{})
		f2 := mc.NewFindsRow(0, tUser3, false, &time.Time{})
		assert.Nil(t, mc.Err())

		err = mc.CreatePrivate(db, m, []*MediaRow{md}, []*FindsRow{f1, f2})
		assert.Nil(t, err)

		assert.Nil(t, mock.ExpectationsWereMet())
	})
}

func TestCreatePublic(t *testing.T) {
	t.Run("Parameter Checks", func(t *testing.T) {
		db, _, err := sqlmock.New()
		mc := new(MomentClient)
		err = mc.CreatePublic(db, nil, nil)
		assert.Equal(t, ErrorParameterEmpty, err)
	})

	t.Run("1", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		dt := time.Now().UTC()

		mock.ExpectBegin()

		mock.ExpectExec(MomentsRowRegexpStr).
			WithArgs(tUser, lat, long, false, false, &dt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(MediaRowRegexpStr).
			WithArgs(1, "Helloworld.", DNE, "").
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		mc := new(MomentClient)
		m := mc.NewMomentsRow(mc.NewLocation(lat, long), tUser, false, false, &dt)
		md := mc.NewMediaRow(0, "Helloworld.", DNE, "")
		assert.Nil(t, mc.Err())

		err = mc.CreatePublic(db, m, []*MediaRow{md})
		assert.Nil(t, err)

		assert.Nil(t, mock.ExpectationsWereMet())
	})
}

func TestShare(t *testing.T) {
	t.Run("Parameter Checks", func(t *testing.T) {
		db, _, err := sqlmock.New()
		mc := new(MomentClient)
		err = mc.Share(db, nil, nil)
		assert.Equal(t, ErrorParameterEmpty, err)
	})

	t.Run("1", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		mc := new(MomentClient)

		mock.ExpectBegin()

		mock.ExpectExec(SharesRowRegexpStr).
			WithArgs(1, tUser).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(RecipientsRowRegexpStr).
			WithArgs(1, false, tUser2).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		s := mc.NewSharesRow(0, 1, tUser)
		rs := []*RecipientsRow{mc.NewRecipientsRow(0, false, tUser2)}
		t.Log(s)
		t.Log(rs[0])

		err = mc.Share(db, s, rs)
		assert.Nil(t, err)

		assert.Nil(t, mock.ExpectationsWereMet())
	})
}

func Test_insert(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.Nil(t, err)

	invalidParameter := 1
	_, err = insert(db, invalidParameter)
	assert.Equal(t, ErrorTypeNotImplemented, err)
}

func Test_update(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.Nil(t, err)

	invalidParameter := 1
	err = update(db, invalidParameter)
	assert.Equal(t, ErrorTypeNotImplemented, err)
}

func TestLocationShared(t *testing.T) {
	t.Run("Parameter Checks", func(t *testing.T) {
		db, _, err := sqlmock.New()
		mc := new(MomentClient)
		_, err = mc.LocationShared(db, nil, "")
		assert.Equal(t, ErrorParameterEmpty, err)
	})

	t.Run("1", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		mc := new(MomentClient)

		s := fmt.Sprintf(`
		^SELECT 
		` + momentsAlias + `\.\` + iD + `, 
		` + momentsAlias + `\.\` + latStr + `, 
		` + momentsAlias + `\.\` + longStr + `, 
		` + mediaAlias + `\.\` + message + `, 
		` + mediaAlias + `\.\` + mtype + `, 
		` + mediaAlias + `\.\` + dir + `, 
		` + momentsAlias + `\.\` + createDate + `, 
		` + momentsAlias + `\.\` + userID + `, 
		` + momentsAlias + `\.\` + public + `, 
		` + momentsAlias + `\.\` + hidden + ` 
		FROM \` + momentSchema + `\.\` + moments + ` ` + momentsAlias + `  
		JOIN \` + momentSchema + `\.\` + media + ` ` + mediaAlias + `
		  ON ` + mediaAlias + `\.\` + momentID + ` = ` + momentsAlias + `\.\` + iD + `
		JOIN \` + momentSchema + `\.\` + shares + ` ` + sharesAlias + `
		  ON ` + sharesAlias + `\.\` + momentID + ` = ` + momentsAlias + `\.\` + iD + `
		JOIN \` + momentSchema + `\.\` + recipients + ` ` + recipientsAlias + `
		  ON ` + recipientsAlias + `\.\` + sharesID + ` = ` + sharesAlias + `\.\` + iD + `
		WHERE ` + momentsAlias + `\.\` + latStr + ` BETWEEN \? AND \?
			  AND ` + momentsAlias + `\.\` + longStr + ` BETWEEN \? AND \?
			  AND \(` + recipientsAlias + `\.\` + recipientID + ` = \? OR ` + recipientsAlias + `\.\` + all + ` = 1\)$`)

		rows := sqlmock.NewRows([]string{"NoColumns"})

		mock.ExpectQuery(s).WithArgs(lat-1, lat+1, long-1, long+1, tUser).WillReturnRows(rows)

		_, err = mc.LocationShared(db, mc.NewLocation(lat, long), tUser)
		assert.Nil(t, err)

		assert.Nil(t, mock.ExpectationsWereMet())
	})
}

func TestLocationPublic(t *testing.T) {
	t.Run("Parameter Checks", func(t *testing.T) {
		db, _, err := sqlmock.New()
		mc := new(MomentClient)
		_, err = mc.LocationPublic(db, nil)
		assert.Equal(t, ErrorParameterEmpty, err)
	})

	t.Run("1", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		mc := new(MomentClient)

		s := fmt.Sprintf(`
		^SELECT 
		` + momentsAlias + `\.\` + iD + `, 
		` + momentsAlias + `\.\` + latStr + `, 
		` + momentsAlias + `\.\` + longStr + `, 
		` + mediaAlias + `\.\` + message + `, 
		` + mediaAlias + `\.\` + mtype + `, 
		` + mediaAlias + `\.\` + dir + `, 
		` + momentsAlias + `\.\` + createDate + `, 
		` + momentsAlias + `\.\` + userID + ` 
		FROM \` + momentSchema + `\.\` + moments + ` ` + momentsAlias + `  
		JOIN \` + momentSchema + `\.\` + media + ` ` + mediaAlias + `
		  ON ` + mediaAlias + `\.\` + momentID + ` = ` + momentsAlias + `\.\` + iD + `
		WHERE ` + momentsAlias + `\.\` + latStr + ` BETWEEN \? AND \?
			  AND ` + momentsAlias + `\.\` + longStr + ` BETWEEN \? AND \?
			  AND ` + momentsAlias + `\.\` + public + ` = true 
			  AND ` + momentsAlias + `\.\` + hidden + ` = false$`)

		rows := sqlmock.NewRows([]string{"NoColumns"})

		mock.ExpectQuery(s).WithArgs(lat-1, lat+1, long-1, long+1).WillReturnRows(rows)

		_, err = mc.LocationPublic(db, mc.NewLocation(lat, long))
		assert.Nil(t, err)

		assert.Nil(t, mock.ExpectationsWereMet())
	})
}

func TestLocationHidden(t *testing.T) {
	t.Run("Parameter Checks", func(t *testing.T) {
		db, _, err := sqlmock.New()
		mc := new(MomentClient)
		_, err = mc.LocationHidden(db, nil)
		assert.Equal(t, ErrorParameterEmpty, err)
	})

	t.Run("1", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		mc := new(MomentClient)

		s := fmt.Sprintf(`
		^SELECT 
		` + momentsAlias + `\.\` + iD + `, 
		` + momentsAlias + `\.\` + latStr + `, 
		` + momentsAlias + `\.\` + longStr + ` 
		FROM \` + momentSchema + `\.\` + moments + ` ` + momentsAlias + `  
		WHERE ` + momentsAlias + `\.\` + latStr + ` BETWEEN \? AND \?
			  AND ` + momentsAlias + `\.\` + longStr + ` BETWEEN \? AND \?
			  AND ` + momentsAlias + `\.\` + public + ` = true 
			  AND ` + momentsAlias + `\.\` + hidden + ` = true$`)

		rows := sqlmock.NewRows([]string{"NoColumns"})
		mock.ExpectQuery(s).WithArgs(lat-1, lat+1, long-1, long+1).WillReturnRows(rows)

		_, err = mc.LocationHidden(db, mc.NewLocation(lat, long))
		assert.Nil(t, err)

		assert.Nil(t, mock.ExpectationsWereMet())
	})
}

func TestLocationLost(t *testing.T) {
	t.Run("Parameter Checks", func(t *testing.T) {
		db, _, err := sqlmock.New()
		mc := new(MomentClient)
		_, err = mc.LocationLost(db, nil, "")
		assert.Equal(t, ErrorParameterEmpty, err)
	})

	t.Run("1", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		mc := new(MomentClient)

		s := fmt.Sprintf(`
		^SELECT 
		` + momentsAlias + `\.\` + iD + `, 
		` + momentsAlias + `\.\` + latStr + `, 
		` + momentsAlias + `\.\` + longStr + ` 
		FROM \` + momentSchema + `\.\` + moments + ` ` + momentsAlias + `  
		JOIN \` + momentSchema + `\.\` + finds + ` ` + findsAlias + `
		  ON ` + findsAlias + `\.\` + momentID + ` = ` + momentsAlias + `\.\` + iD + `
		WHERE ` + momentsAlias + `\.\` + latStr + ` BETWEEN \? AND \?
			  AND ` + momentsAlias + `\.\` + longStr + ` BETWEEN \? AND \?
			  AND ` + momentsAlias + `\.\` + public + ` = false 
			  AND ` + momentsAlias + `\.\` + hidden + ` = false 
			  AND ` + findsAlias + `\.\` + userID + ` = \?$`)

		rows := sqlmock.NewRows([]string{"NoColumns"})
		mock.ExpectQuery(s).WithArgs(lat-1, lat+1, long-1, long+1, tUser).WillReturnRows(rows)

		_, err = mc.LocationLost(db, mc.NewLocation(lat, long), tUser)
		assert.Nil(t, err)

		assert.Nil(t, mock.ExpectationsWereMet())
	})
}

func TestUserShared(t *testing.T) {
	t.Run("Parameter Checks", func(t *testing.T) {
		db, _, err := sqlmock.New()
		mc := new(MomentClient)
		_, err = mc.UserShared(db, "", "")
		assert.Equal(t, ErrorParameterEmpty, err)
	})

	t.Run("1", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		mc := new(MomentClient)

		s := fmt.Sprintf(`
		^SELECT 
		` + momentsAlias + `\.\` + iD + `, 
		` + momentsAlias + `\.\` + latStr + `, 
		` + momentsAlias + `\.\` + longStr + `, 
		` + mediaAlias + `\.\` + message + `, 
		` + mediaAlias + `\.\` + mtype + `, 
		` + mediaAlias + `\.\` + dir + `, 
		` + momentsAlias + `\.\` + createDate + `, 
		` + momentsAlias + `\.\` + userID + `,
		` + momentsAlias + `\.\` + public + `,
		` + momentsAlias + `\.\` + hidden + `
		FROM \` + momentSchema + `\.\` + moments + ` ` + momentsAlias + `  
		JOIN \` + momentSchema + `\.\` + media + ` ` + mediaAlias + `
		  ON ` + mediaAlias + `\.\` + momentID + ` = ` + momentsAlias + `\.\` + iD + `
		JOIN \` + momentSchema + `\.\` + shares + ` ` + sharesAlias + `
		  ON ` + sharesAlias + `\.\` + momentID + ` = ` + momentsAlias + `\.\` + iD + `
		JOIN \` + momentSchema + `\.\` + recipients + ` ` + recipientsAlias + `
		  ON ` + recipientsAlias + `\.\` + sharesID + ` = ` + sharesAlias + `\.\` + iD + `
		WHERE ` + sharesAlias + `\.\` + userID + ` = \?
			  AND \(` + recipientsAlias + `\.\` + recipientID + ` = \? OR ` + recipientsAlias + `\.\` + all + ` = true\)$`)

		rows := sqlmock.NewRows([]string{"NoColumns"})
		mock.ExpectQuery(s).WithArgs(tUser, tUser2).WillReturnRows(rows)

		_, err = mc.UserShared(db, tUser, tUser2)
		assert.Nil(t, err)

		assert.Nil(t, mock.ExpectationsWereMet())
	})
}

func TestUserLeft(t *testing.T) {
	t.Run("Parameter Checks", func(t *testing.T) {
		db, _, err := sqlmock.New()
		mc := new(MomentClient)
		_, err = mc.UserLeft(db, "")
		assert.Equal(t, ErrorParameterEmpty, err)
	})

	t.Run("1", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		mc := new(MomentClient)

		s := fmt.Sprintf(`
		^SELECT 
		` + momentsAlias + `\.\` + iD + `, 
		` + momentsAlias + `\.\` + latStr + `, 
		` + momentsAlias + `\.\` + longStr + `, 
		` + mediaAlias + `\.\` + message + `, 
		` + mediaAlias + `\.\` + mtype + `, 
		` + mediaAlias + `\.\` + dir + `, 
		` + momentsAlias + `\.\` + createDate + `, 
		` + momentsAlias + `\.\` + public + `,
		` + momentsAlias + `\.\` + hidden + `,
		` + findsAlias + `\.\` + userID + `,
		` + findsAlias + `\.\` + findDate + `
		FROM \` + momentSchema + `\.\` + moments + ` ` + momentsAlias + `  
		JOIN \` + momentSchema + `\.\` + media + ` ` + mediaAlias + `
		  ON ` + mediaAlias + `\.\` + momentID + ` = ` + momentsAlias + `\.\` + iD + `
		JOIN \` + momentSchema + `\.\` + finds + ` ` + findsAlias + `
		  ON ` + findsAlias + `\.\` + momentID + ` = ` + momentsAlias + `\.\` + iD + `
		WHERE ` + momentsAlias + `\.\` + userID + ` = \?$`)

		rows := sqlmock.NewRows([]string{"NoColumns"})
		mock.ExpectQuery(s).WithArgs(tUser).WillReturnRows(rows)

		_, err = mc.UserLeft(db, tUser)
		assert.Nil(t, err)

		assert.Nil(t, mock.ExpectationsWereMet())
	})

}

func TestUserFound(t *testing.T) {
	t.Run("Parameter Checks", func(t *testing.T) {
		db, _, err := sqlmock.New()
		mc := new(MomentClient)
		_, err = mc.UserFound(db, "")
		assert.Equal(t, ErrorParameterEmpty, err)
	})

	t.Run("1", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		mc := new(MomentClient)

		s := fmt.Sprintf(`
		^SELECT 
		` + momentsAlias + `\.\` + iD + `, 
		` + momentsAlias + `\.\` + latStr + `, 
		` + momentsAlias + `\.\` + longStr + `, 
		` + mediaAlias + `\.\` + message + `, 
		` + mediaAlias + `\.\` + mtype + `, 
		` + mediaAlias + `\.\` + dir + `, 
		` + momentsAlias + `\.\` + createDate + `, 
		` + momentsAlias + `\.\` + userID + `,
		` + momentsAlias + `\.\` + public + `,
		` + momentsAlias + `\.\` + hidden + `,
		` + findsAlias + `\.\` + findDate + `
		FROM \` + momentSchema + `\.\` + moments + ` ` + momentsAlias + `  
		JOIN \` + momentSchema + `\.\` + media + ` ` + mediaAlias + `
		  ON ` + mediaAlias + `\.\` + momentID + ` = ` + momentsAlias + `\.\` + iD + `
		JOIN \` + momentSchema + `\.\` + finds + ` ` + findsAlias + `
		  ON ` + findsAlias + `\.\` + momentID + ` = ` + momentsAlias + `\.\` + iD + `
		WHERE ` + findsAlias + `\.\` + userID + ` = \?
			  AND ` + findsAlias + `\.\` + found + ` = true$`)

		rows := sqlmock.NewRows([]string{"NoColumns"})
		mock.ExpectQuery(s).WithArgs(tUser).WillReturnRows(rows)

		_, err = mc.UserFound(db, tUser)
		assert.Nil(t, err)

		assert.Nil(t, mock.ExpectationsWereMet())
	})

}

var (
	fakeSelect       = sq.Select("fakeColumn").From("fakeTbl")
	fakeSelectRegexp = `^SELECT fakeColumn FROM fakeTbl$`
)

func Test_selectMoments(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	query := fakeSelect

	dt := time.Now().UTC()
	rows := sqlmock.NewRows([]string{iD, latStr, longStr, message, mtype, dir, createDate, userID, public, hidden}).
		AddRow(1, lat, long, "Hello there.", DNE, "", &dt, tUser, false, false).
		AddRow(1, lat, long, "Enjoy this photo.", Image, "D:/ImageDir/image.png", &dt, tUser, false, false).
		AddRow(2, lat, long, "Where am I? :p", DNE, "", &dt, tUser, true, true)

	mock.ExpectQuery(fakeSelectRegexp).WillReturnRows(rows)

	mc := new(MomentClient)
	rs, err := mc.selectMoments(db, query)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(rs))

	assert.Nil(t, mock.ExpectationsWereMet())
}

func Test_selectPublicMoments(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	dt := time.Now().UTC()
	rows := sqlmock.NewRows([]string{iD, latStr, longStr, message, mtype, dir, createDate, userID}).
		AddRow(1, lat, long, "message 1", DNE, "", &dt, tUser).
		AddRow(1, lat, long, "message 2", DNE, "", &dt, tUser).
		AddRow(2, lat, long, "message 3", DNE, "", &dt, tUser)

	mock.ExpectQuery(fakeSelectRegexp).WillReturnRows(rows)

	mc := new(MomentClient)
	rs, err := mc.selectPublicMoments(db, fakeSelect)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(rs))
	assert.Nil(t, mock.ExpectationsWereMet())
}

func Test_selectLostMoments(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	rows := sqlmock.NewRows([]string{iD, latStr, longStr}).
		AddRow(1, lat, long).
		AddRow(2, lat, long).
		AddRow(3, lat, long)

	mock.ExpectQuery(fakeSelectRegexp).WillReturnRows(rows)

	mc := new(MomentClient)
	rs, err := mc.selectLostMoments(db, fakeSelect)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(rs))
	assert.Nil(t, mock.ExpectationsWereMet())
}

func Test_selectLeftMoments(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	dt := time.Now().UTC()
	rows := sqlmock.NewRows([]string{iD, latStr, longStr, message, mtype, dir, createDate, public, hidden, userID, findDate}).
		AddRow(1, lat, long, "message 1", DNE, "", &dt, false, false, tUser2, &dt).
		AddRow(2, lat, long, "message 2", DNE, "", &dt, false, false, tUser2, &dt).
		AddRow(2, lat, long, "message 3", Image, "D:/Image/image.png", &dt, false, false, tUser2, &dt).
		AddRow(3, lat, long, "message 4", DNE, "", &dt, false, false, tUser2, &dt).
		AddRow(3, lat, long, "message 4", DNE, "", &dt, false, false, tUser3, &dt)

	mock.ExpectQuery(fakeSelectRegexp).WillReturnRows(rows)

	mc := new(MomentClient)
	rs, err := mc.selectLeftMoments(db, fakeSelect)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(rs))

	assert.Nil(t, mock.ExpectationsWereMet())
}

func Test_selectFoundMoments(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	dt := time.Now().UTC()
	rows := sqlmock.NewRows([]string{iD, latStr, longStr, message, mtype, dir, createDate, userID, public, hidden, findDate}).
		AddRow(1, lat, long, "message 1", DNE, "", &dt, tUser, false, false, &dt).
		AddRow(2, lat, long, "message 2", DNE, "", &dt, tUser, false, false, &dt).
		AddRow(2, lat, long, "message 3", Image, "D:/Image/image.png", &dt, tUser, false, false, &dt).
		AddRow(3, lat, long, "message 4", DNE, "", &dt, tUser, true, true, &dt).
		AddRow(3, lat, long, "message 5", DNE, "", &dt, tUser, true, true, &dt)

	mock.ExpectQuery(fakeSelectRegexp).WillReturnRows(rows)

	mc := new(MomentClient)
	rs, err := mc.selectFoundMoments(db, fakeSelect)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(rs))

	assert.Nil(t, mock.ExpectationsWereMet())
}
