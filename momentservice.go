package main

import (
	"encoding/json"
	"fmt"
	"github.com/penutty/moment"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	MomentEndpoint = "/moment"

	listenPort = ":8081"
)

func main() {
	a := new(app)
	a.c = new(moment.MomentClient)

	mux := http.NewServeMux()

	mux.HandleFunc(MomentEndpoint, momentHandler)

	log.Fatal(http.ListenAndServe(listenPort, mux))
}

var (
	ErrorMethodNotImplemented = errors.New("Request method is not implemented by API endpoint.")
)

type app struct {
	c moment.Client
}

func (a *app) momentHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.momentGetHandler(w, r)
	case http.MethodPost:
		a.momentPostHandler(w, r)
		if err := a.postMoment(r); err != nil {
			genErrorHandler(w, err)
			return
		}
	case http.MethodPatch:
		a.MomentPatchHandler(w, r)
	default:
		log.Println(ErrorBadRequest)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func (a *app) momentPostHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	momentType := r.Form.Get("type")
	switch momentType {
	case "private":
		err = a.postPrivateMoment(w, r)
	case "public":
		err = a.postPublicMoment(w, r)
	default:
		log.Println(ErrorBadRequest)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err != nil {
		genErrorHandler(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (a *app) momentGetHandler(w http.ResponseWriter, r *http.Request) {

	var err error
	momentType := r.Form.Get("type")
	switch momentType {
	case "hidden":
		err = a.getHiddenMoment(r)
	case "lost":
		err = a.getLostMoment(r)
	case "shared":
		err = a.getSharedMoment(r)
	case "found":
		err = a.getFoundMoment(r)
	case "left":
		err = a.getLeftMoment(r)
	case "public":
		err = a.getPublicMoment(r)
	default:
		log.Println(ErrorBadRequest)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err != nil {
		genErrorHander(w, err)
		return
	}
}

func (a *app) momentPatchHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	action := r.Form.Get("action")
	switch action {
	case "findpublic":
		err = a.findPublicMoment(r)
	case "findprivate":
		err = a.findPrivateMoment(r)
	case "share":
		err = a.shareMoment(r)
	default:
		log.Println(ErrorBadRequest)
	}
	if err != nil {
		genErrorHandler(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *app) postPrivateMoment(w http.ResponseWriter, r *http.Request) error {
	type medium struct {
		message string
		mtype   uint8
	}
	type recipient struct {
		userID string
	}
	type body struct {
		latitude   float32
		longitude  float32
		userID     string
		public     bool
		hidden     bool
		createDate time.Time
		recipients []recipient
		media      []medium
	}
	b := new(body)
	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		return err
	}

	l := a.c.NewLocation(b.latitude, b.longitude)
	m := a.c.NewMomentsRow(l, b.userID, b.public, b.hidden, b.createDate)

	var ms []*moment.MediaRow
	for _, md := range b.media {
		ms := append(ms, a.c.NewMediaRow(0, md.message, md.mtype, ""))
	}

	var fs []*moment.FindsRow
	for _, r := range b.recipients {
		fs := append(rs, a.c.NewFindsRow(0, r.userID, false, &time.Time{}))
	}
	if err := a.c.Err(); err != nil {
		return err
	}

	if err := a.c.CreatePrivate(a.c.MomentDB(), m, ms, fs); err != nil {
		return err
	}
	return nil
}

func (a *app) postPublicMoment(w http.ResponseWriter, r *http.Request) error {
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
		createDate time.Time
		recipients []recipient
		media      []medium
	}
	b := new(body)
	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		return err
	}

	l := a.c.NewLocation(b.latitude, b.longitude)
	m := a.c.NewMomentsRow(l, b.userID, b.public, b.hidden, b.createDate)

	var ms []*moment.MediaRow
	for _, md := range b.media {
		ms := append(ms, a.c.NewMediaRow(0, md.message, md.mtype, ""))
	}
	if err := a.c.Err(); err != nil {
		return err
	}

	if err := a.c.CreatePublic(a.c.MomentDB(), m, ms); err != nil {
		return err
	}
	return nil
}

func (a *app) getHiddenMoment(r *http.Request) error {
	type body struct {
		latitude  float32
		longitude float32
	}
	b := new(body)
	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		return err
	}

	l := a.c.NewLocation(b.latitude, b.longitude)
	if err := a.c.Err(); err != nil {
		return err
	}

	moments, err := a.c.LocationHidden(a.c.MomentDB(), l)
	if err != nil {
		return err
	}

	if err = json.NewEncoder(w).Encode(moments); err != nil {
		return err
	}
}

func (a *app) getLostMoment(r *http.Request) error {
	type body struct {
		latitude  float32
		longitude float32
		me        string
	}
	b := new(body)
	if err := json.NewDecoder(r.body).Decode(b); err != nil {
		return err
	}

	l := a.c.NewLocation(b.latitude, b.longitude)
	if err := a.c.Err(); err != nil {
		return err
	}

	moments, err := a.c.LocationLost(moment.MomentDB(), l)
	if err != nil {
		return err
	}

	if err = json.NewEncoder(w).Encode(moments); err != nil {
		return err
	}
}

func (a *app) getSharedMomentbyLocation(r *http.Request) error {
	type body struct {
		latitude  float32
		longitude float32
	}
	b := new(body)
	if err := json.NewDecoder(r.body).Decode(b); err != nil {
		return err
	}

	l := a.c.NewLocation(b.latitude, b.longitude)
	if err := a.c.Err(); err != nil {
		return err
	}

	moments, err := a.c.LocationShared(moment.MomentDB(), l)
	if err != nil {
		return err
	}

	if err = json.NewEncoder(w).Encode(moments); err != nil {
		return err
	}
}

func (a *app) getSharedMomentbyUser(r *http.Request) error {
	type body struct {
		you string
		me  string
	}
	b := new(body)
	if err := json.NewDecoder(r.body).Decode(b); err != nil {
		return err
	}

	moments, err := a.c.UserShared(moment.MomentDB(), b.you, b.me)
	if err != nil {
		return err
	}

	if err = json.NewEncoder(w).Encode(moments); err != nil {
		return err
	}
}

func (a *app) getFoundMoment(r *http.Request) {
	type body struct {
		me string
	}
	b := new(body)
	if err := json.NewDecoder(r.body).Decode(b); err != nil {
		return err
	}

	moments, err := a.c.UserFound(moment.MomentDB(), me)
	if err != nil {
		return err
	}

	if err = json.NewEncoder(w).Encode(moments); err != nil {
		return err
	}
}

func (a *app) getLeftMoment(r *http.Request) error {
	type body struct {
		me string
	}
	b := new(body)
	if err := json.NewDecoder(r.body).Decode(b); err != nil {
		return err
	}

	moments, err := a.c.UserLeft(moment.MomentDB(), me)
	if err != nil {
		return err
	}

	if err = json.NewEncoder(w).Encode(moments); err != nil {
		return err
	}
	return nil
}

func (a *app) getPublicMoment(r *http.Request) error {
	type body struct {
		latitude  float32
		longitude float32
	}
	b := new(body)
	if err := json.NewDecoder(r.body).Decode(b); err != nil {
		return err
	}

	l := a.c.NewLocation(b.latitude, b.longitude)
	if err := a.c.Err(); err != nil {
		return err
	}

	moments, err := a.c.LocationPublic(moment.MomentDB(), l)
	if err != nil {
		return err
	}

	if err = json.NewEncoder(w).Encode(moments); err != nil {
		return err
	}
	return nil
}

func (a *app) findPrivateMoment(r *http.Request) error {
	type body struct {
		momentID int64
		me       string
	}
	b := new(body)
	if err := json.NewDecoder(r.body).Decode(b); err != nil {
		return err
	}

	dt := time.Now().UTC()
	f := a.c.NewFindsRow(b.momentID, b.me, true, &dt)
	if err := a.c.Err(); err != nil {
		return err
	}

	if err := a.c.FindPrivate(moment.MomentDB(), f); err != nil {
		return err
	}
	return nil
}

func (a *app) findPublicMoment(r *http.Request) error {
	type body struct {
		momentID int64
		me       string
	}
	b := new(body)
	if err := json.NewDecoder(r.body).Decode(b); err != nil {
		return err
	}

	dt := time.Now().UTC()
	f := a.c.NewFindsRow(b.momentID, b.me, true, &dt)
	if err := a.c.Err(); err != nil {
		return err
	}

	_, err := a.c.FindPublic(moment.MomentDB(), f)
	if err != nil {
		return err
	}
	return nil
}

func (a *app) shareMoment(r *http.Request) {
	type recipient struct {
		all       bool
		recipient string
	}
	type body struct {
		momentID   int64
		userID     string
		recipients []recipient
	}
	b := new(body)
	if err := json.NewDecoder(r.body).Decode(b); err != nil {
		return err
	}

	s := a.c.NewSharesRow(0, b.momentID, b.userID)
	var rs []*moment.RecipientsRow
	for _, r := range s.recipients {
		rs = append(rs, a.c.NewSharesRow(0, r.all, r.recipient))
	}
	if err := a.c.Err(); err != nil {
		return err
	}

	err := a.c.Share(moment.MomentDB(), s, rs)
	if err != nil {
		return err
	}
	return nil
}

func genErrorHandler(w http.ResponseWriter, err error) {
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}
