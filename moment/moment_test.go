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
)

func Test_Moment_leave(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		m := new(Moment)
		m.SenderID = User_1
		m.CreateDate = time.Now().UTC()
		m.Latitude = 43.043978
		m.Longitude = -87.899151
		m.Type = DNE
		m.Message = "Hello World"
		m.MediaLocation = ""
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
		m.Latitude = 43.043978
		m.Longitude = -87.899151
		m.Type = Image
		m.Message = "Hello World"
		m.MediaLocation = "image location"
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
		m.Latitude = 43.043978
		m.Longitude = -87.899151
		m.Type = Video
		m.Message = "Hello World"
		m.MediaLocation = "video location"
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
		m.Latitude = 43.043978
		m.Longitude = -87.899151
		m.Type = DNE
		m.Message = "Hello World"
		m.MediaLocation = ""
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
		m.Latitude = 43.043978
		m.Longitude = -87.899151
		m.Type = Image
		m.Message = "Hello World"
		m.MediaLocation = "image location"
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
