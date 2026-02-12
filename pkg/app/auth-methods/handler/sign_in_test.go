package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	auth_method "github.com/amorindev/go-tmpl/mocks/app/auth-methods"
	"github.com/amorindev/go-tmpl/pkg/app/auth-methods/core"
	"github.com/amorindev/go-tmpl/pkg/app/auth-methods/handler"
	sessionD "github.com/amorindev/go-tmpl/pkg/app/session/domain"
	userD "github.com/amorindev/go-tmpl/pkg/app/users/domain"
	sharedD "github.com/amorindev/go-tmpl/pkg/shared/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// * en todo el flujo se valida que se propague el contexto de la request
func TestHandler_SignIn(t *testing.T) {

	testTable := map[string]struct {
		setup         func(authMethodServiceMock *auth_method.MockAuthMethodSrv)
		reqBodyStr    string
		assertionFunc func(subTest *testing.T, rr *httptest.ResponseRecorder)
	}{
		"invalid json": {
			setup:      func(_ *auth_method.MockAuthMethodSrv) {},
			reqBodyStr: ``,
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusBadRequest, rr.Code)

				var appErr sharedD.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, sharedD.ErrCodeInvalidParams, appErr.Code)
				assert.Equal(subTest, "invalid request body", appErr.Msg)
			},
		},
		"empty email field": {
			setup:      func(_ *auth_method.MockAuthMethodSrv) {},
			reqBodyStr: `{}`,
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusBadRequest, rr.Code)

				var appErr sharedD.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, sharedD.ErrCodeInvalidParams, appErr.Code)
				assert.Equal(subTest, "email is required", appErr.Msg)
			},
		},
		"invalid email format": {
			setup: func(_ *auth_method.MockAuthMethodSrv) {},
			reqBodyStr: `{
				"email":"invalid_example.com"
			}`,
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusBadRequest, rr.Code)

				var appErr sharedD.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, sharedD.ErrCodeInvalidParams, appErr.Code)
				assert.Equal(subTest, "invalid email format", appErr.Msg)
			},
		},
		"empty password field": {
			setup: func(_ *auth_method.MockAuthMethodSrv) {},
			reqBodyStr: `{
				"email":"user@example.com"
			}`,
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusBadRequest, rr.Code)

				var appErr sharedD.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, sharedD.ErrCodeInvalidParams, appErr.Code)
				assert.Equal(subTest, "password is required", appErr.Msg)
			},
		},
		"user not found": {
			setup: func(authMethodServiceMock *auth_method.MockAuthMethodSrv) {
				authMethodServiceMock.EXPECT().SignIn(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, email string, password string, rememberMe bool) (*userD.User, *sessionD.Session, error) {
					return nil, nil, sharedD.ManageError(sharedD.ErrNotFound, "")
				})
			},
			reqBodyStr: `{
    			"email":"user@example.com",
    			"password":"123456()Pass"
			}`,
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusNotFound, rr.Code)

				var appErr sharedD.AppError
				err := json.Unmarshal(rr.Body.Bytes(), &appErr)
				require.NoError(subTest, err, "error unmarshaling response")

				assert.Equal(subTest, sharedD.ErrCodeNotFound, appErr.Code)
				assert.Equal(subTest, "Not found", appErr.Msg)

				fmt.Printf("error status code: %v\n", rr.Code)
				fmt.Printf("error code a: %v\n", appErr.Code)
				fmt.Printf("error message: %v\n", appErr.Msg)
			},
		},
		"success": {
			setup: func(authMethodServiceMock *auth_method.MockAuthMethodSrv) {
				authMethodServiceMock.EXPECT().SignIn(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, email string, password string, rememberMe bool) (*userD.User, *sessionD.Session, error) {
					session := &sessionD.Session{}
					user := &userD.User{}
					return user, session, sharedD.ManageError(sharedD.ErrNotFound, "")
				})
			},
			reqBodyStr: `{
    			"email":"user@example.com",
    			"password":"123456()Pass"
			}`,
			assertionFunc: func(subTest *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(subTest, http.StatusOK, rr.Code)

				var authResp core.AuthResp
				err := json.Unmarshal(rr.Body.Bytes(), &authResp)
				if err != nil {
					subTest.Fatal("error unmarshaling response: ", err.Error())
				}
				
				assert.NotNil(subTest,authResp.Session,"")
				assert.NotNil(subTest,authResp.User,"")
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

			req, err := http.NewRequest(http.MethodPost, "/auth/sign-in", bytes.NewBuffer([]byte(test.reqBodyStr)))
			require.NoError(subTest, err, "unexpected error creating request")

			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			test.setup(authMethodSrvMock)

			h.SignIn(rr, req)

			test.assertionFunc(subTest, rr)
		})
	}
}
