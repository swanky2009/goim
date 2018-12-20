package http

import (
	"net/http"
	"time"

	"github.com/swanky2009/goim/logic"
	"github.com/swanky2009/goim/logic/g"
	"github.com/swanky2009/goim/logic/g/conf"
)

// Server is http server.
type Server struct {
	logic *logic.Server
}

// New new a http server.
func New(c *conf.HTTPServer, l *logic.Server) *http.Server {
	s := &Server{
		logic: l,
	}
	srv := &http.Server{
		Addr:           c.Addr,
		Handler:        s.newHTTPServeMux(),
		ReadTimeout:    time.Duration(c.ReadTimeout),
		WriteTimeout:   time.Duration(c.WriteTimeout),
		MaxHeaderBytes: 1 << 20,
	}
	return srv
}

func Start(s *http.Server, errc chan error) {
	go func() {
		errc <- s.ListenAndServe()
	}()

	g.Logger.Infof("start http server listen: %s", s.Addr)

	return
}

func (s *Server) newHTTPServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/push/keys", s.pushKeys)
	mux.HandleFunc("/push/mids", s.pushMids)
	mux.HandleFunc("/push/room", s.pushRoom)
	mux.HandleFunc("/push/all", s.pushAll)
	mux.HandleFunc("/online/top", s.onlineTop)
	mux.HandleFunc("/online/room", s.onlineRoom)
	mux.HandleFunc("/online/total", s.onlineTotal)
	return mux
}
