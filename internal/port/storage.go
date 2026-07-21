package port

import (
	"context"
	"io"
)

type FileStorage interface {
	Upload(ctx context.Context, filename string, content io.Reader, contentType string) (publicURL string, err error)
	Delete(ctx context.Context, filename string) error
}
