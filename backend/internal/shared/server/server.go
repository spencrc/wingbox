package server

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/lmittmann/tint"
	_ "modernc.org/sqlite"
	"wingbox.spencrc/internal/shared"
	"wingbox.spencrc/internal/shared/middleware"
)

type Server struct {
	Logger *slog.Logger
	mux *http.ServeMux
	Db *sql.DB
	BaseChain shared.Chain
}

// Creates Logger, creates ServeMux, and creates universal middleware chain. These values are then used to create a Server struct.
func Init() *Server {
	// Initialize logger
	loggerHandler := tint.NewHandler(os.Stderr, &tint.Options{})
	logger := slog.New(loggerHandler)

	// We are using http.NewServeMux() to start up a servemux (router)
	mux := http.NewServeMux()

	// Set up the database!
	const DB_PATH = "file:///db/app.db?_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite", DB_PATH)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	// Set up universal middleware!
	baseChain := shared.Chain{
		middleware.LogRequest(logger),
	}

	return &Server{logger, mux, db, baseChain}
}

// Wrapper for ServeMux.Handle
func (s *Server) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

// Effectively same as log.Fatal, but using structured logger instead
func (s *Server) LogFatal(msg string, args ...any) {
	s.Logger.Error(msg, args...)
	os.Exit(1)
}

// Begins listening on server's ServeMux at port specified in Init. Logs and exits on error.
func (s *Server) Listen(port uint64) {
	addr := ":" + strconv.FormatUint(port, 10)
	s.Logger.Info("Starting server", "address", addr)

	// Use http.listenAndServe() to start a new server. We pass the port and the router.
	//  If http.ListenAndServe() returns, then it means it's errored and we log it.
	err := http.ListenAndServe(addr, s.mux)
	// We can only get here if something's gone wrong!
	s.Db.Close()
	// Functionally same as log.Fatal, but using custom, structured logger
	s.LogFatal("Stopping server", "err", err)
}