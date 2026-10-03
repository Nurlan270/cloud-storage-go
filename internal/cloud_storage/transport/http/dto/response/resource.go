package response

import "github.com/Nurlan270/cloud-storage-go/internal/core/models"

type ResourceInfo struct {
	Path string              `json:"path"`
	Name string              `json:"name"`
	Size *int64              `json:"size,omitempty"`
	Type models.ResourceType `json:"type"`
}

func NewResourceInfoFromModel(resource models.Resource) ResourceInfo {
	info := ResourceInfo{
		Path: resource.Path,
		Name: resource.Name,
		Type: resource.Type,
	}

	if !resource.IsDir() {
		info.Size = &resource.Size
	}

	return info
}

func NewResourceInfoListFromModels(resources []models.Resource) []ResourceInfo {
	list := make([]ResourceInfo, 0, len(resources))

	for _, resource := range resources {
		list = append(list, NewResourceInfoFromModel(resource))
	}

	return list
}
