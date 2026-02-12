package minio_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileStorage_UploadImage(t *testing.T) {
	// Skip this test if the short flag is provided
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Minimal 20x20 PNG image encoded in base64 for testing
	const pngBase64 = "iVBORw0KGgoAAAANSUhEUgAAABQAAAAUCAYAAACNiR0NAAAAMUlEQVR42mP8//8/AwXgP6VgIJDYgCSmYIAESEoWYjAksQBaDQWDAwMDAwMDAM7UBfFehi1+AAAAAElFTkSuQmCC"

	// Decode base64 string to raw bytes
	imgData, err := base64.StdEncoding.DecodeString(pngBase64)
	assert.NoError(t, err)
	imgReader := bytes.NewReader(imgData)

	testTable := map[string]struct {
		ImgPath       string
		File          io.Reader
		ContentType   string
		assertionFunc func(subTest *testing.T, goErr error)
	}{
		"should succeed": {
			ImgPath:     "users/img.png",
			File:        imgReader,
			ContentType: "img/png",
			assertionFunc: func(subTest *testing.T, goErr error) {
				assert.NoError(subTest, err)
			},
		},
		"Object name cannot be empty": {
			ImgPath:     "a",
			File:        imgReader, // nil es panic
			ContentType: "",
			assertionFunc: func(subTest *testing.T, goErr error) {
				assert.NoError(subTest, goErr)

			},
		},
	}

	fileStgTC, err := startMinio(context.Background())
	if err != nil {
		t.Fatalf("error starting minio container: %v", err)
	}

	defer fileStgTC.Terminate(context.Background())

	fileStg, err := setupFileStorage(fileStgTC)
	if err != nil {
		t.Fatalf("error settings up file storage: %v", err)
	}

	// Run each test case using the previously uploaded test image
	for testName, test := range testTable {
		t.Run(testName, func(subTest *testing.T) {
			goErr := fileStg.UploadImage(context.Background(), test.ImgPath, test.File, test.ContentType)
			test.assertionFunc(subTest, goErr)
		})
	}
}
