package handler_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amorindev/go-tmpl/mocks/app/users"
	"github.com/amorindev/go-tmpl/pkg/app/users/handler"
	"github.com/amorindev/go-tmpl/pkg/shared/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

/*
fmt.Printf("error status code: %v\n", rr.Code)
fmt.Printf("error code a: %v\n", appErr.Code)
fmt.Printf("error message: %v\n", appErr.Msg)
*/
func TestHandler_UploadAvatar(t *testing.T) {

	testTable := map[string]struct {
		setup         func(userServiceMock *users.MockUserSrv) (*http.Request, error)
		assertionFunc func(subTest *testing.T, rr *httptest.ResponseRecorder)
	}{
		"missing userID": {
			setup: func(_ *users.MockUserSrv) (*http.Request, error) {
				req, err := http.NewRequest(http.MethodPost, "/users//avatar", nil)
				return req, err
			},
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusBadRequest, rr.Code)

				var appErr domain.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, appErr.Code, domain.ErrCodeInvalidParams)
				assert.Equal(subTest, appErr.Msg, "missing user id")
			},
		},
		"parse multiform error (empty or corrupt body)": {
			setup: func(_ *users.MockUserSrv) (*http.Request, error) {
				req, err := http.NewRequest(http.MethodPost, "/users//avatar", bytes.NewReader([]byte("bad body")))
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", "multipart/form-data; boundary=missing")
				req.SetPathValue("userId", "123")
				return req, err
			},
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusBadRequest, rr.Code)

				var appErr domain.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, appErr.Code, domain.ErrCodeInvalidParams)
				assert.Equal(subTest, appErr.Msg, "invalid form")
			},
		},
		"form file missing": {
			setup: func(userServiceMock *users.MockUserSrv) (*http.Request, error) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				err := writer.Close()
				if err != nil {
					return nil, err
				}
				req, err := http.NewRequest(http.MethodPost, "/users//avatar", body)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.SetPathValue("userId", "123")
				return req, err
			},
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusBadRequest, rr.Code)

				var appErr domain.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, appErr.Code, domain.ErrCodeInvalidParams)
				assert.Equal(subTest, appErr.Msg, "error retrieving file")
			},
		},
		"invalid image format": {
			setup: func(userServiceMock *users.MockUserSrv) (*http.Request, error) {
				body, cType, err := newMultipartBody("image", "bad.txt", []byte("plain text"))
				if err != nil {
					return nil, err
				}
				req, err := http.NewRequest(http.MethodPost, "/users//avatar", body)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", cType)
				req.SetPathValue("userId", "123")

				return req, err
			},
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusBadRequest, rr.Code)

				var appErr domain.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, appErr.Code, domain.ErrCodeInvalidParams)
				assert.Equal(subTest, appErr.Msg, "invalid image format")
			},
		},
		"user not found": {
			setup: func(userServiceMock *users.MockUserSrv) (*http.Request, error) {
				imgData, err := createImg()
				if err != nil {
					return nil, err
				}
				body, cType, err := newMultipartBody("image", "img.png", imgData)
				if err != nil {
					return nil, err
				}

				req, err := http.NewRequest(http.MethodPost, "/users//avatar", body)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", cType)
				req.SetPathValue("userId", "123")

				userServiceMock.EXPECT().UploadAvatar(context.Background(), "img.png", gomock.Any(), "123", "image/png").Return(domain.ManageError(domain.ErrNotFound, ""))

				return req, err
			},
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusNotFound, rr.Code)

				var appErr domain.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, appErr.Code, domain.ErrCodeNotFound)
				assert.Equal(subTest, appErr.Msg, "Not found")
			},
		},
		"success": {
			setup: func(userServiceMock *users.MockUserSrv) (*http.Request, error) {
				imgData, err := createImg()
				if err != nil {
					return nil, err
				}

				body, cType, err := newMultipartBody("image", "img.png", imgData)
				if err != nil {
					return nil, err
				}

				req, err := http.NewRequest(http.MethodPost, "/users//avatar", body)
				if err != nil {
					return nil, err
				}

				req.Header.Set("Content-Type", cType)
				req.SetPathValue("userId", "123")

				userServiceMock.EXPECT().UploadAvatar(context.Background(), "img.png", gomock.Any(), "123", "image/png").Return(nil)

				return req, err
			},
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusOK, rr.Code)
			},
		},
	}

	for testName, test := range testTable {
		t.Run(testName, func(subTest *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userSrvMock := users.NewMockUserSrv(ctrl)

			h := handler.Handler{
				UserSrv: userSrvMock,
			}

			rr := httptest.NewRecorder()

			req, err := test.setup(userSrvMock)
			if err != nil {
				t.Fatal("unexpected error creating request: ", err.Error())
			}

			h.UploadAvatar(rr, req)

			test.assertionFunc(subTest, rr)
		})
	}

}

func createImg() ([]byte, error) {
	// Minimal 20x20 PNG image encoded in base64 for testing
	const pngBase64 = "iVBORw0KGgoAAAANSUhEUgAAABQAAAAUCAYAAACNiR0NAAAAMUlEQVR42mP8//8/AwXgP6VgIJDYgCSmYIAESEoWYjAksQBaDQWDAwMDAwMDAM7UBfFehi1+AAAAAElFTkSuQmCC"

	// Decode base64 string to raw bytes
	return base64.StdEncoding.DecodeString(pngBase64)
}

func newMultipartBody(fieldName, fileName string, data []byte) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return nil, "", fmt.Errorf("error creating form file: %w", err)
	}
	_, err = part.Write(data)
	if err != nil {
		return nil, "", fmt.Errorf("error writing data: %w", err)
	}
	err = writer.Close()
	if err != nil {
		return nil, "", fmt.Errorf("error closing writer: %w", err)
	}
	return body, writer.FormDataContentType(), nil
}
