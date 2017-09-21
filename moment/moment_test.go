package moment

import (
	"errors"
	"fmt"
	// "github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"testutil"
	"time"
)

const (
	Latitude_1  = 80
	Longitude_1 = -80
)

var users []string

func TestMain(m *testing.M) {
	users = make([]string, 500)
	for i, _ := range users {
		users[i] = "User_" + strconv.Itoa(i)
	}

	call := m.Run()

	var err error
	err = testutil.TruncateTable("[moment].[Moments]")
	err = testutil.TruncateTable("[moment].[Leaves]")
	err = testutil.TruncateTable("[moment].[Shares]")
	err = testutil.TruncateTable("[moment].[Media]")
	if err != nil {
		fmt.Printf("err = %v", err)
	}

	os.Exit(call)
}

func Test_Moment_leave(t *testing.T) {

	generateLeaveData := func() (m *Moment) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		u := users[r.Intn(len(users))]
		latitude := float32(r.Intn(360) - 180)
		longitude := float32(r.Intn(180) - 90)

		t := uint8(r.Intn(4))
		p := r.Intn(2)
		h := r.Intn(2)

		m = new(Moment)
		m.SenderID = u
		m.CreateDate = time.Now().UTC()
		m.Latitude = latitude
		m.Longitude = longitude
		m.Type = t
		m.Message = "message"
		if m.Type == Image || m.Type == Video {
			m.MediaDir = "D:/dir/"
		} else {
			m.MediaDir = ""
		}
		m.Public = (p == 1)
		if !m.Public {
			m.Hidden = false
		} else {
			m.Hidden = (h == 1)
		}

		m.Shared = false
		if !m.Public {
			recipientIDs := make([]string, r.Intn(50)+1)
			for i, _ := range recipientIDs {
				recipientIDs[i] = "User_" + strconv.Itoa(i)
			}
			m.RecipientIDs = append(m.RecipientIDs, recipientIDs...)
		} else {
			m.RecipientIDs = nil

		}

		return
	}

	for i := 0; i < 100; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			m := generateLeaveData()

			if err := m.leave(); err != nil {
				t.Error(err)
			}
		})
	}

}

func Test_findPublic(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		u := users[r.Intn(len(users))]

		mID, err := publicHiddenMoment()
		if err != nil {
			t.Error(err)
		}

		m := new(Moment)
		m.ID = mID
		m.FinderID = u

		if err = m.findPublic(); err != nil {
			t.Error(err)
		}
	})
}

func Test_searchLost(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		u := users[r.Intn(len(users))]

		_, err := searchLost(u)
		if err != nil {
			t.Error(err)
		}
	})
}

func Test_findLeave(t *testing.T) {

	var ms []Moment
	var r *rand.Rand
	var u string
	for {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
		u = users[r.Intn(len(users))]

		msnew, err := searchLost(u)
		if err != nil {
			t.Error(err)
		}
		if len(msnew) > 0 {
			ms = append(ms, msnew...)
			break
		}
	}

	m := ms[r.Intn(len(ms))]
	m.RecipientID = u

	t.Run("1", func(t *testing.T) {

		if err := m.findLeave(); err != nil {
			t.Error(err)
		}
	})
}

func Test_searchFound(t *testing.T) {
	_, u, err := foundLeave()
	if err != nil {
		t.Error(err)
	}

	t.Run("1", func(t *testing.T) {

		_, err := searchFound(u)
		if err != nil {
			t.Error(err)
		}
	})

}

func Test_share(t *testing.T) {
	generateLeaveData := func() (m *Moment) {
		mID, u, err := foundLeave()
		if err != nil {
			t.Error(err)
		}

		m = new(Moment)
		m.RecipientID = u
		m.ID = mID

		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		recipientIDs := make([]string, r.Intn(50)+1)
		for i, _ := range recipientIDs {
			recipientIDs[i] = "User_" + strconv.Itoa(i)
		}
		m.RecipientIDs = recipientIDs

		return
	}

	t.Run("1", func(t *testing.T) {
		m := generateLeaveData()

		if err := m.share(); err != nil {
			t.Error(err)
		}
	})
}

func Test_searchLeft(t *testing.T) {
	sID, err := senderID()
	if err != nil {
		t.Log(err)
	}

	t.Run("1", func(t *testing.T) {

		_, err := searchLeft(sID)
		if err != nil {
			t.Error(err)
		}
	})
}

func Test_searchShared(t *testing.T) {
	u, err := sharedLeave()
	if err != nil {
		t.Error(err)
	}

	t.Run("1", func(t *testing.T) {
		_, err := searchShared(u)
		if err != nil {
			t.Error(err)
		}
	})
}

