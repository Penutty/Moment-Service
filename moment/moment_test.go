package moment

import (
	"fmt"
	// "github.com/stretchr/testify/assert"
	"os"
	"testing"
	"testutil"
	"time"
)

const (
	User_1 = "James"
	User_2 = "Sadie"
	User_3 = "Frank"

	Latitude_1  = 43.043978
	Longitude_1 = -87.899151
)

func TestMain(m *testing.M) {
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

	t.Run("1", func(t *testing.T) {
		m := new(Moment)
		m.SenderID = User_1
		m.CreateDate = time.Now().UTC()
		m.Latitude = Latitude_1
		m.Longitude = Longitude_1
		m.Type = DNE
		m.Message = "Hello World"
		m.MediaDir = ""
		m.Public = true
		m.Shared = false
		m.RecipientIDs = nil

		if err := m.leave(); err != nil {
			t.Error(err)
		}

	})

	t.Run("2", func(t *testing.T) {
		m := new(Moment)
		m.SenderID = User_1
		m.CreateDate = time.Now().UTC()
		m.Latitude = Latitude_1
		m.Longitude = Longitude_1
		m.Type = Image
		m.Message = "Hello World"
		m.MediaDir = "image location"
		m.Public = true
		m.Shared = false
		m.RecipientIDs = nil

		if err := m.leave(); err != nil {
			t.Error(err)
		}

	})

	t.Run("3", func(t *testing.T) {
		m := new(Moment)
		m.SenderID = User_1
		m.CreateDate = time.Now().UTC()
		m.Latitude = Latitude_1
		m.Longitude = Longitude_1
		m.Type = Video
		m.Message = "Hello World"
		m.MediaDir = "video location"
		m.Public = true
		m.Shared = false
		m.RecipientIDs = nil

		if err := m.leave(); err != nil {
			t.Error(err)
		}

	})

	t.Run("4", func(t *testing.T) {
		m := new(Moment)
		m.SenderID = User_1
		m.CreateDate = time.Now().UTC()
		m.Latitude = Latitude_1
		m.Longitude = Longitude_1
		m.Type = DNE
		m.Message = "Hello World"
		m.MediaDir = ""
		m.Public = false
		m.Shared = false
		m.RecipientIDs = []string{User_2, User_3}

		if err := m.leave(); err != nil {
			t.Error(err)
		}

	})

	t.Run("5", func(t *testing.T) {
		m := new(Moment)
		m.SenderID = User_1
		m.CreateDate = time.Now().UTC()
		m.Latitude = Latitude_1
		m.Longitude = Longitude_1
		m.Type = Image
		m.Message = "Hello World"
		m.MediaDir = "image location"
		m.Public = true
		m.Shared = false
		m.RecipientIDs = nil

		if err := m.leave(); err != nil {
			t.Error(err)
		}
	})
}

func Test_find(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		m := new(Moment)
		m.ID = 1
		m.RecipientIDs = []string{User_2}

		if err := m.find(); err != nil {
			t.Error(err)
		}
	})
}

func Test_share(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		m := new(Moment)
		m.RecipientID = User_2
		m.ID = 4
		m.RecipientIDs = []string{User_1, User_3}

		if err := m.share(); err != nil {
			t.Error(err)
		}
	})
}

func Test_searchLeft(t *testing.T) {
	t.Run("1", func(t *testing.T) {

		_, err := searchLeft(User_1)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("2", func(t *testing.T) {

		_, err := searchLeft(User_2)
		if err != nil {
			t.Error(err)
		}
	})
}

func Test_searchLost(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		_, err := searchLost(User_1)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("2", func(t *testing.T) {
		_, err := searchLost(User_2)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("3", func(t *testing.T) {
		_, err := searchLost(User_3)
		if err != nil {
			t.Error(err)
		}
	})
}

func Test_searchFound(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		_, err := searchFound(User_1)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("2", func(t *testing.T) {
		_, err := searchFound(User_2)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("3", func(t *testing.T) {
		_, err := searchFound(User_3)
		if err != nil {
			t.Error(err)
		}
	})
}

func Test_searchShared(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		_, err := searchShared(User_1)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("2", func(t *testing.T) {
		_, err := searchShared(User_2)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("3", func(t *testing.T) {
		_, err := searchShared(User_3)
		if err != nil {
			t.Error(err)
		}
	})
}

func Test_searchPublic(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		l := Location{
			Latitude:  Latitude_1,
			Longitude: Longitude_1,
		}
		_, err := searchPublic(l)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("2", func(t *testing.T) {
		l := Location{
			Latitude:  Latitude_1,
			Longitude: Longitude_1,
		}
		_, err := searchPublic(l)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("3", func(t *testing.T) {
		l := Location{
			Latitude:  Latitude_1,
			Longitude: Longitude_1,
		}
		_, err := searchPublic(l)
		if err != nil {
			t.Error(err)
		}
	})
}
