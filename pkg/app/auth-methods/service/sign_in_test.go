package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/amorindev/go-tmpl/internal/encryption"
	"github.com/amorindev/go-tmpl/mocks/app/users"
	authMethodD "github.com/amorindev/go-tmpl/pkg/app/auth-methods/domain"
	"github.com/amorindev/go-tmpl/pkg/app/auth-methods/service"
	sessionD "github.com/amorindev/go-tmpl/pkg/app/session/domain"
	sharedD "github.com/amorindev/go-tmpl/pkg/shared/domain"

	"github.com/amorindev/go-tmpl/pkg/app/users/domain"

	"github.com/amorindev/go-tmpl/mocks/app/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ? como probamos los errores que nos retorna HashPassword o comapare passwor password do not match

// como validar las injecciones  como cantida de refresh token
// como validar esto
/*
now := time.Now().UTC()
session.AccessToken = aToken
session.AccessTokenExpIn = aTokenExpIn
session.RefreshTokenID = rTokenID
session.RefreshToken = rToken
session.RefreshTokenExpIn = rTokenExpIn
session.Revoked = false
session.CreatedAt = &now
*/
// ! poner parametros en gomockAny va servir para el upload image el backet mas el nombre de la imagen
// * usar gotErr para todo los test, _ si no se usa el parámetro, ver el context

// El valor de hash es el mismo para todos los sub-tests.
// No necesitas recalcularlo dentro de cada caso.
// evita duplicación,
// deja claro que es un dato común a todos los casos.

// ! falta los mensajes de error como probar en todos los test, equal es casi como contains pero
// ! asegurarse el error fmt.Printf("Error is: %v\n", gotErr.Error())

// ? user ? ver cual se reutiliza por que se usa el email en varios lados
// ? el email se repite ver donde ponerlo para user.email pero que no afecte a los test

// ? de momento para los errores genericos dire generic error pero se debe validar el mensage de error
// ? del genérico

// para ser veloces nos cquedamos con validar el code y message el string de moemnto no

// para los get o sign in el tipo de contexto debe ser request

// aca usa img en ces de repeat user/ ver los casos para no duplicar email

