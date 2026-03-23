package services

import (
	"bytes"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThumbnailExtension(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		content     string
		filename    string
		contentType string
		want        string
		wantErr     error
	}{
		{
			name:        "detect jpeg by bytes",
			content:     "\xff\xd8\xff\xe0\x00\x10JFIF",
			filename:    "thumb.bin",
			contentType: "application/octet-stream",
			want:        ".jpg",
		},
		{
			name:        "fallback to content type",
			content:     "not-an-image",
			filename:    "thumb.bin",
			contentType: "image/png",
			want:        ".png",
		},
		{
			name:        "fallback to filename extension",
			content:     "not-an-image",
			filename:    "thumb.webp",
			contentType: "application/octet-stream",
			want:        ".webp",
		},
		{
			name:        "unsupported format",
			content:     "not-an-image",
			filename:    "thumb.txt",
			contentType: "text/plain",
			wantErr:     ErrUnsupportedThumbnailFormat,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			file := multipart.File(nopMultipartFile{Reader: bytes.NewReader([]byte(tt.content))})
			got, err := thumbnailExtension(file, tt.filename, tt.contentType)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

type nopMultipartFile struct {
	*bytes.Reader
}

func (nopMultipartFile) Close() error {
	return nil
}

func TestAttachThumbnailURLOverridesStaleValue(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	service := NewImageService(nil, nil, nil, nil, "", tempDir)

	id := uuid.New()
	stale := "https://old.example.invalid/thumb.png"
	image := &models.Image{
		ID:           id,
		ThumbnailURL: &stale,
	}

	got := service.attachThumbnailURL(image)
	assert.Nil(t, got.ThumbnailURL)

	targetPath := filepath.Join(tempDir, id.String()+".png")
	require.NoError(t, os.WriteFile(targetPath, []byte("png"), 0o600))

	got = service.attachThumbnailURL(image)
	require.NotNil(t, got.ThumbnailURL)

	want := ThumbnailPublicBasePath + "/" + filepath.Base(targetPath)
	assert.Equal(t, want, *got.ThumbnailURL)
}
