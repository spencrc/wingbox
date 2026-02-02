package main

import (
	"wingbox.spencrc/internal/auth"
)

func main() {
	as := auth.NewAuthService()
	as.RegisterRoutes()
	as.Listen(3002)
}
