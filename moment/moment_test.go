package moment

import (
	// "errors"
	"flag"
	"github.com/stretchr/testify/assert"
	// "math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	// "testutil"
	"time"
)

const (
	TestUser      string = "user00"
	TestEmptyUser string = ""

	lat  float32 = 0.00
	long float32 = 0.00

	base int = 1
)

func errorName(err error) (name string) {
	if err == nil {
		name = "nil error"
	} else {
		name = err.Error()
	}
	return
}

func TestMain(m *testing.M) {
	flag.Parse()

	call := m.Run()

	os.Exit(call)
}

func TestCheckTime(t *testing.T) {
	test := func(td *time.Time, expected error) {
		name := errorName(expected)
		t.Run(name, func(t *testing.T) {
			err := checkTime(td)
			assert.Exactly(t, expected, err)
		})
	}

	test(nil, ErrorTimePtrNil)
	td := time.Now().UTC()
	test(&td, nil)
}

func TestCheckMomentID(t *testing.T) {
	test := func(id int64, expected error) {

		name := errorName(expected)
		t.Run(name, func(t *testing.T) {
			err := checkMomentID(id)
			assert.Exactly(t, expected, err)
		})
	}

	test(1, nil)
	test(0, ErrorMomentID)
	test(100, nil)
	test(1000000, nil)

}

func TestCheckMediaType(t *testing.T) {
	test := func(ty uint8, expected error) {
		name := errorName(expected)
		t.Run(name, func(t *testing.T) {
			err := checkMediaType(ty)
			assert.Exactly(t, expected, err)
		})
	}

	test(0, nil)
	test(1, nil)
	test(2, nil)
	test(3, nil)
	test(4, ErrorMediaTypeDNE)
	test(200, ErrorMediaTypeDNE)
}

func TestCheckUserID(t *testing.T) {
	test := func(id string, expected error) {

		name := errorName(expected)
		t.Run(name, func(t *testing.T) {
			err := checkUserID(id)
			assert.Exactly(t, expected, err)
		})
	}

	test(TestUser, nil)
	test(TestEmptyUser, ErrorUserIDShort)
	test("user", ErrorUserIDShort)
	test(strings.Repeat("c", maxUserChars), nil)
	test(strings.Repeat("c", maxUserChars+1), ErrorUserIDLong)
}

func TestNewLocation(t *testing.T) {
	test := func(lat float32, long float32, expected error) {

		name := errorName(expected)
		t.Run(name, func(t *testing.T) {
			_, err := NewLocation(lat, long)
			assert.Exactly(t, expected, err)
		})
	}

	test(0, 0, nil)
	test(minLat, 0, nil)
	test(minLat-1, 0, ErrorLatitude)
	test(maxLat, 0, nil)
	test(maxLat+1, 0, ErrorLatitude)
	test(0, minLong, nil)
	test(0, minLong-1, ErrorLongitude)
	test(0, maxLong, nil)
	test(0, maxLong+1, ErrorLongitude)
}

func TestNewMedia(t *testing.T) {
	test := func(mID int64, m string, mType uint8, d string, expected error) {

		name := errorName(expected)
		t.Run(name, func(t *testing.T) {
			_, err := NewMedia(mID, m, mType, d)
			assert.Exactly(t, expected, err)
		})
	}

	test(1, "message", DNE, "", nil)
	test(1, strings.Repeat("c", maxMessage), DNE, "", nil)
	test(1, strings.Repeat("c", maxMessage+1), DNE, "", ErrorMessageLong)
	test(1, "message", DNE, "D:/dir/", ErrorMediaDNE)
	test(1, "message", Image, "D:/dir/", nil)
	test(1, "message", Image, "", ErrorMediaExistsDirDNE)
}

func TestNewFind(t *testing.T) {
	test := func(mID int64, uID string, f bool, fd *time.Time, expected error) {

		name := errorName(expected)
		t.Run(name, func(t *testing.T) {
			_, err := NewFind(mID, uID, f, fd)
			assert.Exactly(t, expected, err)
		})
	}

	fd := time.Now().UTC()
	test(1, TestUser, true, &fd, nil)
	test(1, TestUser, true, &time.Time{}, ErrorFoundEmptyFindDate)
	test(1, TestUser, false, &time.Time{}, nil)
	test(1, TestUser, false, &fd, ErrorNotFoundFindDateExists)
}

