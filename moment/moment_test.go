package moment

import (
	// "errors"
	"database/sql"
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
	db := MomentDB()
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

func TestNewSharesRow(t *testing.T) {
	type test struct {
		momentID    int64
		userID      string
		all         bool
		recipientID string
		expected    error
	}
	tests := []test{
		test{1, tUser, false, "user_02", nil},
		test{1, tUser, true, "user_02", ErrorAllRecipientExists},
		test{1, tUser, true, "", nil},
		test{1, tUser, false, "", ErrorNotAllRecipientDNE},
	}

	for _, v := range tests {
		mc := new(MomentClient)
		_ = mc.NewSharesRow(v.momentID, v.userID, v.all, v.recipientID)
		assert.Exactly(t, v.expected, mc.Err())
	}
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

	SharesRowRegexpStr = fmt.Sprintf(`INSERT INTO \%s\.\%s \(\%s,\%s,\%s,\%s\) VALUES \(\?,\?,\?,\?\)$`,
		momentSchema,
		shares,
		momentID,
		userID,
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
		_, err = mc.Share(db, nil)
		assert.Equal(t, ErrorParameterEmpty, err)
	})

	t.Run("1", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		mc := new(MomentClient)

		mock.ExpectExec(SharesRowRegexpStr).
			WithArgs(1, tUser, true, tEmptyUser).
			WillReturnResult(sqlmock.NewResult(0, 1))

		s := mc.NewSharesRow(1, tUser, true, tEmptyUser)
		cnt, err := mc.Share(db, []*SharesRow{s})
		assert.Nil(t, err)
		assert.Equal(t, int64(1), cnt)

		assert.Nil(t, mock.ExpectationsWereMet())
	})
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
		WHERE ` + momentsAlias + `\.\` + latStr + ` BETWEEN \? AND \?
			  AND ` + momentsAlias + `\.\` + longStr + ` BETWEEN \? AND \?
			  AND \(` + sharesAlias + `\.\` + recipientID + ` = \? OR ` + sharesAlias + `\.\` + all + ` = 1\)$`)

		dt := time.Now().UTC()
		rows := sqlmock.
			NewRows([]string{
				iD,
				latStr,
				longStr,
				message,
				mtype,
				dir,
				createDate,
				userID,
				public,
				hidden}).
			AddRow(1, 0.00, 0.00, "Helloworld.", DNE, "", &dt, tUser, false, false)

		mock.ExpectQuery(s).WithArgs(lat-1, lat+1, long-1, long+1, tUser).WillReturnRows(rows)

		_, err := mc.LocationShared(db, mc.NewLocation(lat, long), tUser)
		assert.Nil(t, err)

		assert.Nil(t, mock.ExpectationsWereMet())
	})
}

func TestselectMoments(t *testing.T) {
	dt := time.Now().UTC()
	rows := sqlmock.
		NewRows([]string{
			iD,
			latStr,
			longStr,
			message,
			mtype,
			dir,
			createDate,
			userID,
			public,
			hidden}).
		AddRow(1, 0.00, 0.00, "Helloworld.", DNE, "", &dt, tUser, false, false).
		AddRow(2, 0.00, 0.00, "Hi how are you?", DNE, "", &dt, tUser, false, false)

}

//const (
//	single int = 1
//	start  int = 1
//	length int = 500
//)

//func TestFindsInsertDelete(t *testing.T) {
//	fS, err := newFinds()
//	assert.Nil(t, err)
//
//	subTestFindsInsert(t, fS)
//	subTestFindsDelete(t, fS)
//}
//
//func TestFindsRowDelete(t *testing.T) {
//	td := time.Now().UTC()
//	f, err := NewFind(1, TestUser, true, &td)
//	assert.Nil(t, err)
//
//	fS := make(Finds, single)
//	fS[0] = f
//
//	subTestFindsInsert(t, fS)
//	subTestFindsDelete(t, fS)
//}
//
//func newFinds() (fS Finds, err error) {
//	fS = make(Finds, length)
//
//	for i := 0; i < length; i++ {
//		n := start + i
//		td := time.Now().UTC()
//		fS[i], err = NewFind(int64(n), TestUser+strconv.Itoa(n), true, &td)
//	}
//
//	return
//}
//
//func subTestFindsInsert(t *testing.T, fS Finds) {
//	t.Run("insert", func(t *testing.T) {
//		cnt, err := fS.insert(nil)
//		assert.Nil(t, err)
//		assert.Equal(t, int64(len(fS)), cnt)
//	})
//}
//
//func subTestFindsDelete(t *testing.T, fS Finds) {
//	t.Run("delete", func(t *testing.T) {
//		cnt, err := fS.delete()
//		assert.Nil(t, err)
//		assert.Equal(t, int64(len(fS)), cnt)
//	})
//}
//
//func TestSharesInsertDelete(t *testing.T) {
//	sS, err := newShares(base, length, false)
//	assert.Nil(t, err)
//
//	subTestSharesRowInsert(t, sS)
//	subTestSharesRowDelete(t, sS)
//}
//
//func TestSharesRowDelete(t *testing.T) {
//	sS, err := newShares(base, single, true)
//	assert.Nil(t, err)
//
//	subTestSharesRowInsert(t, sS)
//	subTestSharesRowDelete(t, sS)
//}
//
//func newShares(start int, cnt int, all bool) (sS Shares, err error) {
//	sS = make(Shares, cnt)
//
//	for i := 0; i < cnt; i++ {
//		n := start + i
//		if all {
//			sS[i], err = NewShare(int64(n), TestUser+strconv.Itoa(n), true, TestEmptyUser)
//		} else {
//			sS[i], err = NewShare(int64(n), TestUser+strconv.Itoa(n), false, TestUser+strconv.Itoa(n+1))
//		}
//	}
//	return
//}
//
//func subTestSharesRowInsert(t *testing.T, sS Shares) {
//	t.Run("insert", func(t *testing.T) {
//		cnt, err := sS.insert()
//		assert.Nil(t, err)
//		assert.Equal(t, int64(len(sS)), cnt)
//	})
//}
//
//func subTestSharesRowDelete(t *testing.T, sS Shares) {
//	t.Run("delete", func(t *testing.T) {
//		cnt, err := sS.delete()
//		assert.Empty(t, err)
//		assert.Equal(t, int64(len(sS)), cnt)
//	})
//}
//
//func TestMediaInsertDelete(t *testing.T) {
//	mS, err := newMedia()
//	assert.Nil(t, err)
//
//	subTestMediaInsert(t, mS)
//	subTestMediaDelete(t, mS)
//}
//
//func TestMediaRowDelete(t *testing.T) {
//	m, err := NewMedia(1, "message", DNE, "")
//	assert.Nil(t, err)
//
//	mS := make(Media, single)
//	mS[0] = m
//
//	subTestMediaInsert(t, mS)
//	subTestMediaDelete(t, mS)
//}
//
//func newMedia() (mS Media, err error) {
//	mS = make(Media, length)
//
//	for i := 0; i < length; i++ {
//		n := start + i
//		mS[i], err = NewMedia(int64(n), "message_0"+strconv.Itoa(n), 0, "")
//	}
//
//	return
//}
//
//func subTestMediaInsert(t *testing.T, mS Media) {
//	t.Run("insert", func(t *testing.T) {
//		cnt, err := mS.insert(nil)
//		assert.Equal(t, int64(len(mS)), cnt)
//		assert.Nil(t, err)
//	})
//}
//
//func subTestMediaDelete(t *testing.T, mS Media) {
//	t.Run("delete", func(t *testing.T) {
//		cnt, err := mS.delete()
//		assert.Equal(t, int64(len(mS)), cnt)
//		assert.Nil(t, err)
//	})
//}
//
//func TestMomentsRowInsertDelete(t *testing.T) {
//
//	l, err := NewLocation(lat, long)
//	assert.Nil(t, err)
//	td := time.Now().UTC()
//	m, err := NewMoment(l, TestUser, true, false, &td)
//	assert.Nil(t, err)
//
//	t.Run("insert", func(t *testing.T) {
//		id, err := m.insert(nil)
//		assert.NotZero(t, id)
//		assert.Nil(t, err)
//	})
//
//	t.Run("delete", func(t *testing.T) {
//		cnt, err := m.delete(nil)
//		assert.Equal(t, int64(single), cnt)
//		assert.Nil(t, err)
//	})
//}
//
//func TestMomentsRowCreatePublic(t *testing.T) {
//	r := MomentsRowCreatePublic(t, false)
//	MomentsRowDeletePublic(t, r)
//}
//
//func MomentsRowCreatePublic(t *testing.T, hidden bool) *Result {
//	l, err := NewLocation(lat, long)
//	assert.Nil(t, err)
//
//	td := time.Now().UTC()
//	m, err := NewMoment(l, TestUser, true, hidden, &td)
//	assert.Nil(t, err)
//
//	mediaCnt := 1
//	media := make(Media, mediaCnt)
//	for i := 0; i < len(media); i++ {
//		med, err := NewMedia(1, "HelloWorld", DNE, "")
//		assert.Nil(t, err)
//		media[i] = med
//	}
//
//	err = m.CreatePublic(&media)
//	assert.Nil(t, err)
//
//	return &Result{
//		moment: m,
//		media:  media,
//	}
//}
//
//func MomentsRowDeletePublic(t *testing.T, r *Result) {
//	cnt, err := r.media.delete()
//	assert.Nil(t, err)
//	assert.Equal(t, int64(len(r.media)), cnt)
//
//	cnt, err = r.moment.delete(nil)
//	assert.Nil(t, err)
//	assert.Equal(t, int64(single), cnt)
//
//}
//
//func TestMomentsRowCreatePrivate(t *testing.T) {
//	r := MomentsRowCreatePrivate(t)
//	MomentsRowDeletePrivate(t, r)
//}
//
//func MomentsRowCreatePrivate(t *testing.T) *Result {
//	l, err := NewLocation(lat, long)
//	assert.Nil(t, err)
//
//	td := time.Now().UTC()
//	m, err := NewMoment(l, TestUser, false, false, &td)
//	assert.Nil(t, err)
//
//	mediaCnt := 1
//	media := make(Media, mediaCnt)
//	for i := 0; i < len(media); i++ {
//		med, err := NewMedia(1, "HelloWorld", DNE, "")
//		assert.Nil(t, err)
//		media[i] = med
//	}
//
//	findsCnt := 10
//	finds := make(Finds, findsCnt)
//	for i := 0; i < findsCnt; i++ {
//		find, err := NewFind(1, TestUser+strconv.Itoa(i), false, &time.Time{})
//		assert.Nil(t, err)
//		finds[i] = find
//	}
//
//	err = m.CreatePrivate(&media, &finds)
//	assert.Nil(t, err)
//
//	return &Result{
//		moment: m,
//		finds:  finds,
//		media:  media,
//	}
//
//}
//
//func MomentsRowDeletePrivate(t *testing.T, r *Result) {
//	cnt, err := r.media.delete()
//	assert.Nil(t, err)
//	assert.Equal(t, int64(len(r.media)), cnt)
//
//	cnt, err = r.finds.delete()
//	assert.Nil(t, err)
//	assert.Equal(t, int64(len(r.finds)), cnt)
//
//	cnt, err = r.moment.delete(nil)
//	assert.Nil(t, err)
//	assert.Equal(t, int64(single), cnt)
//
//}
//
//func TestFindsRowFindPublic(t *testing.T) {
//	r := MomentsRowCreatePublic(t, false)
//
//	fd := time.Now().UTC()
//	f, err := NewFind(r.moment.momentID, TestUser+"2", true, &fd)
//
//	cnt, err := f.FindPublic()
//	assert.Nil(t, err)
//	assert.Equal(t, int64(single), cnt)
//
//	cnt, err = f.delete(nil)
//	assert.Nil(t, err)
//	assert.Equal(t, int64(single), cnt)
//
//	MomentsRowDeletePublic(t, r)
//}
//
//func TestFindsRowFindPrivate(t *testing.T) {
//	r := MomentsRowCreatePrivate(t)
//
//	// use momentID and userID to find a private moment
//	err := r.finds[0].FindPrivate()
//	assert.Nil(t, err)
//
//	MomentsRowDeletePrivate(t, r)
//}
//
//func TestLocationShared(t *testing.T) {
//	r := MomentsRowCreatePrivate(t)
//
//	f := r.finds[0]
//	sS := make(Shares, single)
//	s, err := NewShare(f.momentID, f.userID, true, TestEmptyUser)
//	assert.Nil(t, err)
//	sS[0] = s
//
//	cnt, err := sS.Share()
//	assert.Nil(t, err)
//	assert.Equal(t, int64(single), cnt)
//
//	l, err := NewLocation(lat, long)
//	assert.Nil(t, err)
//
//	resM, err := LocationShared(l, TestUser+"0")
//	res := resM.mapToSlice()
//	t.Log(res)
//	assert.Nil(t, err)
//	assert.Equal(t, TestUser, res[0].moment.userID)
//	assert.Equal(t, lat, res[0].moment.latitude)
//	assert.Equal(t, long, res[0].moment.longitude)
//	assert.Equal(t, uint8(DNE), res[0].media[0].mType)
//	assert.Equal(t, "HelloWorld", res[0].media[0].message)
//	assert.Equal(t, "", res[0].media[0].dir)
//
//	subTestSharesRowDelete(t, sS)
//	MomentsRowDeletePrivate(t, r)
//}
//
//func TestLocationPublic(t *testing.T) {
//	hidden := false
//	r := MomentsRowCreatePublic(t, hidden)
//
//	l, err := NewLocation(lat, long)
//	assert.Nil(t, err)
//
//	resM, err := LocationPublic(l)
//	res := resM.mapToSlice()
//	assert.Nil(t, err)
//	assert.Equal(t, TestUser, res[0].moment.userID)
//	assert.Equal(t, lat, res[0].moment.latitude)
//	assert.Equal(t, long, res[0].moment.longitude)
//	assert.Equal(t, uint8(DNE), res[0].media[0].mType)
//	assert.Equal(t, "HelloWorld", res[0].media[0].message)
//	assert.Equal(t, "", res[0].media[0].dir)
//
//	MomentsRowDeletePublic(t, r)
//}
//
//func TestLocationHidden(t *testing.T) {
//	hidden := true
//	r := MomentsRowCreatePublic(t, hidden)
//
//	l, err := NewLocation(lat, long)
//	assert.Nil(t, err)
//
//	resM, err := LocationHidden(l)
//	res := resM.mapToSlice()
//	assert.Nil(t, err)
//	assert.Equal(t, lat, res[0].moment.latitude)
//	assert.Equal(t, long, res[0].moment.longitude)
//
//	MomentsRowDeletePublic(t, r)
//}
//
//func TestLocationLost(t *testing.T) {
//	r := MomentsRowCreatePrivate(t)
//
//	l, err := NewLocation(lat, long)
//	assert.Nil(t, err)
//
//	me := TestUser + "1"
//	resM, err := LocationLost(l, me)
//	res := resM.mapToSlice()
//	assert.Nil(t, err)
//	assert.Equal(t, lat, res[0].moment.latitude)
//	assert.Equal(t, long, res[0].moment.longitude)
//
//	MomentsRowDeletePrivate(t, r)
//}
//
//func TestUserShared(t *testing.T) {
//	r := MomentsRowCreatePrivate(t)
//
//	f := r.finds[0]
//	s, err := NewShare(f.momentID, f.userID, true, TestEmptyUser)
//	assert.Nil(t, err)
//
//	f2 := r.finds[1]
//	s2, err := NewShare(f2.momentID, f2.userID, false, f.userID)
//	assert.Nil(t, err)
//
//	sS := make(Shares, 2)
//	sS[0] = s
//	sS[1] = s2
//
//	cnt, err := sS.Share()
//	assert.Nil(t, err)
//	assert.Equal(t, int64(2), cnt)
//
//	t.Run("1", func(t *testing.T) {
//		resM, err := UserShared(TestUser, f.userID)
//		res := resM.mapToSlice()
//		assert.Nil(t, err)
//		assert.Equal(t, TestUser, res[0].moment.userID)
//		assert.Equal(t, lat, res[0].moment.latitude)
//		assert.Equal(t, long, res[0].moment.longitude)
//		assert.Equal(t, uint8(DNE), res[0].media[0].mType)
//		assert.Equal(t, "HelloWorld", res[0].media[0].message)
//		assert.Equal(t, "", res[0].media[0].dir)
//	})
//
//	t.Run("2", func(t *testing.T) {
//		resM, err := UserShared(f.userID, f2.userID)
//		res := resM.mapToSlice()
//		assert.Nil(t, err)
//		assert.Equal(t, TestUser, res[0].moment.userID)
//		assert.Equal(t, lat, res[0].moment.latitude)
//		assert.Equal(t, long, res[0].moment.longitude)
//		assert.Equal(t, uint8(DNE), res[0].media[0].mType)
//		assert.Equal(t, "HelloWorld", res[0].media[0].message)
//		assert.Equal(t, "", res[0].media[0].dir)
//	})
//
//	subTestSharesRowDelete(t, sS)
//	MomentsRowDeletePrivate(t, r)
//}
//
//func TestUserLeft(t *testing.T) {
//	r := MomentsRowCreatePrivate(t)
//
//	t.Run("1", func(t *testing.T) {
//		resM, err := UserLeft(TestUser)
//		assert.Nil(t, err)
//		res := resM.mapToSlice()
//		assert.Equal(t, single, len(res))
//		assert.Equal(t, 10, len(res[0].finds))
//		assert.Equal(t, single, len(res[0].media))
//	})
//
//	MomentsRowDeletePrivate(t, r)
//}
//
//func TestUserFound(t *testing.T) {
//	r := MomentsRowCreatePrivate(t)
//
//	err := r.finds[0].FindPrivate()
//	assert.Nil(t, err)
//
//	t.Run("1", func(t *testing.T) {
//		resM, err := UserFound(r.finds[0].userID)
//		assert.Nil(t, err)
//		res := resM.mapToSlice()
//		assert.Equal(t, single, len(res))
//	})
//
//	t.Run("2", func(t *testing.T) {
//		resM, err := UserFound(r.finds[1].userID)
//		assert.Nil(t, err)
//		res := resM.mapToSlice()
//		assert.Equal(t, 0, len(res))
//	})
//
//	MomentsRowDeletePrivate(t, r)
//}

//func BenchmarkMomentsRowNew(b *testing.B) {
//	l, _ := NewLocation(lat, long)
//	td := time.Now().UTC()
//	_, _ = NewMoment(l, "user_01", true, false, &td)
//}
//
//func BenchmarkMomentsRowInsert(b *testing.B) {
//	l, _ := NewLocation(lat, long)
//	td := time.Now().UTC()
//	m, _ := NewMoment(l, "user_01", true, false, &td)
//
//	b.ResetTimer()
//	m.insert(nil)
//	b.StopTimer()
//	m.delete(nil)
//}
//
//func BenchmarkFindsInsert(b *testing.B) {
//	fS, _ := newFinds()
//
//	b.ResetTimer()
//	fS.insert(nil)
//	b.StopTimer()
//	fS.delete()
//}
//
//func BenchmarkSharesInsert(b *testing.B) {
//	sS, _ := newShares(base, length, false)
//
//	b.ResetTimer()
//	sS.insert()
//	b.StopTimer()
//	sS.delete()
//}
//
//func BenchmarkMediaInsert(b *testing.B) {
//	mS, _ := newMedia()
//
//	b.ResetTimer()
//	mS.insert(nil)
//	b.StopTimer()
//	mS.delete()
//}
//
//func BenchmarkSharesDelete(b *testing.B) {
//	b.StopTimer()
//	sS, _ := newShares(base, length, false)
//
//	sS.insert()
//	b.StartTimer()
//
//	sS.delete()
//}

// func Test_findPublic(t *testing.T) {
// 	t.Run("1", func(t *testing.T) {
// 		hiddenMoment = new(Moment)
// 		for _, v := range moments {
// 			if v.Hidden && v.Public {
// 				hiddenMoment = v
// 			}
// 		}

// 		if err = hiddenMoment.findPublic(); err != nil {
// 			t.Error(err)
// 		}
// 	})
// }

// func Test_searchLost(t *testing.T) {
// 	t.Run("1", func(t *testing.T) {
// 		var u string
// 		for _, m := range moments {
// 			if !m.Public && len(m.RecipientIDs) > 0 {
// 				u = m.RecipientIDs[0]
// 			}
// 		}

// 		_, err := searchLost(u)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 	})
// }

// func Test_findLeave(t *testing.T) {
// 	m := new(Moment)
// 	for _, v := range moments {
// 		if !v.Public && len(v.RecipientIDs) > 0 {
// 			m.RecipientID = v.RecipientIDs[0]
// 			m.ID = v.ID
// 		}
// 	}

// 	t.Run("1", func(t *testing.T) {

// 		if err := m.findLeave(); err != nil {
// 			t.Error(err)
// 		}
// 	})
// }

// func Test_searchFound(t *testing.T) {
// 	_, u, err := foundLeave()
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	t.Run("1", func(t *testing.T) {

// 		_, err := searchFound(u)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 	})

// }

// func Test_share(t *testing.T) {
// 	generateLeaveData := func() (m *Moment) {
// 		mID, u, err := foundLeave()
// 		if err != nil {
// 			t.Error(err)
// 		}

// 		m = new(Moment)
// 		m.RecipientID = u
// 		m.ID = mID

// 		r := rand.New(rand.NewSource(time.Now().UnixNano()))
// 		recipientIDs := make([]string, r.Intn(50)+1)
// 		for i, _ := range recipientIDs {
// 			recipientIDs[i] = "User_" + strconv.Itoa(i)
// 		}
// 		m.RecipientIDs = recipientIDs

// 		return
// 	}

// 	t.Run("1", func(t *testing.T) {
// 		m := generateLeaveData()

// 		if err := m.share(); err != nil {
// 			t.Error(err)
// 		}
// 	})
// }

// func Test_searchLeft(t *testing.T) {
// 	sID, err := senderID()
// 	if err != nil {
// 		t.Log(err)
// 	}

// 	t.Run("1", func(t *testing.T) {

// 		_, err := searchLeft(sID)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 	})
// }

// func Test_searchShared(t *testing.T) {
// 	u, err := sharedLeave()
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	t.Run("1", func(t *testing.T) {
// 		_, err := searchShared(u)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 	})
// }

// func Test_searchPublic(t *testing.T) {
// 	hidden := false
// 	l, err := publicMomentLocation(hidden)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	t.Run("1", func(t *testing.T) {

// 		_, err := searchPublic(l)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 	})
// }

// func Test_searchHiddenPublic(t *testing.T) {
// 	hidden := true
// 	l, err := publicMomentLocation(hidden)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	t.Run("1", func(t *testing.T) {

// 		_, err = searchHiddenPublic(l)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 	})
// }

// func Test_searchFoundPublic(t *testing.T) {

// 	u, err := finderID()
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	t.Run("1", func(t *testing.T) {

// 		_, err = searchFoundPublic(u)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 	})
// }

// =============================================================================
// func insertDataTestDb() (err error) {
// 	fmt.Printf("Generate Moment Test Data...\n\n")

// 	users = make([]string, *userCnt)
// 	for i, _ := range users {
// 		users[i] = "User_" + strconv.Itoa(i)
// 	}
// 	fmt.Printf("Test Users Generated...\n\n")

// 	moments = make([]*MomentsRow, *momentCnt)
// 	for i := 0; i < *momentCnt; i++ {
// 		r := rand.New(rand.NewSource(time.Now().UnixNano()))
// 		m := new(MomentsRow)

// 		m.UserID = users[i%*userCnt]
// 		m.CreateDate = time.Unix(r.Int63n(time.Now().Unix()), 0)
// 		m.Latitude = float32(r.Intn(180) - 90)
// 		m.Longitude = float32(r.Intn(360) - 180)length restrictions."

// 		if m.Type != DNE {
// 			m.MediaDir = "D:/mediaDir/"
// 		} else {
// 			m.MediaDir = ""
// 		}

// 		m.Public = (r.Intn(2) == 1)
// 		if m.Public {
// 			m.Hidden = (r.Intn(2) == 1)
// 		} else {
// 			m.Hidden = false
// 		}

// 		if !m.Public || (m.Public && m.Hidden) {
// 			if err = generateFinds(m); err != nil {
// 				return
// 			}
// 		}

// 		moments[i] = m
// 	}
// 	fmt.Printf("Test Moments Generated...\n\n")

// 	return
// }

// func generateFinds(m *MomentsRow) (err error) {
// 	rand.Seed(time.Now().UnixNano())
// 	findCnt := rand.Intn(*userCnt) + 1

// 	m.Finds = make([]*FindsRow, findCnt)
// 	for i := 0; i < findCnt; i++ {
// 		f := new(FindsRow)

// 		f.UserID = users[(findCnt+i)%*userCnt]
// 		f.Found = (rand.Intn(8) >= 1)
// 		f.FindDate = m.CreateDate.AddDate(0, 0, rand.Intn(14))

// 		if f.Found {
// 			if err = generateShares(f); err != nil {
// 				return
// 			}
// 		}

// 		m.Finds[i] = f
// 	}

// 	return
// }

// func generateShares(f *FindsRow) (err error) {

// 	rand.Seed(time.Now().UnixNano())
// 	shareCnt := rand.Intn(*userCnt/3) + 1
// 	userStart := rand.Intn(*userCnt)

// 	f.Shares = make([]*SharesRow, shareCnt)
// 	for i := 0; i < shareCnt; i++ {
// 		s := new(SharesRow)

// 		s.UserID = f.UserID
// 		s.All = (rand.Intn(2) == 1)

// 		if !s.All {
// 			s.RecipientID = users[(userStart+i)%*userCnt]
// 		}
// 		f.Shares[i] = s
// 	}

// 	return
// }

// func insertMomentMomentsData() (err error) {

// 	return
// }

// func insertMomentMediaData() (err error) {

// 	return
// }

// func insertMomentFindsData() (err error) {

// 	return
// }

// func insertMomentSharesData() (err error) {

// 	return
// }

// func deleteDataTestDb() (err error) {
// 	if err = testutil.TruncateTable("moment.Moments"); err != nil {
// 		return
// 	}
// 	if err = testutil.TruncateTable("moment.Media"); err != nil {
// 		return
// 	}
// 	if err = testutil.TruncateTable("moment.Finds"); err != nil {
// 		return
// 	}
// 	if err = testutil.TruncateTable("moment.Shares"); err != nil {
// 		return
// 	}

// 	return
// }

// // =============================================================================

// func publicHiddenMoment() (id int, err error) {
// 	db := openDbConn()
// 	defer db.Close()

// 	query := `SELECT ID
// 			  FROM moment.Moments
// 			  WHERE [Public] = 1
// 			  		AND [Hidden] = 1`
// 	rows, err := db.Query(query)
// 	if err != nil {
// 		return
// 	}
// 	defer db.Close()

