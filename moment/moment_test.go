package moment

import (
	// "errors"
	"fmt"
	// "github.com/stretchr/testify/assert"
	"flag"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"testutil"
	"time"
)

var userCnt = flag.Int("userCnt", 100, "The number of unique users to be created in Moment-Db for testing.")
var momentCnt = flag.Int("momentCnt", 1000, "The number of moments to be created in Moment-Db for testing.")

var users []string
var moments []MomentsRow

func TestMain(m *testing.M) {
	flag.Parse()

	if err := insertDataTestDb(); err != nil {
		panic(err)
	}

	fmt.Printf("moments[0]:\n%v\n", moments[0])

	call := m.Run()

	fmt.Printf("moments[0]:\n%v\n", moments[0])

	if err := deleteDataTestDb(); err != nil {
		panic(err)
	}

	os.Exit(call)
}

func Test_Moment_leave(t *testing.T) {
	for i, _ := range moments {

		t.Run(strconv.Itoa(i), func(t *testing.T) {
			// t.Parallel()
			if err := moments[i].create(); err != nil {
				t.Error(err)
			}
		})
	}
}

// func Test_Moment_share(t *testing.T) {
// 	for i, m := range moments {
// 		if m.Finds == nil {
// 			continue
// 		}
// 		for j, v := range *m.Finds {
// 			f := v

// 			t.Run(strconv.Itoa(i)+"_"+strconv.Itoa(j), func(t *testing.T) {
// 				t.Parallel()
// 				if err := f.share(); err != nil {
// 					t.Error(err)
// 				}
// 			})
// 		}
// 	}
// }

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
func insertDataTestDb() (err error) {
	fmt.Printf("Generate Moment Test Data...\n\n")

	users = make([]string, *userCnt)
	for i, _ := range users {
		users[i] = "User_" + strconv.Itoa(i)
	}
	fmt.Printf("Test Users Generated...\n\n")

	moments = make([]MomentsRow, *momentCnt)
	for i, _ := range moments {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		moments[i].UserID = users[i%*userCnt]
		moments[i].CreateDate = time.Unix(r.Int63n(time.Now().Unix()), 0)
		moments[i].Latitude = float32(r.Intn(180) - 90)
		moments[i].Longitude = float32(r.Intn(360) - 180)
		moments[i].Type = uint8(r.Intn(4))
		moments[i].Message = "This message must be less than 256 characters. May adjust length restrictions."

		if moments[i].Type != DNE {
			moments[i].MediaDir = "D:/mediaDir/"
		} else {
			moments[i].MediaDir = ""
		}

		moments[i].Public = (r.Intn(2) == 1)
		if moments[i].Public {
			moments[i].Hidden = (r.Intn(2) == 1)
		} else {
			moments[i].Hidden = false
		}

		if !moments[i].Public || (moments[i].Public && moments[i].Hidden) {
			if err = generateFinds(&moments[i]); err != nil {
				return
			}
		}
	}
	fmt.Printf("Test Moments Generated...\n\n")

	return
}

func generateFinds(m *MomentsRow) (err error) {
	rand.Seed(time.Now().UnixNano())
	findCnt := rand.Intn(*userCnt)

	finds := make([]FindsRow, findCnt)
	for i := 0; i < findCnt; i++ {
		finds[i].UserID = users[(findCnt+i)%*userCnt]
		finds[i].Found = (rand.Intn(2) == 1)
		finds[i].FindDate = m.CreateDate.AddDate(0, 0, rand.Intn(14))

		if err = generateShares(&finds[i]); err != nil {
			return
		}
	}
	m.Finds = &finds

	return
}

func generateShares(f *FindsRow) (err error) {

	rand.Seed(time.Now().UnixNano())
	shareCnt := rand.Intn(*userCnt)

	shares := make([]SharesRow, shareCnt)
	for j := 0; j < shareCnt; j++ {
		shares[j].UserID = f.UserID
		shares[j].All = (rand.Intn(2) == 1)

		if !shares[j].All {
			shares[j].RecipientID = users[(shareCnt+j)%*userCnt]
		}
	}
	f.Shares = &shares

	return
}

func deleteDataTestDb() (err error) {
	if err = testutil.TruncateTable("moment.Moments"); err != nil {
		return
	}
	if err = testutil.TruncateTable("moment.Media"); err != nil {
		return
	}
	if err = testutil.TruncateTable("moment.Finds"); err != nil {
		return
	}
	if err = testutil.TruncateTable("moment.Shares"); err != nil {
		return
	}

	return
}

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
