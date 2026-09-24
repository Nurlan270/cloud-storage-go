package minio

import (
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func connect(conf Config) (*minio.Client, error) {
	client, err := minio.New(conf.GetAddr(), &minio.Options{
		Creds: credentials.NewStaticV4(conf.User, conf.Password, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("minio: failed to connect: %w", err)
	}

	return client, nil
}
