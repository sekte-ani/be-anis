package repository

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var allowedMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

type StorageRepository interface {
	SaveImage(file *multipart.FileHeader) (publicURL string, err error)
}

type storageRepository struct {
	uploadDir  string
	appBaseURL string
}

func NewStorageRepository(uploadDir, appBaseURL string) StorageRepository {
	return &storageRepository{
		uploadDir:  uploadDir,
		appBaseURL: strings.TrimRight(appBaseURL, "/"),
	}
}

func (r *storageRepository) SaveImage(file *multipart.FileHeader) (string, error) {
	contentType := file.Header.Get("Content-Type")
	ext, ok := allowedMIME[contentType]
	if !ok {
		// Fallback: infer from original filename
		orig := strings.ToLower(filepath.Ext(file.Filename))
		for _, v := range allowedMIME {
			if v == orig {
				ext = orig
				ok = true
				break
			}
		}
		if !ok {
			return "", fmt.Errorf("unsupported file type: %s", contentType)
		}
	}

	if err := os.MkdirAll(r.uploadDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	filename := uuid.NewString() + ext
	dest := filepath.Join(r.uploadDir, filename)

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	out, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	buf := make([]byte, 32*1024)
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return "", fmt.Errorf("failed to write file: %w", writeErr)
			}
		}
		if readErr != nil {
			break
		}
	}

	publicURL := fmt.Sprintf("%s/img/%s", r.appBaseURL, filename)
	return publicURL, nil
}
