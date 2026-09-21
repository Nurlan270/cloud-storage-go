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
	UserID uint64       `json:"-"`
	Path   string       `json:"path"`
	Name   string       `json:"name"`
	Size   *int64       `json:"size,omitempty"`
	Type   ResourceType `json:"type"`
}

func (r Resource) IsDir() bool {
	return r.Type == TypeDir
}

func (r Resource) ObjectKey() string {
	var rootDir = fmt.Sprintf("user-%d-files", r.UserID)

	key := path.Join(rootDir, r.Path, r.Name)

	if r.IsDir() {
		//	Add trailing slash so that MinIO will recognize it as directory
		key += "/"
	}

	return key
}

func (r Resource) FullPath() string {
	str := strings.TrimLeft(path.Join(r.Path, r.Name), "/")

	if r.IsDir() {
		str += "/"
	}

	return str
}
