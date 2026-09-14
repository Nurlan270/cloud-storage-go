package response

import "github.com/Nurlan270/cloud-storage-go/internal/core/models"

type ResourceInfoList = []*ResourceInfo

type ResourceInfo struct {
	Path string              `json:"path"`
	Name string              `json:"name"`
	Size *int64              `json:"size,omitempty"`
	Type models.ResourceType `json:"type"`
}