func TestNewShare(t *testing.T) {
	test := func(mID int64, uID string, all bool, r string, expected error) {

		name := errorName(expected)
		t.Run(name, func(t *testing.T) {
			_, err := NewShare(mID, uID, all, r)
			assert.Exactly(t, expected, err)
		})
	}

	test(1, TestUser, false, "user_02", nil)
	test(1, TestUser, true, "user_02", ErrorAllRecipientExists)
	test(1, TestUser, true, "", nil)
	test(1, TestUser, false, "", ErrorNotAllRecipientDNE)
}
func TestNewMoment(t *testing.T) {

	test := func(l *Location, uID string, p bool, h bool, c *time.Time, expected error) {

		name := errorName(expected)
		t.Run(name, func(t *testing.T) {
			_, err := NewMoment(l, uID, p, h, c)
			assert.Exactly(t, expected, err)
		})
	}

	lo, _ := NewLocation(lat, long)
	cd := time.Now().UTC()

	test(lo, TestUser, true, false, &cd, nil)
	test(nil, TestUser, true, false, &cd, ErrorLocationIsNil)
	test(lo, TestUser, true, true, &cd, nil)
	test(lo, TestUser, false, false, &cd, nil)
	test(lo, TestUser, false, true, &cd, ErrorPrivateHiddenMoment)
}

const (
	single int = 1
	start  int = 1
	length int = 500
)

func TestFindsInsertDelete(t *testing.T) {
	fS, err := newFinds()
	assert.Nil(t, err)

	subTestFindsInsert(t, fS)
	subTestFindsDelete(t, fS)
}

func TestFindsRowDelete(t *testing.T) {
	td := time.Now().UTC()
	f, err := NewFind(1, TestUser, true, &td)
	assert.Nil(t, err)

	fS := make(Finds, single)
	fS[0] = f

	subTestFindsInsert(t, fS)
	subTestFindsDelete(t, fS)
}

func newFinds() (fS Finds, err error) {
	fS = make(Finds, length)

	for i := 0; i < length; i++ {
		n := start + i
		td := time.Now().UTC()
		fS[i], err = NewFind(int64(n), TestUser+strconv.Itoa(n), true, &td)
	}

	return
}

func subTestFindsInsert(t *testing.T, fS Finds) {
	t.Run("insert", func(t *testing.T) {
		cnt, err := fS.insert(nil)
		assert.Nil(t, err)
		assert.Equal(t, int64(len(fS)), cnt)
	})
}

func subTestFindsDelete(t *testing.T, fS Finds) {
	t.Run("delete", func(t *testing.T) {
		cnt, err := fS.delete()
		assert.Nil(t, err)
		assert.Equal(t, int64(len(fS)), cnt)
	})
}

func TestSharesInsertDelete(t *testing.T) {
	sS, err := newShares(base, length, false)
	assert.Nil(t, err)

	subTestSharesRowInsert(t, sS)
	subTestSharesRowDelete(t, sS)
}

func TestSharesRowDelete(t *testing.T) {
	sS, err := newShares(base, single, true)
	assert.Nil(t, err)

	subTestSharesRowInsert(t, sS)
	subTestSharesRowDelete(t, sS)
}

func newShares(start int, cnt int, all bool) (sS Shares, err error) {
	sS = make(Shares, cnt)

	for i := 0; i < cnt; i++ {
		n := start + i
		if all {
			sS[i], err = NewShare(int64(n), TestUser+strconv.Itoa(n), true, TestEmptyUser)
		} else {
			sS[i], err = NewShare(int64(n), TestUser+strconv.Itoa(n), false, TestUser+strconv.Itoa(n+1))
		}
	}
	return
}

func subTestSharesRowInsert(t *testing.T, sS Shares) {
	t.Run("insert", func(t *testing.T) {
		cnt, err := sS.insert()
		assert.Nil(t, err)
		assert.Equal(t, int64(len(sS)), cnt)
	})
}

func subTestSharesRowDelete(t *testing.T, sS Shares) {
	t.Run("delete", func(t *testing.T) {
		cnt, err := sS.delete()
		assert.Empty(t, err)
		assert.Equal(t, int64(len(sS)), cnt)
	})
}

func TestMediaInsertDelete(t *testing.T) {
	mS, err := newMedia()
	assert.Nil(t, err)

	subTestMediaInsert(t, mS)
	subTestMediaDelete(t, mS)
}

func TestMediaRowDelete(t *testing.T) {
	m, err := NewMedia(1, "message", DNE, "")
	assert.Nil(t, err)

	mS := make(Media, single)
	mS[0] = m

	subTestMediaInsert(t, mS)
	subTestMediaDelete(t, mS)
}

