package http

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) onlineTop(w http.ResponseWriter, r *http.Request) {
	var (
		limit int64
		err   error
		res   []string
	)
	query := r.URL.Query()
	typeStr := query.Get("type")
	limitStr := query.Get("limit")
	limit, err = strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		limit = 2
	}
	res, err = s.logic.OnlineTop(context.TODO(), typeStr, int64(limit))
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	writeJSON(w, OK, res)
}

func (s *Server) onlineRoom(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	roomStr := query.Get("room")
	res, err := s.logic.OnlineRoom(context.TODO(), strings.Split(roomStr, ","))
	if err != nil {
		writeJSON(w, RequestErr, nil)
		return
	}
	writeJSON(w, OK, res)
}
