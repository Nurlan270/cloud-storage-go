package service

import (
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"strings"

	"github.com/Nurlan270/cloud-storage-go/internal/core/models"

	gopath "path"
)

// getFullPath returns full path of provided resource using its Content-Disposition header.
func getFullPath(resource *multipart.FileHeader, rootDir string) (string, error) {
	header := resource.Header.Get("Content-Disposition")
	if header == "" {
		return "", errors.New("missing Content-Disposition")
	}

	_, params, err := mime.ParseMediaType(header)
	if err != nil {
		return "", fmt.Errorf("invalid Content-Disposition: %w", err)
	}

	filename, ok := params["filename"]
	if !ok || filename == "" {
		return "", errors.New("missing filename")
	}

	return gopath.Join(rootDir, filename), nil
}

// buildResourcesFromPath recursively loops throw all path elements
// and creates models.Resource for each element.
func buildResourcesFromPath(userID uint64, path string, size *int64) []models.Resource {
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")

	resources := make([]models.Resource, 0, len(parts))

	dir := "/"

	// All components except the final one are directories.
	for i := 0; i < len(parts)-1; i++ {
		//	If this is not first element of path, then join it with all previous
		//	elements using "/" and add trailing slash because this is a directory
		if i > 0 {
			dir = strings.Join(parts[:i], "/") + "/"
		}

		resources = append(resources, models.Resource{
			UserID: userID,
			Path:   dir,
			Name:   parts[i],
			Type:   models.TypeDir,
		})
	}

	if len(parts) > 1 {
		dir = strings.Join(parts[:len(parts)-1], "/") + "/"
	}

	//	Determine last resource's type
	resourceType := getResourceType(parts[len(parts)-1])

	//	Set last element of slice
	resources = append(resources, models.Resource{
		UserID: userID,
		Path:   dir,
		Name:   parts[len(parts)-1],
		Size:   size,
		Type:   resourceType,
	})

	return resources
}

// getResourceType determines passed element's type.
func getResourceType(el string) models.ResourceType {
	if strings.HasSuffix(el, "/") {
		return models.TypeDir
	}

	return models.TypeFile
}

// splitPath splits provided path returning path to element and last element.
func splitPath(path string) (string, string) {
	path = strings.Trim(path, "/")

	dir, name := gopath.Split(path)
	if dir == "" {
		dir = "/"
	}

	return dir, name
}
