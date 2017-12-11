package main

import (
	"encoding/json"
	"errors"
	"github.com/penutty/momentservice/moment"
	"log"
	"net/http"
	"time"
)

const (
	MomentEndpoint = "/moment"

	listenPort = ":8081"
)

func main() {
	a := new(app)
	a.c = new(moment.MomentClient)

	mux := http.NewServeMux()

	mux.HandleFunc(MomentEndpoint, a.momentHandler)

	log.Fatal(http.ListenAndServe(listenPort, mux))
}

var (
	ErrorMethodNotImplemented = errors.New("Request method is not implemented by API endpoint.")
	ErrorBadRequest           = errors.New("Request is invalid.")
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
	case http.MethodPatch:
		a.momentPatchHandler(w, r)
	default:
		log.Println(ErrorMethodNotImplemented)
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}
}

func (a *app) momentPostHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	momentType := r.Form.Get("type")
	switch momentType {
	case "private":
		err = a.postPrivateMoment(r)
	case "public":
		err = a.postPublicMoment(r)
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
		err = a.getHiddenMoment(w, r)
	case "lost":
		err = a.getLostMoment(w, r)
	case "sharedbyuser":
		err = a.getSharedMomentbyUser(w, r)
	case "sharedbylocation":
		err = a.getSharedMomentbyLocation(w, r)
	case "found":
		err = a.getFoundMoment(w, r)
	case "left":
		err = a.getLeftMoment(w, r)
	case "public":
		err = a.getPublicMoment(w, r)
	default:
		log.Println(ErrorBadRequest)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err != nil {
		genErrorHandler(w, err)
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
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
	if err != nil {
		genErrorHandler(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *app) postPrivateMoment(r *http.Request) error {
	type medium struct {
		Message string
		Mtype   uint8
	}
	type recipient struct {
		UserID string
	}
	type body struct {
		Latitude   float32
		Longitude  float32
		UserID     string
		Public     bool
		Hidden     bool
		CreateDate time.Time
		Recipients []recipient
		Media      []medium
	}
	b := new(body)
	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		return err
	}

	l := a.c.NewLocation(b.Latitude, b.Longitude)
	m := a.c.NewMomentsRow(l, b.UserID, b.Public, b.Hidden, &b.CreateDate)

	var ms []*moment.MediaRow
	for _, md := range b.Media {
		ms = append(ms, a.c.NewMediaRow(0, md.Message, md.Mtype, ""))
	}

	var fs []*moment.FindsRow
	for _, r := range b.Recipients {
		fs = append(fs, a.c.NewFindsRow(0, r.UserID, false, &time.Time{}))
	}
	if err := a.c.Err(); err != nil {
		return err
	}

	if err := a.c.CreatePrivate(moment.DB(), m, ms, fs); err != nil {
		return err
	}
	return nil
}

func (a *app) postPublicMoment(r *http.Request) error {
	type medium struct {
		Message string
		Mtype   uint8
	}
	type body struct {
		Latitude   float32
		Longitude  float32
		UserID     string
		Public     bool
		Hidden     bool
		CreateDate time.Time
		Media      []medium
	}
	b := new(body)
	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		return err
	}

	l := a.c.NewLocation(b.Latitude, b.Longitude)
	m := a.c.NewMomentsRow(l, b.UserID, b.Public, b.Hidden, &b.CreateDate)

	var ms []*moment.MediaRow
	for _, md := range b.Media {
		ms = append(ms, a.c.NewMediaRow(0, md.Message, md.Mtype, ""))
	}
	if err := a.c.Err(); err != nil {
		return err
	}

	if err := a.c.CreatePublic(moment.DB(), m, ms); err != nil {
		return err
	}
	return nil
}

func (a *app) getHiddenMoment(w http.ResponseWriter, r *http.Request) error {
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

	moments, err := a.c.LocationHidden(moment.DB(), l)
	if err != nil {
		return err
	}

	if err = json.NewEncoder(w).Encode(moments); err != nil {
		return err
	}
	return nil
}

func (a *app) getLostMoment(w http.ResponseWriter, r *http.Request) error {
	type body struct {
		latitude  float32
		longitude float32
		me        string
	}
	b := new(body)
	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		return err
	}

	l := a.c.NewLocation(b.latitude, b.longitude)
	if err := a.c.Err(); err != nil {
		return err
	}

	moments, err := a.c.LocationLost(moment.DB(), l, b.me)
	if err != nil {
		return err
	}

	if err = json.NewEncoder(w).Encode(moments); err != nil {
		return err
	}
	return nil
}

func (a *app) getSharedMomentbyLocation(w http.ResponseWriter, r *http.Request) error {
	type body struct {
		latitude  float32
		longitude float32
		me        string
	}
	b := new(body)
	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		return err
	}

	l := a.c.NewLocation(b.latitude, b.longitude)
	if err := a.c.Err(); err != nil {
		return err
	}

	moments, err := a.c.LocationShared(moment.DB(), l, b.me)
	if err != nil {
		return err
	}

	if err = json.NewEncoder(w).Encode(moments); err != nil {
		return err
	}
	return nil
}

func (a *app) getSharedMomentbyUser(w http.ResponseWriter, r *http.Request) error {
	type body struct {
		you string
		me  string
	}
	b := new(body)
	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		return err
	}

	moments, err := a.c.UserShared(moment.DB(), b.you, b.me)
	if err != nil {
		return err
	}

	if err = json.NewEncoder(w).Encode(moments); err != nil {
		return err
	}
	return nil
}

func (a *app) getFoundMoment(w http.ResponseWriter, r *http.Request) error {
	type body struct {
		me string
	}
	b := new(body)
	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		return err
	}

	moments, err := a.c.UserFound(moment.DB(), b.me)
	if err != nil {
		return err
	}

	if err = json.NewEncoder(w).Encode(moments); err != nil {
		return err
	}
	return nil
}

func (a *app) getLeftMoment(w http.ResponseWriter, r *http.Request) error {
	type body struct {
		me string
	}
	b := new(body)
	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		return err
	}

	moments, err := a.c.UserLeft(moment.DB(), b.me)
	if err != nil {
		return err
	}

	if err = json.NewEncoder(w).Encode(moments); err != nil {
		return err
	}
	return nil
}

func (a *app) getPublicMoment(w http.ResponseWriter, r *http.Request) error {
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

	moments, err := a.c.LocationPublic(moment.DB(), l)
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
	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		return err
	}

	dt := time.Now().UTC()
	f := a.c.NewFindsRow(b.momentID, b.me, true, &dt)
	if err := a.c.Err(); err != nil {
		return err
	}

	if err := a.c.FindPrivate(moment.DB(), f); err != nil {
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
	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		return err
	}

	dt := time.Now().UTC()
	f := a.c.NewFindsRow(b.momentID, b.me, true, &dt)
	if err := a.c.Err(); err != nil {
		return err
	}

	_, err := a.c.FindPublic(moment.DB(), f)
	if err != nil {
		return err
	}
	return nil
}

func (a *app) shareMoment(r *http.Request) error {
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
	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		return err
	}

	s := a.c.NewSharesRow(0, b.momentID, b.userID)
	var rs []*moment.RecipientsRow
	for _, r := range b.recipients {
		rs = append(rs, a.c.NewRecipientsRow(0, r.all, r.recipient))
	}
	if err := a.c.Err(); err != nil {
		return err
	}

	err := a.c.Share(moment.DB(), s, rs)
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
