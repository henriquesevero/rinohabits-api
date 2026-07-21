package handler

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
)

var allowedImageTypes = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

var errImageContentMismatch = errors.New("file content does not match its extension")

func validateImageContent(file multipart.File, expectedContentType string) error {
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if http.DetectContentType(buf[:n]) != expectedContentType {
		return errImageContentMismatch
	}
	return nil
}
