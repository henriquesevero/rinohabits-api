package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	uploadsDir string
	baseURL    string
}

func NewLocalStorage(uploadsDir, baseURL string) *LocalStorage {
	return &LocalStorage{uploadsDir: uploadsDir, baseURL: baseURL}
}

func (s *LocalStorage) Upload(_ context.Context, filename string, content io.Reader, _ string) (string, error) {
	dest := filepath.Join(s.uploadsDir, filename)
	if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
		return "", err
	}
	f, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	if _, err := io.Copy(f, content); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/uploads/%s", s.baseURL, filename), nil
}

func (s *LocalStorage) Delete(_ context.Context, filename string) error {
	err := os.Remove(filepath.Join(s.uploadsDir, filename))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
