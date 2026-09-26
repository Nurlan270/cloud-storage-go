package archiver

import (
	"context"
	"io"
	"os"

	"github.com/mholt/archives"
	"go.uber.org/zap"

	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"
)

type ArchiveType string

const (
	TypeZip ArchiveType = ".zip"
)

const (
	// ArchiveDir is the root directory for all temporary archives.
	ArchiveDir = "tmp/archive"

	// ArchiveFilesDir is the root directory for all temporary archive files.
	ArchiveFilesDir = ArchiveDir + "/files"
)

type ArchiveResult struct {
	Content io.ReadCloser
	Name    string
	Path    string
	Size    int64
}

type ArchiveFile struct {
	PathOnDisk string
	Filename   string
}

// Archive archives all provided files under zip archive.
func Archive(ctx context.Context, archiveName string, files []ArchiveFile) (ArchiveResult, error) {
	log := logger.Get().With(
		zap.String("src", "archiver"))

	filenames := make(map[string]string, len(files))
	for _, file := range files {
		filenames[file.PathOnDisk] = file.Filename
	}

	//	Remove all opened files after putting them into archive
	defer func() {
		for _, file := range files {
			if err := os.Remove(file.PathOnDisk); err != nil {
				log.Warn("failed to remove file",
					zap.String("path", file.PathOnDisk), zap.Error(err))
			}
		}
	}()

	diskFiles, err := archives.FilesFromDisk(ctx, nil, filenames)
	if err != nil {
		log.Error("failed to get files from disk", zap.Error(err))

		return ArchiveResult{}, err
	}

	archiveName = buildArchiveName(archiveName, TypeZip)
	archivePath := buildArchivePath(archiveName)

	//	Create the output file we'll write to
	out, err := os.Create(archivePath)
	if err != nil {
		log.Error("failed to create archive file", zap.Error(err))

		return ArchiveResult{}, err
	}

	var format archives.Zip

	//	Archive files
	if err = format.Archive(ctx, out, diskFiles); err != nil {
		log.Error("failed to archive files", zap.Error(err))

		out.Close()

		return ArchiveResult{}, err
	}

	//	Get archive file's info
	info, err := out.Stat()
	if err != nil {
		log.Error("failed to get stat", zap.Error(err))

		out.Close()

		return ArchiveResult{}, err
	}

	//	Reset read position
	if _, err = out.Seek(0, io.SeekStart); err != nil {
		log.Error("failed to reset read position", zap.Error(err))

		out.Close()

		return ArchiveResult{}, err
	}

	return ArchiveResult{
		Content: out,
		Name:    archiveName,
		Path:    archivePath,
		Size:    info.Size(),
	}, nil
}

func buildArchivePath(filename string) string {
	return ArchiveDir + "/" + filename
}

func buildArchiveName(filename string, archiveType ArchiveType) string {
	return filename + string(archiveType)
}
