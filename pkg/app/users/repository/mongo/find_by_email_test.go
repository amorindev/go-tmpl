package mongo_test

import (
	"context"
	"fmt"
	"testing"

	mongoC "github.com/amorindev/go-tmpl/internal/mongo"
	"github.com/amorindev/go-tmpl/pkg/app/users/domain"
	"github.com/amorindev/go-tmpl/pkg/app/users/repository/mongo"
	sharedD "github.com/amorindev/go-tmpl/pkg/shared/domain"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestRepository_FindByEmail(t *testing.T) {
	// Skip this test if the short flag is provided
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	testTable := map[string]struct {
		Email         string
		assertionFunc func(subTest *testing.T, u *domain.User, err error)
	}{
		"should return user by email": {
			Email: "user@yopmail.com",
			assertionFunc: func(subTest *testing.T, u *domain.User, err error) {
				assert.Nil(subTest, err)
				assert.NotNil(subTest, u)
				assert.NotEmpty(subTest, u.ID)
			},
		},
		"should return a null user and an error of type ErrNotFound": {
			Email: "user2@yopmail.com",
			assertionFunc: func(subTest *testing.T, u *domain.User, err error) {
				assert.Nil(subTest, u)
				assert.NotNil(subTest, err)
				assert.ErrorIs(subTest, err, sharedD.ErrNotFound)
			},
		},
		"should fail if ID is not a bson.ObjectID": {
			Email: "invalidid@yopmail.com",
			assertionFunc: func(subTest *testing.T, u *domain.User, err error) {
				assert.Nil(subTest, u)
				assert.NotNil(subTest, err)
				assert.ErrorContains(subTest, err, "ID is not a bson.ObjectID")
			},
		},
		// ! el otro seria que te retorne el tipo string no object, esta parte seria insertando un id no valido
		//
	}

	db, err := startMongoDB(context.Background())
	if err != nil {
		t.Fatalf("error starting mongodb container: %v", err)
	}

	defer db.Container.Terminate(context.Background())

	repo, err := setupRepository(db)
	if err != nil {
		t.Fatalf("error setting up repository: %v", err)
	}

	// ! ver como pasar diferentes user o no se
	myUser := &domain.User{
		// ! de momento no estamos probando el id
		ID:    bson.NewObjectID(),
		Email: "user@yopmail.com",
	}

	if err = addUserToCollection(repo, myUser); err != nil {
		t.Fatal("error adding user to collection: %w", err)
	}

	// user with incorrect ID type (string en vez de bson.ObjectID)
	invalidIDEmail := "invalidid@yopmail.com"
	_, err = repo.Collection.InsertOne(
		context.Background(),
		bson.M{
			"_id": "invalid-bson-id", // we force the wrong type
			"email": invalidIDEmail,
		},
	)
	if err != nil {
		t.Fatalf("error adding invalid user: %v", err)
	}

	for testName, test := range testTable {
		t.Run(testName, func(subTest *testing.T) {
			user, err := repo.FindByEmail(context.Background(), test.Email)
			if err != nil {
				test.assertionFunc(subTest, user, err)
			}
		})
	}

}

func addUserToCollection(repo *mongo.Repository, u *domain.User) error {
	_, err := repo.Collection.InsertOne(context.Background(), u)
	if err != nil {
		return err
	}
	return nil
}

func setupRepository(db *mongodb.MongoDBContainer) (*mongo.Repository, error) {
	connStr, err := db.ConnectionString(context.Background())
	if err != nil {
		return nil, err
	}

	client, err := mongoC.GetConnection(connStr)
	if err != nil {
		return nil, err
	}

	userRepo := &mongo.Repository{
		Client:     client,
		Collection: client.Database("auth-tmpl").Collection("users"),
	}

	if err = userRepo.CreateIndexes(); err != nil {
		return nil, err
	}

	return userRepo, nil
}

// startMongoDB will return a mongodb testcontainer instance or an error
func startMongoDB(ctx context.Context) (*mongodb.MongoDBContainer, error) {
	mongodbContainer, err := mongodb.RunContainer(ctx)
	if err != nil {
		return nil, fmt.Errorf("error running mongodb container: %w", err)
	}
	return mongodbContainer, nil
}
