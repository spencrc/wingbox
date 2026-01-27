package server

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

type Server struct {
	addr string;
	Logger *slog.Logger
	mux *http.ServeMux
}

func Init() *Server {
	// Obtain port from any flags given by the user
	var port uint64
	flag.Uint64Var(&port, "port", 3000, "specify port for server to use")
	flag.Parse()
	// Build address using strconv (fastest method to go from uint -> string)
	addr := ":" + strconv.FormatUint(port, 10)

	// Initialize logger
	loggerHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})
	logger := slog.New(loggerHandler)

	// We are using http.NewServeMux() to start up a servemux (router)
	mux := http.NewServeMux()

	return &Server{addr, logger, mux}
}

func (s *Server) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

func (s *Server) Listen() {
	s.Logger.Info("Starting server", "address", s.addr)

	// Use http.listenAndServe() to start a new server. We pass the port and the router.
	//  If http.ListenAndServe() returns, then it means it's errored and we log it.
	err := http.ListenAndServe(s.addr, s.mux)
	// Functionally same as log.Fatal, but using custom, structured logger
	s.Logger.Error("Stopping server", "err", err)
	os.Exit(1)
}