// * falta esto user.UserPassAuth = nil, es con el do and return
func TestService_SignIn(t *testing.T) {
	// Data common to all cases
	// * hash? user?
	// * depende de la logica al fuinal si creas desde dentro igualmente lo asignas
	ctx := context.TODO()

	hash, err := encryption.HashPassword("password")
	require.NoError(t, err)

	testTable := map[string]struct {
		setup         func(mockUserRepo *users.MockUserRepo, mockUserFileStg *users.MockUserFileStg, mockSessionSrv *sessions.MockSessionSrv)
		ctx           context.Context
		email         string
		password      string
		rememberMe    bool
		assertionFunc func(subTest *testing.T, gotUser *domain.User, gotSession *sessionD.Session, gotErr error)
	}{

		"FindByEmail, generic error": {
			setup: func(mockUserRepo *users.MockUserRepo, _ *users.MockUserFileStg, _ *sessions.MockSessionSrv) {
				mockUserRepo.EXPECT().FindByEmail(gomock.Any(), "user@example.com").Return(nil, errors.New("generic error"))
			},
			ctx:        ctx,
			email:      "user@example.com",
			password:   "",
			rememberMe: false,
			assertionFunc: func(subTest *testing.T, gotUser *domain.User, gotSession *sessionD.Session, gotErr error) {
				require.Error(subTest, gotErr)
				assert.Nil(subTest, gotUser)
				assert.Nil(subTest, gotSession)

				var appErr sharedD.AppError
				if !errors.As(gotErr, &appErr) {
					t.Fatalf("except AppError, got %T", gotErr)
				}

				assert.Equal(subTest, sharedD.ErrCodeInternalServerError, appErr.Code)
				assert.Equal(subTest, appErr.Msg, "Server Error")
			},
		},

		/* "FindByEmail, user not found": {}, */

		/* "The user doesn't have an active account": {}, */

		"CheckPassword, password do not match": {
			setup: func(mockUserRepo *users.MockUserRepo, _ *users.MockUserFileStg, _ *sessions.MockSessionSrv) {
				user := &domain.User{
					ID:       "123",
					IsActive: true,
					UserPassAuth: &authMethodD.UserPasswordAuth{
						PasswordHash: hash,
					},
				}
				mockUserRepo.EXPECT().FindByEmail(gomock.Any(), "user@example.com").Return(user, nil)
			},
			ctx:        ctx,
			email:      "user@example.com",
			password:   "badPassword",
			rememberMe: false,
			assertionFunc: func(subTest *testing.T, gotUser *domain.User, gotSession *sessionD.Session, gotErr error) {
				require.Error(subTest, gotErr)
				assert.Nil(subTest, gotUser)
				assert.Nil(subTest, gotSession)

				var appErr sharedD.AppError
				if !errors.As(gotErr, &appErr) {
					t.Fatalf("except AppError, got %T", gotErr)
				}

				assert.Equal(subTest, sharedD.ErrCodeUnauthorized, appErr.Code)
				assert.Contains(subTest, "password does not match", appErr.Msg)
			},
		},

		"CheckPassword, generic error": {
			setup: func(mockUserRepo *users.MockUserRepo, _ *users.MockUserFileStg, _ *sessions.MockSessionSrv) {
				badHashUser := &domain.User{
					ID:       "123",
					Email:    "user@example.com",
					IsActive: true,
					UserPassAuth: &authMethodD.UserPasswordAuth{
						PasswordHash: "not-a-bcrypt-hash",
					},
				}
				mockUserRepo.EXPECT().FindByEmail(gomock.Any(), "user@example.com").Return(badHashUser, nil)
			},
			ctx:        ctx,
			email:      "user@example.com",
			password:   "password",
			rememberMe: false,
			assertionFunc: func(subTest *testing.T, gotUser *domain.User, gotSession *sessionD.Session, gotErr error) {
				require.Error(subTest, gotErr)
				assert.Nil(subTest, gotUser)
				assert.Nil(subTest, gotSession)

				var appErr sharedD.AppError
				if !errors.As(gotErr, &appErr) {
					t.Fatalf("except AppError, got %T", gotErr)
				}

				assert.Equal(subTest, sharedD.ErrCodeInternalServerError, appErr.Code)
				assert.Contains(subTest, "Server Error", appErr.Msg)
			},
		},

		"GetImage, generic error": {
			setup: func(mockUserRepo *users.MockUserRepo, mockUserFileStg *users.MockUserFileStg, _ *sessions.MockSessionSrv) {
				img := "users/user_id.png"
				user := &domain.User{
					ID:       "123",
					IsActive: true,
					UserPassAuth: &authMethodD.UserPasswordAuth{
						PasswordHash: hash,
					},
					ImgPath: &img,
				}

				mockUserRepo.EXPECT().FindByEmail(gomock.Any(), "user@example.com").Return(user, nil)

				mockUserFileStg.EXPECT().GetImage(gomock.Any(), img).Return("", errors.New("generic error"))
			},
			ctx:        ctx,
			email:      "user@example.com",
			password:   "password",
			rememberMe: false,
			assertionFunc: func(subTest *testing.T, gotUser *domain.User, gotSession *sessionD.Session, gotErr error) {
				require.Error(subTest, gotErr)
				assert.Nil(subTest, gotUser)
				assert.Nil(subTest, gotSession)

				var appErr sharedD.AppError
				if !errors.As(gotErr, &appErr) {
					t.Fatalf("except AppError, got %T", gotErr)
				}

				assert.Equal(subTest, sharedD.ErrCodeInternalServerError, appErr.Code)
				assert.Equal(subTest, "Server Error", appErr.Msg)
			},
		},

		"success without image": {
			setup: func(mockUserRepo *users.MockUserRepo, _ *users.MockUserFileStg, mockSessionSrv *sessions.MockSessionSrv) {
				user := &domain.User{
					ID:       "123",
					IsActive: true,
					UserPassAuth: &authMethodD.UserPasswordAuth{
						PasswordHash: hash,
					},
				}
				mockUserRepo.EXPECT().FindByEmail(gomock.Any(), "user@example.com").Return(user, nil)
				mockSessionSrv.EXPECT().Create(gomock.Any(), gomock.Any(), nil, "user@example.com").Return(nil)
			},
			ctx:        ctx,
			email:      "user@example.com",
			password:   "password",
			rememberMe: false,
			assertionFunc: func(subTest *testing.T, gotUser *domain.User, gotSession *sessionD.Session, gotErr error) {
				assert.Nil(t, err)
				assert.Equal(t, "123", gotUser.ID)
				assert.NotNil(t, gotSession)
			},
		},
		/* "success with image":{}, */
	}

	for testName, test := range testTable {
		t.Run(testName, func(subTest *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := users.NewMockUserRepo(ctrl)
			mockUserFileStg := users.NewMockUserFileStg(ctrl)
			mockSessionSrv := sessions.NewMockSessionSrv(ctrl)

			s := service.Service{
				UserRepo:    mockUserRepo,
				SessionSrv:  mockSessionSrv,
				UserFileStg: mockUserFileStg,
			}

			test.setup(mockUserRepo, mockUserFileStg, mockSessionSrv)

			gotUser, gotSession, gotErr := s.SignIn(test.ctx, test.email, test.password, test.rememberMe)

			test.assertionFunc(subTest, gotUser, gotSession, gotErr)
		})
	}
}
