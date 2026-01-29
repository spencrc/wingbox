package main

import (
	"net/http"

	"wingbox.spencrc/internal/shared/server"
)

func home(w http.ResponseWriter, r *http.Request) {
	
}

func main() {
	s := server.Init()

	// Register home function as route (handler) for "/" page
	s.Handle("/", s.BaseChain.ThenFunc(home))	

	s.Listen(3001)
}