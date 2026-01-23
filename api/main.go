package main

import (
	"embed"
	"flag"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"wingbox.spencrc/lib"
	"wingbox.spencrc/middleware"
)

//go:embed static
var static embed.FS
var fSys fs.FS

func home(w http.ResponseWriter, r *http.Request) {
	var bytes, err = fs.ReadFile(fSys, "index.html")
	if err != nil {
		// We panic here because it's **absolutely required** for index.html to exist
		panic("index.html file doesn't exist!")
	}
	w.Write(bytes)
}

func init() {
	var err error
	fSys, err = fs.Sub(static, "static")
	if err != nil {
		// We panic here because it's **absolutely required** for the file system to be created without issue
		panic(err)
	}
}

func main() {
	// Obtain port from any flags given by the user
	var port uint64
	flag.Uint64Var(&port, "port", 3000, "specify port for server to use")
	flag.Parse()
	// Build address using strconv (fastest method to go from uint -> string)
	var addr string = ":" + strconv.FormatUint(port, 10)

	// Initialize logger
	loggerHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})
	logger := slog.New(loggerHandler)

	// We are using http.NewServeMux() to start up a servemux (router)
	mux := http.NewServeMux()

	// Set up universal middleware!
	baseChain := lib.Chain{
		middleware.LogRequest(logger),
	}

	// Register home function as route (handler) for "/" page
	mux.Handle("/", baseChain.ThenFunc(home))	

	// Print what port the server is listening to
	logger.Info("Starting server", "address", addr)

	// Use http.listenAndServe() to start a new server. We pass the port and the router.
	//  If http.ListenAndServe() returns, then it means it's errored and we log it.
	err := http.ListenAndServe(addr, mux)
	// Functionally same as log.Fatal, but using custom, structured logger
	logger.Error("Stopping server", "err", err)
	os.Exit(1)
}