package server

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/lmittmann/tint"
	"wingbox.spencrc/internal/shared"
	"wingbox.spencrc/internal/shared/middleware"
)

type Server struct {
	Logger *slog.Logger
	mux *http.ServeMux
	BaseChain shared.Chain
}

// Creates Logger, creates ServeMux, and creates universal middleware chain. These values are then used to create a Server struct.
func Init() *Server {
	// Initialize logger
	loggerHandler := tint.NewHandler(os.Stderr, &tint.Options{})
	logger := slog.New(loggerHandler)

	// We are using http.NewServeMux() to start up a servemux (router)
	mux := http.NewServeMux()

	// Set up universal middleware!
	baseChain := shared.Chain{
		middleware.LogRequest(logger),
	}

	return &Server{logger, mux, baseChain}
}

// Wrapper for ServeMux.Handle
func (s *Server) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

// Begins listening on server's ServeMux at port specified in Init. Logs and exits on error.
func (s *Server) Listen(port uint64) {
	addr := ":" + strconv.FormatUint(port, 10)
	s.Logger.Info("Starting server", "address", addr)

	// Use http.listenAndServe() to start a new server. We pass the port and the router.
	//  If http.ListenAndServe() returns, then it means it's errored and we log it.
	err := http.ListenAndServe(addr, s.mux)
	// Functionally same as log.Fatal, but using custom, structured logger
	s.Logger.Error("Stopping server", "err", err)
	os.Exit(1)
}