package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const bucket = "covers"

type SupabaseStorage struct {
	projectURL string
	serviceKey string
	client     *http.Client
}

func NewSupabaseStorage(projectURL, serviceKey string) *SupabaseStorage {
	return &SupabaseStorage{
		projectURL: strings.TrimRight(projectURL, "/"),
		serviceKey: serviceKey,
		client:     &http.Client{},
	}
}

func (s *SupabaseStorage) Upload(ctx context.Context, filename string, content io.Reader, contentType string) (string, error) {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.projectURL, bucket, filename)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, content)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("supabase storage upload failed (%d): %s", resp.StatusCode, string(body))
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.projectURL, bucket, filename)
	return publicURL, nil
}

func (s *SupabaseStorage) Delete(ctx context.Context, filename string) error {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.projectURL, bucket, filename)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase storage delete failed (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}
