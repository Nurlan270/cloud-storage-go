package minio

import (
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func MustConnect(conf Config) *minio.Client {
	client, err := connect(conf)
	if err != nil {
		panic(err)
	}

	return client
}

func Connect(conf Config) (*minio.Client, error) {
	return connect(conf)
}

func connect(conf Config) (*minio.Client, error) {
	client, err := minio.New("minio:9000", &minio.Options{
		Creds: credentials.NewStaticV4(conf.User, conf.Password, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("minio: failed to connect: %w", err)
	}

	return client, nil
}
