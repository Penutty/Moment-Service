package moment

import (
	// "errors"
	"flag"
	// "fmt"
	"github.com/stretchr/testify/assert"
	// "math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	// "testutil"
	"time"
)

var userCnt = flag.Int("userCnt", 100, "The number of unique users to be created in Moment-Db for testing.")
var momentCnt = flag.Int("momentCnt", 1000, "The number of moments to be created in Moment-Db for testing.")

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

	test("user123", nil)
	test("", ErrorUserIDEmpty)
	test("user", ErrorUserIDShort)
	test(strings.Repeat("c", maxUserChars), nil)
	test(strings.Repeat("c", maxUserChars+1), ErrorUserIDLong)
}

func TestCheckMediaMessage(t *testing.T) {
	test := func(m string, expected error) {

		name := errorName(expected)
		t.Run(name, func(t *testing.T) {
			err := checkMediaMessage(m)
			assert.Exactly(t, expected, err)
		})
	}

	test("", nil)
	test("message", nil)
	test(strings.Repeat("c", maxMessage), nil)
	test(strings.Repeat("c", maxMessage+1), ErrorMediaMessageLong)
}

func TestCheckMediaDir(t *testing.T) {
	test := func(ty uint8, d string, expected error) {

		name := errorName(expected)
		t.Run(name, func(t *testing.T) {
			err := checkMediaDir(ty, d)
			assert.Exactly(t, expected, err)
		})
	}

	test(0, "", nil)
	test(0, "D:/Dir/", ErrorNoMediaTypeHasDir)
	test(1, "", ErrorMediaTypeNoDir)
	test(2, "", ErrorMediaTypeNoDir)
	test(3, "", ErrorMediaTypeNoDir)
	test(1, "D:/Dir/", nil)
	test(2, "D:/Dir/", nil)
	test(3, "D:/Dir/", nil)
}

func TestCheckTime(t *testing.T) {
	test := func(f bool, fd *time.Time, expected error) {

		name := errorName(expected)
		t.Run(name, func(t *testing.T) {
			err := checkTime(f, fd)
			assert.Exactly(t, expected, err)
		})
	}

	fd := time.Now().UTC()
	test(false, nil, nil)
	test(false, &fd, ErrorFindDateWithFalseFound)
	test(true, &fd, nil)
	test(true, nil, ErrorFindDateDNEWithFound)
}

func TestValidateShareAll(t *testing.T) {
	test := func(a bool, r string, expected error) {

		name := errorName(expected)
		t.Run(name, func(t *testing.T) {
			err := validateShareAll(a, r)
			assert.Exactly(t, expected, err)
		})
	}

	test(false, "user123", nil)
	test(false, "", ErrorShareAllNoRecipients)
	test(true, "", nil)
	test(true, "user123", ErrorShareAllPublicWithRecipients)
}

func TestValidateLatitude(t *testing.T) {
	test := func(l float32, expected error) {

		name := errorName(expected)
		t.Run(name, func(t *testing.T) {
			err := validateLatitude(l)
			assert.Exactly(t, expected, err)
		})
	}

	test(0, nil)
	test(-180.00, nil)
	test(-181.00, ErrorLatitude)
	test(180.00, nil)
	test(181.00, ErrorLatitude)
}

func TestValidateLongitude(t *testing.T) {
	test := func(l float32, expected error) {

		name := errorName(expected)
		t.Run(name, func(t *testing.T) {
			err := validateLongitude(l)
			assert.Exactly(t, expected, err)
		})
	}

	test(0, nil)
	test(-90, nil)
	test(-91.00, ErrorLongitude)
	test(90.00, nil)
	test(91.00, ErrorLongitude)
}

func TestValidateLocation(t *testing.T) {
	test := func(l *Location, expected error) {
		name := errorName(expected)
		t.Run(name, func(t *testing.T) {
			err := validateLocation(l)
			assert.Exactly(t, expected, err)
		})
	}

	l := &Location{
		latitude:  0.00,
		longitude: 0.00,
	}
	test(l, nil)
	test(nil, ErrorLocationReference)
}

func TestValidateMomentPublicHidden(t *testing.T) {
	test := func(p, h bool, expected error) {
		name := errorName(expected)
		t.Run(name, func(t *testing.T) {
			err := validateMomentPublicHidden(p, h)
			assert.Exactly(t, expected, err)
		})
	}

	test(false, false, nil)
	test(false, true, ErrorPublicHiddenCombination)
	test(true, false, nil)
	test(true, true, nil)
}

var start int = 1
var length int = 500

func TestFindsInsertDelete(t *testing.T) {
	fS, err := newFinds()
	assert.Nil(t, err)

	t.Run("insert", func(t *testing.T) {
		cnt, err := fS.insert(nil)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, length, cnt)
	})

	t.Run("delete", func(t *testing.T) {
		cnt, err := fS.delete()
		assert.Nil(t, err)
		assert.Equal(t, length, cnt)
	})
}