// 	var idS []int
// 	var idTemp int
// 	for rows.Next() {
// 		if err = rows.Scan(&idTemp); err != nil {
// 			return
// 		}
// 		idS = append(idS, idTemp)
// 	}
// 	if err = rows.Err(); err != nil {
// 		return
// 	}

// 	rand.Seed(time.Now().UnixNano())
// 	id = idS[rand.Intn(len(idS))]

// 	return
// }

// func publicMomentLocation(hidden bool) (l Location, err error) {
// 	db := openDbConn()
// 	defer db.Close()

// 	query := `SELECT Latitude, Longitude
// 			  FROM [moment].[Moments]
// 			  WHERE [Public] = 1
// 			  		AND `
// 	if hidden {
// 		query = query + `[Hidden] = 1`
// 	} else {
// 		query = query + `[Hidden] = 0`
// 	}

// 	rows, err := db.Query(query)
// 	if err != nil {
// 		return
// 	}
// 	defer rows.Close()

// 	var idSlice []Location
// 	var lat, long float32
// 	for rows.Next() {
// 		if err = rows.Scan(&lat, &long); err != nil {
// 			return
// 		}

// 		idSlice = append(idSlice, Location{lat, long})
// 	}
// 	if err = rows.Err(); err != nil {
// 		return
// 	}

