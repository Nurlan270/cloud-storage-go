package main

import (
	"log"

	application "github.com/Nurlan270/cloud-storage-go/internal/auth_server/app"
)

func main() {
	app := application.New()

	if err := app.Run(); err != nil {
		log.Fatalf("Auth server finished unexpectedly: %s", err)
	}
}