func newFinds() (fS Finds, err error) {
	fS = make(Finds, length)

	for i := 0; i < length; i++ {
		n := start + i
		td := time.Now().UTC()
		fS[i], err = NewFind(n, "user_0"+strconv.Itoa(n), true, &td)
	}

	return
}

func TestFindsRowDelete(t *testing.T) {
	td := time.Now().UTC()
	f, err := NewFind(1, "user_00", true, &td)
	assert.Nil(t, err)

	fS := make(Finds, 1)
	fS[0] = f

	t.Run("insert", func(t *testing.T) {
		cnt, err := fS.insert(nil)
		assert.Equal(t, 1, cnt)
		assert.Nil(t, err)
	})

	t.Run("delete", func(t *testing.T) {
		cnt, err := fS.delete()
		assert.Equal(t, 1, cnt)
		assert.Nil(t, err)
	})
}

func TestSharesInsertDelete(t *testing.T) {
	sS, err := newShares()
	assert.Nil(t, err)

	t.Run("insert", func(t *testing.T) {
		cnt, err := sS.insert(nil)
		assert.Nil(t, err)
		assert.Equal(t, length, cnt)
	})

	t.Run("delete", func(t *testing.T) {
		cnt, err := sS.delete()
		assert.Empty(t, err)
		assert.Equal(t, length, cnt)
	})
}

func newShares() (sS Shares, err error) {
	sS = make(Shares, length)

	for i := 0; i < length; i++ {
		n := start + i
		sS[i], err = NewShare(n, "user_0"+strconv.Itoa(n), false, "user_1"+strconv.Itoa(n))
	}
	return
}

func TestSharesRowDelete(t *testing.T) {
	s, err := NewShare(1, "user_01", true, "")
	assert.Nil(t, err)

	sS := make(Shares, 1)
	sS[0] = s

	t.Run("insert", func(t *testing.T) {
		cnt, err := sS.insert(nil)
		assert.Equal(t, 1, cnt)
		assert.Nil(t, err)
	})

	t.Run("delete", func(t *testing.T) {
		cnt, err := sS[0].delete(nil)
		assert.Equal(t, 1, cnt)
		assert.Nil(t, err)
	})
}

func TestMediaInsertDelete(t *testing.T) {
	mS, err := newMedia()
	assert.Nil(t, err)

	t.Run("insert", func(t *testing.T) {
		cnt, err := mS.insert(nil)
		assert.Equal(t, length, cnt)
		assert.Nil(t, err)
	})

	t.Run("delete", func(t *testing.T) {
		cnt, err := mS.delete()
		assert.Equal(t, length, cnt)
		assert.Nil(t, err)
	})
}

func newMedia() (mS Media, err error) {
	mS = make(Media, length)

	for i := 0; i < length; i++ {
		n := start + i
		mS[i], err = NewMedia(n, "message_0"+strconv.Itoa(n), 0, "")
	}

	return
}

func TestMediaRowDelete(t *testing.T) {
	m, err := NewMedia(1, "message", DNE, "")
	assert.Nil(t, err)

	mS := make(Media, 1)
	mS[0] = m

	t.Run("insert", func(t *testing.T) {
		cnt, err := mS.insert(nil)
		assert.Equal(t, 1, cnt)
		assert.Nil(t, err)
	})

	t.Run("delete", func(t *testing.T) {
		cnt, err := mS[0].delete(nil)
		assert.Equal(t, 1, cnt)
		assert.Nil(t, err)
	})
}

func TestMomentsRowInsertDelete(t *testing.T) {

	l, err := NewLocation(0.00, 0.00)
	assert.Nil(t, err)
	m, err := NewMoment(l, "user_01", true, false, time.Now().UTC())
	assert.Nil(t, err)

	t.Run("insert", func(t *testing.T) {
		id, err := m.insert(nil)
		assert.NotZero(t, id)
		assert.Nil(t, err)
	})

	t.Run("delete", func(t *testing.T) {
		cnt, err := m.delete(nil)
		assert.Equal(t, int64(1), cnt)
		assert.Nil(t, err)
	})
}

var lat float32 = 0.00
var long float32 = 0.00

func BenchmarkMomentsRowNew(b *testing.B) {
	l, _ := NewLocation(lat, long)
	_, _ = NewMoment(l, "user_01", true, false, time.Now().UTC())
}

func BenchmarkMomentsRowInsert(b *testing.B) {
	l, _ := NewLocation(lat, long)
	m, _ := NewMoment(l, "user_01", true, false, time.Now().UTC())

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
	sS, _ := newShares()

	b.ResetTimer()
	sS.insert(nil)
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
	sS, _ := newShares()

	sS.insert(nil)
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
// 		m.Longitude = float32(r.Intn(360) - 180)
// 		m.Type = uint8(r.Intn(4))
// 		m.Message = "This message must be less than 256 characters. May adjust length restrictions."

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
