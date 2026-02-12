package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/amorindev/go-tmpl/mocks/app/users"
	"github.com/amorindev/go-tmpl/pkg/app/auth-methods/domain"
	"github.com/amorindev/go-tmpl/pkg/app/auth-methods/service"
	userD "github.com/amorindev/go-tmpl/pkg/app/users/domain"
	sharedD "github.com/amorindev/go-tmpl/pkg/shared/domain"
	sharedDomain "github.com/amorindev/go-tmpl/pkg/shared/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestService_SignUp(t *testing.T) {
	ctx := context.TODO()

	testTable := map[string]struct {
		setup         func(mockRepo *users.MockUserRepo)
		ctx           context.Context
		user          userD.User
		assertionFunc func(subTest *testing.T, gotUser *userD.User, gotErr error)
	}{
		"generic error in ExistsByEmail": {
			setup: func(mockRepo *users.MockUserRepo) {
				mockRepo.EXPECT().ExistsByEmail(ctx, "test@gmail.com").Return(false, errors.New("generic error"))
			},
			ctx: ctx,
			user: userD.User{
				Email: "test@gmail.com",
			},
			assertionFunc: func(subTest *testing.T, gotUser *userD.User, gotErr error) {
				assert.NotNil(subTest, gotErr)
				// ? el usuario debe ner nil revisa asi para todos los tes y sign up

				var appErr sharedD.AppError
				if !errors.As(gotErr, &appErr) {
					t.Fatalf("except AppError, got %T", gotErr)
				}
				assert.Equal(subTest, sharedD.ErrCodeInternalServerError, appErr.Code)
				assert.Equal(subTest, "Server Error", appErr.Msg)
			},
		},
		"success": {
			setup: func(mockRepo *users.MockUserRepo) {
				mockRepo.EXPECT().ExistsByEmail(ctx, gomock.Any()).Return(false, nil)

				mockRepo.EXPECT().Insert(ctx, gomock.Any()).Return(nil)
			},
			ctx: ctx,
			user: userD.User{
				ID: "123",
				UserPassAuth: &domain.UserPasswordAuth{
					Password: "pass123",
				},
			},
			assertionFunc: func(subTest *testing.T, gotUser *userD.User, gotErr error) {
				//fmt.Printf("Userid: %v\n", user.ID)
				//fmt.Printf("Userid2: %v\n", user.ID)
				assert.Nil(subTest, gotErr)
				assert.NotNil(subTest, gotUser.ID)
			},
		},
		"email already in use": {
			setup: func(mockRepo *users.MockUserRepo) {
				mockRepo.EXPECT().ExistsByEmail(gomock.Any(), gomock.Any()).Return(true, nil)
			},
			user: userD.User{},
			ctx:  nil,
			assertionFunc: func(subTest *testing.T, gotUser *userD.User, gotErr error) {
				assert.NotNil(subTest, gotErr)

				// ! asi o poner todo el texto
				expectErr := sharedDomain.ManageError(sharedDomain.ErrDuplicateKey, "email already in use")

				var appErr sharedD.AppError
				if !errors.As(gotErr, &appErr) {
					t.Fatalf("except AppError, got %T", gotErr)
				}

				assert.Equal(subTest, expectErr.Error(), appErr.Error())
			},
		},
	}

	for testName, test := range testTable {
		t.Run(testName, func(subTest *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mocksUserRepo := users.NewMockUserRepo(ctrl)

			s := service.Service{
				UserRepo: mocksUserRepo,
			}

			test.setup(mocksUserRepo)

			gotErr := s.SignUp(test.ctx, &test.user)

			test.assertionFunc(subTest, &test.user, gotErr)
		})
	}
}
