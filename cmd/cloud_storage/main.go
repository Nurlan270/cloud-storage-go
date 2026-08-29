package main

import (
	"errors"
	"log"
	"net/http"

	application "github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/app"
)

func main() {
	app := application.New()

	if err := app.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Cloud storage finished unexpectedly: %s", err)
	}
}