// 	rand.Seed(time.Now().UnixNano())
// 	l = idSlice[rand.Intn(len(idSlice))]

// 	return
// }

// // UserWithFound selects a user that has a found moment.
// func foundLeave() (mID int, u string, err error) {
// 	db := openDbConn()
// 	defer db.Close()

// 	query := `SELECT ID, MomentID, RecipientID
// 			  FROM [moment].[Leaves]
// 			  WHERE Found = 1
// 			  		AND Shared = 0`
// 	rows, err := db.Query(query)
// 	if err != nil {
// 		return 0, "", err
// 	}
// 	defer rows.Close()

// 	var lID int
// 	m := make(map[int][]interface{})
// 	keys := make([]int, 0)
// 	for rows.Next() {
// 		if err = rows.Scan(&lID, &mID, &u); err != nil {
// 			return 0, "", err
// 		}
// 		m[lID] = []interface{}{mID, u}
// 		keys = append(keys, lID)
// 	}
// 	rand.Seed(time.Now().UnixNano())
// 	key := keys[rand.Intn(len(keys))]
// 	iSlice, ok := m[key]
// 	if !ok {
// 		return 0, "", errors.New("Invalid map key.")
// 	}
// 	u = iSlice[1].(string)
// 	mID = iSlice[0].(int)

