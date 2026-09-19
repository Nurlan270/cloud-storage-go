package models

import (
	"fmt"
	"path"
)

type ResourceType string

const (
	TypeFile ResourceType = "FILE"
	TypeDir  ResourceType = "DIRECTORY"
)

type Resource struct {
	UserID uint64
	Path   string
	Name   string
	Size   *int64
	Type   ResourceType
}

func (r Resource) IsDir() bool {
	return r.Type == TypeDir
}

func (r Resource) ObjectKey() string {
	var rootDir = fmt.Sprintf("user-%d-files", r.UserID)

	return path.Join(rootDir, r.Path, r.Name)
}

func (r Resource) FullPath() string {
	return path.Join(r.Path, r.Name)
}
