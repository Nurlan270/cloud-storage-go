package main

import (
	"log"

	"go.uber.org/zap"

	application "github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/app"
	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"
)

func main() {
	app := application.New()
	l := logger.Get()

	if err := app.Run(); err != nil {
		l.Error("Cloud storage finished unexpectedly", zap.Error(err))
	}

	//	Close logger
	if err := l.Close(); err != nil {
		log.Fatalf("Failed to close logger: %v", err)
	}
}
