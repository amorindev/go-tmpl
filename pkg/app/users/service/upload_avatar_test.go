package service_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/amorindev/go-tmpl/mocks/app/users"
	"github.com/amorindev/go-tmpl/pkg/app/users/service"
	sharedD "github.com/amorindev/go-tmpl/pkg/shared/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// * ver en que capa se testa la url la iamgen
// * aqui esta el tema de validar el path + la imagen
// * con doreturn del mock
// Faltaria el caso de generic en update image path
// ? 2 ways, complete text or use ManageError (esto es en el mock al hacer return del error)
// ? expErr := sharedD.ManageError(sharedD.ErrNotFound, "")
// ! va faltar el file en las test tables ver

// ver cual se va usar la imagen en texto 
// strings.NewReader("hola")
func TestService_UploadAvatar(t *testing.T) {
	ctx := context.TODO()

	testTable := map[string]struct {
		setup         func(mockRepo *users.MockUserRepo, mockFileStorage *users.MockUserFileStg)
		ctx           context.Context
		userID        string
		img           string
		file          io.Reader
		contentType   string
		assertionFunc func(subTest *testing.T, gotErr error)
	}{
		"Exists, generic error": {
			setup: func(mockRepo *users.MockUserRepo, mockFileStorage *users.MockUserFileStg) {
				mockRepo.EXPECT().Exists(ctx, gomock.Any()).Return(false, errors.New("generic error"))
			},
			ctx:    ctx,
			userID: "123",
			assertionFunc: func(subTest *testing.T, gotErr error) {
				assert.NotNil(subTest, gotErr)

				var appErr sharedD.AppError
				if !errors.As(gotErr, &appErr) {
					subTest.Fatalf("expect AppError, got %T", gotErr)
				}

				assert.Equal(subTest, sharedD.ErrCodeInternalServerError, appErr.Code)
				assert.Equal(subTest, "Server Error", appErr.Msg)
			},
		},

		"user does not exists": {
			setup: func(mockRepo *users.MockUserRepo, mockFileStorage *users.MockUserFileStg) {
				mockRepo.EXPECT().Exists(ctx, "123").Return(false, nil)
			},
			ctx:    ctx,
			userID: "123",
			assertionFunc: func(subTest *testing.T, gotErr error) {
				assert.NotNil(subTest, gotErr)

				var appErr sharedD.AppError
				if !errors.As(gotErr, &appErr) {
					t.Fatalf("except AppError, got %T", gotErr)
				}

				assert.Equal(subTest, sharedD.ErrCodeNotFound, appErr.Code)
				assert.Equal(subTest, appErr.Msg, "Not found")
			},
		},

		"UploadImage, generic error": {
			setup: func(mockRepo *users.MockUserRepo, mockFileStorage *users.MockUserFileStg) {
				mockRepo.EXPECT().Exists(ctx, "123").Return(true, nil)

				path := "users/123.png"

				mockFileStorage.EXPECT().UploadImage(ctx, path, gomock.Any(), "image/png").Return(errors.New("generic error"))
			},
			ctx:         ctx,
			userID:      "123",
			img:         "img.png",
			contentType: "image/png",
			assertionFunc: func(subTest *testing.T, gotErr error) {
				assert.NotNil(subTest, gotErr)

				var appErr sharedD.AppError
				if !errors.As(gotErr, &appErr) {
					t.Fatalf("expected AppError, got %T", gotErr)
				}

				assert.Equal(subTest, sharedD.ErrCodeInternalServerError, appErr.Code)
				assert.Equal(subTest, appErr.Msg, "Server Error")
			},
		},
		"UpdateAvatarPath, incorrectID": {
			setup: func(mockRepo *users.MockUserRepo, mockFileStorage *users.MockUserFileStg) {
				mockRepo.EXPECT().Exists(ctx, "123").Return(true, nil)

				path := "users/123.png"

				mockFileStorage.EXPECT().UploadImage(ctx, path, gomock.Any(), "image/png").Return(nil)

				someError := errors.New("some error")

				mockRepo.EXPECT().UpdateAvatarPath(ctx, "123", path, gomock.Any()).DoAndReturn(func(ctx context.Context, userID string, imgPath string, updatedAt time.Time) error {
					require.False(t, updatedAt.IsZero())
					return fmt.Errorf("%w: failed to convert userID to ObjectID :%w", sharedD.ErrIncorrectID, someError)
				})
			},
			ctx:         ctx,
			userID:      "123",
			img:         "img.png",
			contentType: "image/png",
			assertionFunc: func(subTest *testing.T, gotErr error) {
				assert.NotNil(subTest, gotErr)

				var appErr sharedD.AppError
				if !errors.As(gotErr, &appErr) {
					t.Fatalf("expected AppError, got %T", gotErr)
				}

				assert.Equal(subTest, sharedD.ErrCodeInvalidParams, appErr.Code)
				assert.Equal(subTest, "Incorrect id", appErr.Msg)
			},
		},
		"success": {
			setup: func(mockRepo *users.MockUserRepo, mockFileStorage *users.MockUserFileStg) {
				mockRepo.EXPECT().Exists(ctx, "123").Return(true, nil)

				// ! validar el compo imagen?
				mockFileStorage.EXPECT().UploadImage(ctx, "users/123.png", gomock.Any(), "image/png").Return(nil)

				mockRepo.EXPECT().UpdateAvatarPath(ctx, "123", "users/123.png", gomock.Any()).DoAndReturn(func(ctx context.Context, userID string, imgPath string, updatedAt time.Time) error {
					require.False(t, updatedAt.IsZero())
					return nil
				})
			},
			ctx:         ctx,
			userID:      "123",
			img:         "img.png",
			contentType: "image/png",
			file:        bytes.NewBufferString("fake"),
			assertionFunc: func(subTest *testing.T, gotErr error) {
				assert.Nil(subTest, gotErr)
			},
		},
	}

	for testName, test := range testTable {
		t.Run(testName, func(subTest *testing.T) {

			repoCtrl := gomock.NewController(t)
			defer repoCtrl.Finish()

			fileStgCtrl := gomock.NewController(t)
			defer fileStgCtrl.Finish()

			mockUserRepo := users.NewMockUserRepo(repoCtrl)
			mockUserFileStg := users.NewMockUserFileStg(fileStgCtrl)

			s := &service.Service{
				UserRepo:    mockUserRepo,
				UserFileStg: mockUserFileStg,
			}

			test.setup(mockUserRepo, mockUserFileStg)

			gotErr := s.UploadAvatar(test.ctx, test.img, test.file, test.userID, test.contentType)

			test.assertionFunc(subTest, gotErr)
		})
	}

}