func Test_searchPublic(t *testing.T) {
	hidden := false
	l, err := publicMomentLocation(hidden)
	if err != nil {
		t.Error(err)
	}

	t.Run("1", func(t *testing.T) {

		_, err := searchPublic(l)
		if err != nil {
			t.Error(err)
		}
	})
}

func Test_searchHiddenPublic(t *testing.T) {
	hidden := true
	l, err := publicMomentLocation(hidden)
	if err != nil {
		t.Error(err)
	}

	t.Run("1", func(t *testing.T) {

		_, err = searchHiddenPublic(l)
		if err != nil {
			t.Error(err)
		}
	})
}

func Test_searchFoundPublic(t *testing.T) {

	u, err := finderID()
	if err != nil {
		t.Error(err)
	}

	t.Run("1", func(t *testing.T) {

		_, err = searchFoundPublic(u)
		if err != nil {
			t.Error(err)
		}
	})
}

func publicHiddenMoment() (id int, err error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT ID
			  FROM moment.Moments
			  WHERE [Public] = 1 
			  		AND [Hidden] = 1`
	rows, err := db.Query(query)
	if err != nil {
		return
	}
	defer db.Close()

	var idS []int
	var idTemp int
	for rows.Next() {
		if err = rows.Scan(&idTemp); err != nil {
			return
		}
		idS = append(idS, idTemp)
	}
	if err = rows.Err(); err != nil {
		return
	}

	rand.Seed(time.Now().UnixNano())
	id = idS[rand.Intn(len(idS))]

	return
}

func publicMomentLocation(hidden bool) (l Location, err error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT Latitude, Longitude
			  FROM [moment].[Moments]
			  WHERE [Public] = 1
			  		AND `
	if hidden {
		query = query + `[Hidden] = 1`
	} else {
		query = query + `[Hidden] = 0`
	}

	rows, err := db.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()

	var idSlice []Location
	var lat, long float32
	for rows.Next() {
		if err = rows.Scan(&lat, &long); err != nil {
			return
		}

		idSlice = append(idSlice, Location{lat, long})
	}
	if err = rows.Err(); err != nil {
		return
	}

	rand.Seed(time.Now().UnixNano())
	l = idSlice[rand.Intn(len(idSlice))]

	return
}

// UserWithFound selects a user that has a found moment.
func foundLeave() (mID int, u string, err error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT ID, MomentID, RecipientID
			  FROM [moment].[Leaves]
			  WHERE Found = 1
			  		AND Shared = 0`
	rows, err := db.Query(query)
	if err != nil {
		return 0, "", err
	}
	defer rows.Close()

	var lID int
	m := make(map[int][]interface{})
	keys := make([]int, 0)
	for rows.Next() {
		if err = rows.Scan(&lID, &mID, &u); err != nil {
			return 0, "", err
		}
		m[lID] = []interface{}{mID, u}
		keys = append(keys, lID)
	}
	rand.Seed(time.Now().UnixNano())
	key := keys[rand.Intn(len(keys))]
	iSlice, ok := m[key]
	if !ok {
		return 0, "", errors.New("Invalid map key.")
	}
	u = iSlice[1].(string)
	mID = iSlice[0].(int)

	return mID, u, nil
}

func sharedLeave() (u string, err error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT RecipientID
			  FROM [moment].[Leaves]
			  WHERE shared = 1`
	rows, err := db.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()

	var rs []string
	var r string
	for rows.Next() {
		if err = rows.Scan(&r); err != nil {
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		return
	}

	rand.Seed(time.Now().UnixNano())
	u = rs[rand.Intn(len(rs))]
	return
}

func senderID() (senderID string, err error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT SenderID
			  FROM [moment].[Moments]`
	rows, err := db.Query(query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var senders []string
	for rows.Next() {
		if err = rows.Scan(&senderID); err != nil {
			return "", err
		}
		senders = append(senders, senderID)
	}
	rand.Seed(time.Now().UnixNano())
	senderID = senders[rand.Intn(len(senders))]

	return senderID, nil
}

func finderID() (finderID string, err error) {
	db := openDbConn()
	defer db.Close()

	query := `SELECT FinderID
			  FROM [moment].[Finds]`
	rows, err := db.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()

	var finders []string
	for rows.Next() {
		if err = rows.Scan(&finderID); err != nil {
			return
		}
		finders = append(finders, finderID)
	}
	if err = rows.Err(); err != nil {
		return
	}

	rand.Seed(time.Now().UnixNano())
	finderID = finders[rand.Intn(len(finders))]

	return
}
