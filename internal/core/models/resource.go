package models

import (
	"fmt"
	"path"
	"strings"
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
	Size   int64
	Type   ResourceType
}

func (r Resource) IsDir() bool {
	return r.Type == TypeDir
}

// ObjectKey returns key of resource within bucket.
func (r Resource) ObjectKey() string {
	var rootDir = fmt.Sprintf("user-%d-files", r.UserID)

	key := path.Join(rootDir, r.Path, r.Name)

	if r.IsDir() {
		//	Add trailing slash so that MinIO will recognize it as directory
		key += "/"
	}

	return key
}

// FullPath returns full path to resource (Path + Name).
func (r Resource) FullPath() string {
	str := strings.TrimLeft(path.Join(r.Path, r.Name), "/")

	if r.IsDir() {
		str += "/"
	}

	return str
}