func newMedia() (mS Media, err error) {
	mS = make(Media, length)

	for i := 0; i < length; i++ {
		n := start + i
		mS[i], err = NewMedia(int64(n), "message_0"+strconv.Itoa(n), 0, "")
	}

	return
}

func subTestMediaInsert(t *testing.T, mS Media) {
	t.Run("insert", func(t *testing.T) {
		cnt, err := mS.insert(nil)
		assert.Equal(t, int64(len(mS)), cnt)
		assert.Nil(t, err)
	})
}

func subTestMediaDelete(t *testing.T, mS Media) {
	t.Run("delete", func(t *testing.T) {
		cnt, err := mS.delete()
		assert.Equal(t, int64(len(mS)), cnt)
		assert.Nil(t, err)
	})
}

func TestMomentsRowInsertDelete(t *testing.T) {

	l, err := NewLocation(lat, long)
	assert.Nil(t, err)
	td := time.Now().UTC()
	m, err := NewMoment(l, TestUser, true, false, &td)
	assert.Nil(t, err)

	t.Run("insert", func(t *testing.T) {
		id, err := m.insert(nil)
		assert.NotZero(t, id)
		assert.Nil(t, err)
	})

	t.Run("delete", func(t *testing.T) {
		cnt, err := m.delete(nil)
		assert.Equal(t, int64(single), cnt)
		assert.Nil(t, err)
	})
}

func TestMomentsRowCreatePublic(t *testing.T) {
	r := MomentsRowCreatePublic(t, false)
	MomentsRowDeletePublic(t, r)
}

func MomentsRowCreatePublic(t *testing.T, hidden bool) *Result {
	l, err := NewLocation(lat, long)
	assert.Nil(t, err)

	td := time.Now().UTC()
	m, err := NewMoment(l, TestUser, true, hidden, &td)
	assert.Nil(t, err)

	mediaCnt := 1
	media := make(Media, mediaCnt)
	for i := 0; i < len(media); i++ {
		med, err := NewMedia(1, "HelloWorld", DNE, "")
		assert.Nil(t, err)
		media[i] = med
	}

	err = m.CreatePublic(&media)
	assert.Nil(t, err)

	return &Result{
		moment: m,
		media:  media,
	}
}

func MomentsRowDeletePublic(t *testing.T, r *Result) {
	cnt, err := r.media.delete()
	assert.Nil(t, err)
	assert.Equal(t, int64(len(r.media)), cnt)

	cnt, err = r.moment.delete(nil)
	assert.Nil(t, err)
	assert.Equal(t, int64(single), cnt)

}

func TestMomentsRowCreatePrivate(t *testing.T) {
	r := MomentsRowCreatePrivate(t)
	MomentsRowDeletePrivate(t, r)
}

func MomentsRowCreatePrivate(t *testing.T) *Result {
	l, err := NewLocation(lat, long)
	assert.Nil(t, err)

	td := time.Now().UTC()
	m, err := NewMoment(l, TestUser, false, false, &td)
	assert.Nil(t, err)

	mediaCnt := 1
	media := make(Media, mediaCnt)
	for i := 0; i < len(media); i++ {
		med, err := NewMedia(1, "HelloWorld", DNE, "")
		assert.Nil(t, err)
		media[i] = med
	}

	findsCnt := 10
	finds := make(Finds, findsCnt)
	for i := 0; i < findsCnt; i++ {
		find, err := NewFind(1, TestUser+strconv.Itoa(i), false, &time.Time{})
		assert.Nil(t, err)
		finds[i] = find
	}

	err = m.CreatePrivate(&media, &finds)
	assert.Nil(t, err)

	return &Result{
		moment: m,
		finds:  finds,
		media:  media,
	}

}

func MomentsRowDeletePrivate(t *testing.T, r *Result) {
	cnt, err := r.media.delete()
	assert.Nil(t, err)
	assert.Equal(t, int64(len(r.media)), cnt)

	cnt, err = r.finds.delete()
	assert.Nil(t, err)
	assert.Equal(t, int64(len(r.finds)), cnt)

	cnt, err = r.moment.delete(nil)
	assert.Nil(t, err)
	assert.Equal(t, int64(single), cnt)

}

func TestFindsRowFindPublic(t *testing.T) {
	r := MomentsRowCreatePublic(t, false)

	fd := time.Now().UTC()
	f, err := NewFind(r.moment.momentID, TestUser+"2", true, &fd)

	cnt, err := f.FindPublic()
	assert.Nil(t, err)
	assert.Equal(t, int64(single), cnt)

	MomentsRowDeletePublic(t, r)
}

