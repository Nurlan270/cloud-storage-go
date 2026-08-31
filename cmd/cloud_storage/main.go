package main

import (
	"log"

	application "github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/app"
)

func main() {
	app := application.New()

	if err := app.Run(); err != nil {
		log.Fatalf("Cloud storage finished unexpectedly: %s", err)
	}
}
