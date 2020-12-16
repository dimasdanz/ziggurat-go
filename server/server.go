package server

import (
	"github.com/gojekfarm/ziggurat"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

var defaultHTTPPort = "8080"

type DefaultHttpServer struct {
	server *http.Server
	router *httprouter.Router
}

func WithPort(port string) func(s *DefaultHttpServer) {
	return func(s *DefaultHttpServer) {
		s.server.Addr = "localhost:" + port
	}
}

func New(opts ...func(s *DefaultHttpServer)) *DefaultHttpServer {
	router := httprouter.New()
	server := &http.Server{Handler: requestLogger(router)}
	s := &DefaultHttpServer{
		server: server,
		router: router,
	}
	for _, opt := range opts {
		opt(s)
	}
	s.server.Addr = "localhost:" + defaultHTTPPort
	return s
}

func (s *DefaultHttpServer) Start(app ziggurat.App) error {
	s.router.GET("/v1/ping", pingHandler)
	go func(server *http.Server) {
		if err := server.ListenAndServe(); err != nil {
			ziggurat.LogError(err, "default http server error", nil)
		}
	}(s.server)
	ziggurat.LogInfo("http server start on "+s.server.Addr, nil)
	return nil
}

func (s *DefaultHttpServer) ConfigureHTTPEndpoints(f func(r *httprouter.Router)) {
	f(s.router)
}

func (s *DefaultHttpServer) ConfigureHandler(f func(r *httprouter.Router) http.Handler) {
	s.server.Handler = f(s.router)
}

func (s *DefaultHttpServer) Stop(app ziggurat.App) {
	ziggurat.LogError(s.server.Shutdown(app.Context()), "default http server: stopping http server", nil)
}