func TestFindsRowFindPrivate(t *testing.T) {
	r := MomentsRowCreatePrivate(t)

	// use momentID and userID to find a private moment
	err := r.finds[0].FindPrivate()
	assert.Nil(t, err)

	MomentsRowDeletePrivate(t, r)
}

func TestQueryLocationShared(t *testing.T) {
	r := MomentsRowCreatePrivate(t)

	f := r.finds[0]
	sS := make(Shares, single)
	s, err := NewShare(f.momentID, f.userID, true, TestEmptyUser)
	assert.Nil(t, err)
	sS[0] = s

	cnt, err := sS.Share()
	assert.Nil(t, err)
	assert.Equal(t, int64(single), cnt)

	l, err := NewLocation(lat, long)
	assert.Nil(t, err)

	res, err := QueryLocationShared(l, TestUser+"0")
	t.Log(res)
	assert.Nil(t, err)
	assert.Equal(t, TestUser, res[0].moment.userID)
	assert.Equal(t, lat, res[0].moment.latitude)
	assert.Equal(t, long, res[0].moment.longitude)
	assert.Equal(t, uint8(DNE), res[0].media[0].mType)
	assert.Equal(t, "HelloWorld", res[0].media[0].message)
	assert.Equal(t, "", res[0].media[0].dir)

	subTestSharesRowDelete(t, sS)
	MomentsRowDeletePrivate(t, r)
}

func TestQueryLocationPublic(t *testing.T) {
	hidden := false
	r := MomentsRowCreatePublic(t, hidden)

	l, err := NewLocation(lat, long)
	assert.Nil(t, err)

	res, err := QueryLocationPublic(l)
	assert.Nil(t, err)
	assert.Equal(t, TestUser, res[0].moment.userID)
	assert.Equal(t, lat, res[0].moment.latitude)
	assert.Equal(t, long, res[0].moment.longitude)
	assert.Equal(t, uint8(DNE), res[0].media[0].mType)
	assert.Equal(t, "HelloWorld", res[0].media[0].message)
	assert.Equal(t, "", res[0].media[0].dir)

	MomentsRowDeletePublic(t, r)
}

func TestQueryLocationHidden(t *testing.T) {
	hidden := true
	r := MomentsRowCreatePublic(t, hidden)

	l, err := NewLocation(lat, long)
	assert.Nil(t, err)

	res, err := QueryLocationHidden(l)
	assert.Nil(t, err)
	assert.Equal(t, lat, res[0].moment.latitude)
	assert.Equal(t, long, res[0].moment.longitude)

	MomentsRowDeletePublic(t, r)
}

func TestQueryLocationLost(t *testing.T) {
	r := MomentsRowCreatePrivate(t)

	l, err := NewLocation(lat, long)
	assert.Nil(t, err)

	me := TestUser + "1"
	res, err := QueryLocationLost(l, me)
	t.Logf("res = %v\n", res)
	assert.Nil(t, err)
	assert.Equal(t, lat, res[0].moment.latitude)
	assert.Equal(t, long, res[0].moment.longitude)

	MomentsRowDeletePrivate(t, r)
}

func BenchmarkMomentsRowNew(b *testing.B) {
	l, _ := NewLocation(lat, long)
	td := time.Now().UTC()
	_, _ = NewMoment(l, "user_01", true, false, &td)
}

func BenchmarkMomentsRowInsert(b *testing.B) {
	l, _ := NewLocation(lat, long)
	td := time.Now().UTC()
	m, _ := NewMoment(l, "user_01", true, false, &td)

	b.ResetTimer()
	m.insert(nil)
	b.StopTimer()
	m.delete(nil)
}

func BenchmarkFindsInsert(b *testing.B) {
	fS, _ := newFinds()

	b.ResetTimer()
	fS.insert(nil)
	b.StopTimer()
	fS.delete()
}

func BenchmarkSharesInsert(b *testing.B) {
	sS, _ := newShares(base, length, false)

	b.ResetTimer()
	sS.insert()
	b.StopTimer()
	sS.delete()
}

func BenchmarkMediaInsert(b *testing.B) {
	mS, _ := newMedia()

	b.ResetTimer()
	mS.insert(nil)
	b.StopTimer()
	mS.delete()
}

func BenchmarkSharesDelete(b *testing.B) {
	b.StopTimer()
	sS, _ := newShares(base, length, false)

	sS.insert()
	b.StartTimer()

	sS.delete()
}

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