// 	return mID, u, nil
// }

// func sharedLeave() (u string, err error) {
// 	db := openDbConn()
// 	defer db.Close()

// 	query := `SELECT RecipientID
// 			  FROM [moment].[Leaves]
// 			  WHERE shared = 1`
// 	rows, err := db.Query(query)
// 	if err != nil {
// 		return
// 	}
// 	defer rows.Close()

// 	var rs []string
// 	var r string
// 	for rows.Next() {
// 		if err = rows.Scan(&r); err != nil {
// 			return
// 		}
// 		rs = append(rs, r)
// 	}
// 	if err = rows.Err(); err != nil {
// 		return
// 	}

// 	rand.Seed(time.Now().UnixNano())
// 	u = rs[rand.Intn(len(rs))]
// 	return
// }

// func senderID() (senderID string, err error) {
// 	db := openDbConn()
// 	defer db.Close()

// 	query := `SELECT SenderID
// 			  FROM [moment].[Moments]`
// 	rows, err := db.Query(query)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer rows.Close()

// 	var senders []string
// 	for rows.Next() {
// 		if err = rows.Scan(&senderID); err != nil {
// 			return "", err
// 		}
// 		senders = append(senders, senderID)
// 	}
// 	rand.Seed(time.Now().UnixNano())
// 	senderID = senders[rand.Intn(len(senders))]

// 	return senderID, nil
// }

// func finderID() (finderID string, err error) {
// 	db := openDbConn()
// 	defer db.Close()

// 	query := `SELECT FinderID
// 			  FROM [moment].[Finds]`
// 	rows, err := db.Query(query)
// 	if err != nil {
// 		return
// 	}
// 	defer rows.Close()

// 	var finders []string
// 	for rows.Next() {
// 		if err = rows.Scan(&finderID); err != nil {
// 			return
// 		}
// 		finders = append(finders, finderID)
// 	}
// 	if err = rows.Err(); err != nil {
// 		return
// 	}

// 	rand.Seed(time.Now().UnixNano())
// 	finderID = finders[rand.Intn(len(finders))]

// 	return
// }

// // BENCHMARKS
