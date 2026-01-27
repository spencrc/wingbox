package main

import (
	"net/http"

	"wingbox.spencrc/internal/shared"
	"wingbox.spencrc/internal/shared/middleware"
	"wingbox.spencrc/internal/shared/server"
)

func home(w http.ResponseWriter, r *http.Request) {
	
}

func main() {
	s := server.Init()

	// Set up universal middleware!
	baseChain := shared.Chain{
		middleware.LogRequest(s.Logger),
	}

	// Register home function as route (handler) for "/" page
	s.Handle("/", baseChain.ThenFunc(home))	

	s.Listen()
}