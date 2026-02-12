package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	auth_method "github.com/amorindev/go-tmpl/mocks/app/auth-methods"
	"github.com/amorindev/go-tmpl/pkg/app/auth-methods/handler"
	"github.com/amorindev/go-tmpl/pkg/app/users/core"
	"github.com/amorindev/go-tmpl/pkg/app/users/domain"
	sharedD "github.com/amorindev/go-tmpl/pkg/shared/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)


func TestHandler_SignUp(t *testing.T) {

	testTable := map[string]struct {
		setup         func(authMethodServiceMock *auth_method.MockAuthMethodSrv)
		reqBodyStr    string
		assertionFunc func(subTest *testing.T, rr *httptest.ResponseRecorder)
	}{
		"invalid json": {
			setup:      func(_ *auth_method.MockAuthMethodSrv) {},
			reqBodyStr: `{"invalid":"json"`,
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusBadRequest, rr.Code)

				var appErr sharedD.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, appErr.Code, sharedD.ErrCodeInvalidParams)
				assert.Equal(subTest, appErr.Msg, "invalid request body")
			},
		},
		"IsSignUpValid, email is required": {
			setup:      func(_ *auth_method.MockAuthMethodSrv) {},
			reqBodyStr: `{}`,
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusBadRequest, rr.Code)

				var appErr sharedD.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, appErr.Code, sharedD.ErrCodeInvalidParams)
				assert.Equal(subTest, appErr.Msg, "email is required")
			},
		},
		"IsSignUpValid, invalid email format": {
			setup:      func(_ *auth_method.MockAuthMethodSrv) {},
			reqBodyStr: `{"email":"user_example.com"}`,
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusBadRequest, rr.Code)

				var appErr sharedD.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, sharedD.ErrCodeInvalidParams, appErr.Code)
				assert.Equal(subTest, "invalid email format", appErr.Msg)
			},
		},
		"IsSignUpValid, password is required": {
			setup:      func(_ *auth_method.MockAuthMethodSrv) {},
			reqBodyStr: `{"email":"user@example.com"}`,
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusBadRequest, rr.Code)

				var appErr sharedD.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, sharedD.ErrCodeInvalidParams, appErr.Code)
				assert.Equal(subTest, "password is required", appErr.Msg)
			},
		},
		"IsSignUpValid, confirm password is required": {
			setup: func(_ *auth_method.MockAuthMethodSrv) {},
			reqBodyStr: `{
				"email":"user@example.com",
				"password":"password"
			}`,
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusBadRequest, rr.Code)

				var appErr sharedD.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, sharedD.ErrCodeInvalidParams, appErr.Code)
				assert.Equal(subTest, "confirm password is required", appErr.Msg)

			},
		},
		"IsSignUpValid, weak password": {
			setup: func(_ *auth_method.MockAuthMethodSrv) {},
			reqBodyStr: `{
				"email":"user@example.com",
				"password":"pass",
				"confirm_password":"pass"
			}`,
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusBadRequest, rr.Code)

				var appErr sharedD.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, sharedD.ErrCodeInvalidParams, appErr.Code)
				assert.Equal(subTest, "password must be at least 8 characters long", appErr.Msg)

			},
		},
		"IsSignUpValid, invalid password: missing uppercase, lowercase, or number": {
			setup: func(_ *auth_method.MockAuthMethodSrv) {},
			reqBodyStr: `{
				"email":"user@example.com",
				"password":"password_123",
				"confirm_password":"password_123"
			}`,
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusBadRequest, rr.Code)

				var appErr sharedD.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				msg := "password must contain at least one uppercase letter, one lowercase letter, and one number"
				assert.Equal(subTest, sharedD.ErrCodeInvalidParams, appErr.Code)
				assert.Equal(subTest, msg, appErr.Msg)

				fmt.Printf("error status code: %v\n", rr.Code)
				fmt.Printf("error code: %v\n", appErr.Code)
				fmt.Printf("error message: %v\n", appErr.Msg)

			},
		},
		"passwords don't match": {
			setup: func(_ *auth_method.MockAuthMethodSrv) {},
			reqBodyStr: `{
				"email":"user@example.com",
				"password": "validPass()123",
				"confirm_password": "validPass()"
			}`,
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusBadRequest, rr.Code)

				var appErr sharedD.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, sharedD.ErrCodeInvalidParams, appErr.Code)
				assert.Equal(subTest, "passwords do not match", appErr.Msg)
			},
		},
		"email already exists": {
			setup: func(authMethodServiceMock *auth_method.MockAuthMethodSrv) {
				authMethodServiceMock.EXPECT().SignUp(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *domain.User) error {
					return sharedD.ManageError(sharedD.ErrDuplicateKey, "email already in use")
				})
			},
			reqBodyStr: `{
				"email":"user@yopmail.com",
                "password":"validPass()123",
                "confirm_password":"validPass()123"
			}`,
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusConflict, rr.Code)

				var appErr sharedD.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, sharedD.ErrCodeDuplicateKey, appErr.Code)
				assert.Equal(subTest, "Duplicate key: email already in use", appErr.Msg)
			},
		},

		"duplicate key": {
			setup: func(authMethodServiceMock *auth_method.MockAuthMethodSrv) {
				authMethodServiceMock.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(sharedD.ManageError(sharedD.ErrDuplicateKey, "email already in use"))
			},
			reqBodyStr: `
                {
                    "email":"user@yopmail.com",
                    "password":"123456()Pass",
                    "confirm_password":"123456()Pass"
                }
            `,
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusConflict, rr.Code)

				var respAppErr sharedD.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &respAppErr)
				if err != nil {
					subTest.Fatal("error unmarshaling response: ", err.Error())
				}

				assert.Equal(subTest, respAppErr.Code, sharedD.ErrCodeDuplicateKey)

				// ! hay dos formas la otra es poniend la entidad como en email service ver
				assert.Contains(subTest, respAppErr.Error(), "Duplicate key: email already in use")
			},
		},
		"success": {
			setup: func(authMethodServiceMock *auth_method.MockAuthMethodSrv) {
				authMethodServiceMock.EXPECT().SignUp(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *domain.User) error {
					now := time.Now().UTC()
					u.ID = "123"
					u.CreatedAt = &now
					u.UpdatedAt = &now
					return nil
				})
			},
			reqBodyStr: `
                {
                    "email":"user@yopmail.com",
                    "password":"123456()Pass",
                    "confirm_password":"123456()Pass"
                }
            `,
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusOK, rr.Code)

				var respUser core.UserCore
				err := json.Unmarshal(rr.Body.Bytes(), &respUser)
				if err != nil {
					subTest.Fatal("error unmarshaling response: ", err.Error())
				}

				assert.Equal(subTest, "user@yopmail.com", respUser.Email)
				assert.False(subTest, respUser.EmailVerified)
				assert.Equal(subTest, "123", respUser.ID)
				assert.Nil(subTest, respUser.ImgUrl)
				assert.NotNil(subTest, respUser.CreatedAt)
				assert.NotNil(subTest, respUser.UpdatedAt)
			},
		},
	}

	for testName, test := range testTable {
		t.Run(testName, func(subTest *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authMethodSrvMock := auth_method.NewMockAuthMethodSrv(ctrl)

			h := handler.Handler{
				AuthMethodSrv: authMethodSrvMock,
			}

			req, err := http.NewRequest(http.MethodPost, "/auth/sign-up", bytes.NewBuffer([]byte(test.reqBodyStr)))
			require.NoError(subTest, err, "unexpected error creating request")

			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			test.setup(authMethodSrvMock)

			h.SignUp(rr, req)

			test.assertionFunc(subTest, rr)
		})
	}

}
