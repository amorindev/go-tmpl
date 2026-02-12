package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/amorindev/go-tmpl/mocks/app/sessions"
	tokens "github.com/amorindev/go-tmpl/mocks/in_ter_nal/token"
	"github.com/amorindev/go-tmpl/pkg/app/session/domain"
	"github.com/amorindev/go-tmpl/pkg/app/session/service"
	sharedD "github.com/amorindev/go-tmpl/pkg/shared/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_Create(t *testing.T) {

	ctx := context.TODO()

	testTable := map[string]struct {
		setup         func(subTest *testing.T, mockSessionRepo *sessions.MockSessionRepo, mockTokenSrv *tokens.MockTokenSrv)
		ctx           context.Context
		session       *domain.Session
		roles         []string
		email         string
		assertionFunc func(subTest *testing.T, gotSession *domain.Session, gotErr error)
	}{
		"fails when access token generation fails": {
			// Verifies that when CreateAccessToken fails,
			// the service returns the same error and stops execution.
			setup: func(_ *testing.T, _ *sessions.MockSessionRepo, mockTokenSrv *tokens.MockTokenSrv) {
				mockTokenSrv.EXPECT().CreateAccessToken("123", "user@example.com", nil).Return("", int64(0), errors.New("generic error"))
			},
			ctx:     ctx,
			session: &domain.Session{UserID: "123"},
			roles:   nil,
			email:   "user@example.com",
			assertionFunc: func(subTest *testing.T, gotSession *domain.Session, gotErr error) {
				require.Error(subTest, gotErr)
				require.Equal(subTest, errors.New("generic error"), gotErr)
			},
		},
		"fails when refresh token generation fails": {
			// Verifies that when CreateRefreshToken fails,
			// the service returns the same error and stops execution.
			setup: func(_ *testing.T, mockSessionRepo *sessions.MockSessionRepo, mockTokenSrv *tokens.MockTokenSrv) {
				mockTokenSrv.EXPECT().CreateAccessToken("123", "user@example.com", nil).Return("", int64(0), nil)
				mockTokenSrv.EXPECT().CreateRefreshToken("123", true).Return("", "", int64(0), errors.New("generic error"))
			},
			ctx:     ctx,
			session: &domain.Session{UserID: "123", RememberMe: true},
			email:   "user@example.com",
			roles:   nil,
			assertionFunc: func(subTest *testing.T, gotSession *domain.Session, gotErr error) {
				require.Error(subTest, gotErr)
				assert.Equal(subTest, errors.New("generic error"), gotErr)
			},
		},

		"fails when repository insert returns ErrIncorrectID": {
			// Ensures that if the repository fails on Insert
			// (for example due to an invalid ObjectID),
			// the service propagates the exact sharedD.ErrIncorrectID error.
			setup: func(_ *testing.T, mockSessionRepo *sessions.MockSessionRepo, mockTokenSrv *tokens.MockTokenSrv) {
				mockTokenSrv.EXPECT().CreateAccessToken("invalid-id", "user@example.com", nil).Return("access-token", int64(3600), nil)

				mockTokenSrv.EXPECT().CreateRefreshToken("invalid-id", true).Return("132", "refresh-token", int64(7200), nil)

				mockSessionRepo.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(sharedD.ErrIncorrectID)
			},
			ctx:     ctx,
			session: &domain.Session{UserID: "invalid-id", RememberMe: true},
			roles:   nil,
			email:   "user@example.com",
			assertionFunc: func(subTest *testing.T, gotSession *domain.Session, gotErr error) {

			},
		},

		"success": {
			setup: func(subTest *testing.T, mockSessionRepo *sessions.MockSessionRepo, mockTokenSrv *tokens.MockTokenSrv) {
				mockTokenSrv.EXPECT().CreateAccessToken("123", "user@example.com", nil).Return("access-token", int64(3600), nil)

				mockTokenSrv.EXPECT().CreateRefreshToken("123", true).Return("132", "refresh-token", int64(7200), nil)

				mockSessionRepo.EXPECT().Insert(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, s *domain.Session) error {
					require.NotEmpty(subTest, s.AccessToken)
					require.NotZero(subTest, s.AccessTokenExpIn)
					require.NotEmpty(subTest, s.RefreshToken)
					require.NotZero(subTest, s.RefreshTokenExpIn)
					require.Equal(subTest, "123", s.UserID)
					require.False(subTest, s.Revoked)
					require.NotNil(subTest, s.CreatedAt)
					s.ID = "123"
					return nil
				})
			},
			ctx:     ctx,
			session: &domain.Session{UserID: "123", RememberMe: true},
			roles:   nil,
			email:   "user@example.com",
			assertionFunc: func(subTest *testing.T, gotSession *domain.Session, gotErr error) {
				require.NoError(subTest, gotErr)
				assert.NotEmpty(subTest, gotSession.AccessToken)
				assert.NotZero(subTest, gotSession.AccessTokenExpIn)
				assert.NotEmpty(subTest, gotSession.RefreshToken)
				assert.NotZero(subTest, gotSession.RefreshTokenExpIn)
				require.Equal(subTest, "123", gotSession.ID)
				assert.False(subTest, gotSession.Revoked)
				assert.NotNil(subTest, gotSession.CreatedAt)
			},
		},
	}

	for testName, test := range testTable {
		t.Run(testName, func(subTest *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTokenSrv := tokens.NewMockTokenSrv(ctrl)

			mockSessionRepo := sessions.NewMockSessionRepo(ctrl)

			srv := service.NewSessionSrv(mockSessionRepo, mockTokenSrv)

			test.setup(subTest, mockSessionRepo, mockTokenSrv)

			gotErr := srv.Create(test.ctx, test.session, test.roles, test.email)

			test.assertionFunc(subTest, test.session, gotErr)
		})
	}
}
