package minio_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	minioC "github.com/amorindev/go-tmpl/internal/minio"
	userFileStorage "github.com/amorindev/go-tmpl/pkg/app/users/file-storage/minio"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	minioTC "github.com/testcontainers/testcontainers-go/modules/minio"
)

func TestFileStorage_GetImage(t *testing.T) {
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
		assertionFunc func(subTest *testing.T, url string, err error)
	}{
		"should succeed": {
			ImgPath: "users/img.png",
			assertionFunc: func(subTest *testing.T, url string, err error) {
				assert.NoError(subTest, err)
				assert.NotEmpty(subTest, url)
				assert.Contains(subTest, url,"users/img.png")

				// Download the image from the signed URL
				resp, reqErr := http.Get(url)
				require.NoError(subTest, reqErr)
				defer resp.Body.Close()

				// The response should be HTTP 200 OK
				assert.Equal(subTest, http.StatusOK, resp.StatusCode)

				// Read the body and compare with the original bytes
				body, readErr := io.ReadAll(resp.Body)
				require.NoError(subTest, readErr)
				assert.Equal(subTest, imgData, body)
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

	// Prepare the test image
	err = fileStg.UploadImage(context.Background(), "users/img.png", imgReader, "img/png")
	require.NoError(t, err)

	// Run each test case using the previously uploaded test image
	for testName, test := range testTable {
		t.Run(testName, func(subTest *testing.T) {
			url, err := fileStg.GetImage(context.Background(), test.ImgPath)
			test.assertionFunc(subTest, url, err)
		})
	}
}

// setupFileStorage initializes the MinIO client and bucket
func setupFileStorage(fileStg *minioTC.MinioContainer) (*userFileStorage.FileStorage, error) {
	// Get the container endpoint
	endPoint, err := fileStg.Endpoint(context.Background(), "")
	if err != nil {
		return nil, fmt.Errorf("failed to get minio endpoint: %w", err)
	}

	// Create minio Client
	client, err := minioC.NewClient(endPoint, "minio", "minioPass", false)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MinIO Client: %w", err)
	}

	// Create bucket if not exists
	const bucketName = "go-tmpl"
	if err := client.CreateStorage(bucketName); err != nil {
		return nil, fmt.Errorf("failed to create bucket: %w", err)
	}

	return &userFileStorage.FileStorage{
		ExpTime:     time.Hour * 24 * 7,
		MinioClient: client.Client,
		BucketName:  bucketName,
	}, nil
}

// startMinio will return a minio testcontainer instance or an error
func startMinio(ctx context.Context) (*minioTC.MinioContainer, error) {
	minioContainer, err := minioTC.Run(ctx, "minio/minio:RELEASE.2024-01-16T16-07-38Z", minioTC.WithUsername("minio"), minioTC.WithPassword("minioPass"))
	if err != nil {
		return nil, fmt.Errorf("error running minio container: %w", err)
	}
	return minioContainer, nil
}
