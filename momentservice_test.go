package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"os"
	"testing"
	"testutil"
)

func TestMain(m *testing.M) {
	// Create http.Client or http.Transport if necessary

	os.Exit(m.Run())
}

func Test_postPrivateMoment(t *testing.T) {
	type recipient struct {
		userID string
	}
	type medium struct {
		message string
		mtype   uint8
	}
	type body struct {
		latitude   float32
		longitude  float32
		userID     string
		public     bool
		hidden     bool
		creatDate  *time.Time
		recipients []recipient
		media      []medium
	}

}

type MockClient struct {
	err error
}

func (mc *MockClient) Err() error {

}

func (mc *MockClient) FindPublic(db DbRunner, f *moment.FindsRow) (int64, error) {

}

func (mc *MockClient) FindPrivate(db DbRunner, f *moment.FindsRow) error {

}

func (mc *MockClient) Share(db DbRunnerTrans, m *moment.MomentsRow, ms []*MediaRow) error {

}

func (mc *MockClient) CreatePublic(db DbRunnerTrans, m *MomentsRow, ms []*MediaRow) error {

}

func (mc *MockClient) CreatePrivate(db DbRunnerTrans, m *MomentsRow, ms []*MediaRow, fs []*FindsRow) error {

}

func (mc *MockClient) NewMomentsRow(l *Location, userID string, public bool, hidden bool, createDate *time.Time) *MomentsRow {

}

func (mc *MockClient) NewLocation(lat string, long string) *Location {

}

func (mc *MockClient) NewMediaRow(momentID int64, userID string, mtype uint8, dir string) *MediaRow {

}

func (mc *MockClient) NewFindsRow(momentID int64, userID string, found bool, findDate *time.Time) *FindsRow {

}

func (mc *MockClient) NewSharesRow(sharesID int64, momentID int64, userID string) *SharesRow {

}

func (mc *MockClient) NewRecipientsRow(sharesID int64, all bool, recipientID string) *RecipientsRow {

}

func (mc *MockClient) LocationShared(db DbRunner, l *Location, me string) ([]*Moment, error) {

}

func (mc *MockClient) LocationPublic(db DbRunner, l *Location) ([]*Moment, error) {

}

func (mc *MockClient) LocationHidden(db DbRunner, l *Location) ([]*Moment, error) {

}

func (mc *MockClient) LocationLost(db DbRunner, l *Location) ([]*Moment, error) {

}

func (mc *MockClient) UserShared(db DbRunner, you string, me string) ([]*Moment, error) {

}

func (mc *MockClient) UserLeft(db DbRunner, me string) ([]*Moment, error) {

}

func (mc *MockClient) UserFound(db DbRunner, me string) ([]*Moment, error) {

}
