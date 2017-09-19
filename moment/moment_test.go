package moment

import (
	// "fmt"
	// "github.com/stretchr/testify/assert"
	"testing"
	// "testutil"
	"time"
)

const (
	User_1 = "James"
	User_2 = "Sadie"
	User_3 = "Frank"

	Latitude_1  = 43.043978
	Longitude_1 = -87.899151
)

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
		m.RecipientIDs = nil

		if err := m.leave(); err != nil {
			t.Error(err)
		}

		return
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
		m.RecipientIDs = nil

		if err := m.leave(); err != nil {
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
