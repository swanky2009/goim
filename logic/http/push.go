package http

import (
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	xstrings "github.com/swanky2009/goim/pkg/strings"
)

func (s *Server) pushKeys(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	opStr := query.Get("op")
	keysStr := query.Get("keys")
	// read message
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	op, err := strconv.ParseInt(opStr, 10, 32)
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	if err = s.logic.PushKeys(context.TODO(), int32(op), strings.Split(keysStr, ","), msg); err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	writeJSON(w, OK, nil)
}

func (s *Server) pushMids(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	opStr := query.Get("op")
	midsStr := query.Get("mids")
	// read message
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	op, err := strconv.ParseInt(opStr, 10, 32)
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	mids, err := xstrings.SplitInt64s(midsStr, ",")
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	if err = s.logic.PushMids(context.TODO(), int32(op), mids, msg); err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	writeJSON(w, OK, nil)
}

func (s *Server) pushRoom(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	opStr := query.Get("op")
	room := query.Get("room")
	// read message
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	op, err := strconv.ParseInt(opStr, 10, 32)
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	if err = s.logic.PushRoom(context.TODO(), int32(op), room, msg); err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	writeJSON(w, OK, nil)
}

func (s *Server) pushAll(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	opStr := query.Get("op")
	speedStr := query.Get("speed")
	platStr := query.Get("plat")
	// read message
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, RequestErr, err)
		return
	}
	op, err := strconv.ParseInt(opStr, 10, 32)
	if err != nil {
		writeJSON(w, RequestErr, err)
		return
	}
	speed, err := strconv.ParseInt(speedStr, 10, 32)
	if err != nil {
		writeJSON(w, RequestErr, err)
		return
	}
	if err = s.logic.PushAll(context.TODO(), int32(op), int32(speed), platStr, msg); err != nil {
		writeJSON(w, RequestErr, err)
		return
	}
	writeJSON(w, OK, nil)
}
