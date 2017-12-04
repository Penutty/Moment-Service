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
	case "find":
		err = a.findMoment(r)
	case "share":
		err = a.shareMoment(r)
	default:
		log.Println(ErrorBadRequest)
	}
}

func (a *app) postMoment(r *http.Request) {

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

	moments, err := a.c.LocationHidden(moment.MomentDB(), l)
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
}

func (a *app) findMoment(r *http.Request) {

}

func (a *app) shareMoment(r *http.Request) {

}

func genErrorHandler(w http.ResponseWriter, err error) {
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}
