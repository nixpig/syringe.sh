package services

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var mockAppStore = new(MockAppStore)

func TestAppService(t *testing.T) {
	scenarios := map[string]func(t *testing.T, service AppService){
		"test app service register user (success)":                 testAppServiceRegisterUserSuccess,
		"test app service register user (field validation error)":  testAppServiceRegisterUserFieldValidationError,
		"test app service register user (insert user store error)": testAppServiceRegisterUserInsertUserStoreError,
		"test app service register user (insert key store error)":  testAppServiceRegisterUserInsertKeyStoreError,

		"test app service add public key (success)":                testAppServiceAddPublicKeySuccess,
		"test app service add public key (field validation error)": testAppServiceAddPublicKeyFieldValidationError,
		"test app service add public key (key store error)":        testAppServiceAddPublicKeyKeyStoreError,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			validate := validator.New()
			service := NewAppServiceImpl(mockAppStore, validate)

			fn(t, service)
		})
	}
}

type MockAppStore struct {
	mock.Mock
}

func (m *MockAppStore) InsertUser(username, email, publicKey string) (*models.User, error) {
	args := m.Called(username, email, publicKey)

	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAppStore) UpdateUser(user models.User) (*models.User, error) {
	args := m.Called(user)

	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAppStore) DeleteUserByUsername(username string) error {
	args := m.Called(username)

	return args.Error(0)
}

func (m *MockAppStore) GetUserByUsername(username string) (*models.User, error) {
	args := m.Called(username)

	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAppStore) InsertKey(userId int, publicKey string) (*models.Key, error) {
	args := m.Called(userId, publicKey)

	return args.Get(0).(*models.Key), args.Error(1)
}

func (m *MockAppStore) InsertDatabase(name string, userId int) (*models.Database, error) {
	args := m.Called(name, userId)

	return args.Get(0).(*models.Database), args.Error(1)
}

func testAppServiceRegisterUserSuccess(t *testing.T, service AppService) {
	createdAt := "2024-06-05 05:29:16"

	mockAppStoreInsertUser := mockAppStore.
		On("InsertUser", "janedoe", "jane@example.org", "active").
		Return(&models.User{
			Id:        23,
			Username:  "janedoe",
			Email:     "jane@example.org",
			CreatedAt: createdAt,
		}, nil)

	mockAppStoreInsertKey := mockAppStore.
		On("InsertKey", 23, "some_public_key").
		Return(&models.Key{
			Id:        42,
			PublicKey: "some_public_key",
			UserId:    23,
			CreatedAt: createdAt,
		}, nil)

	registeredUser, err := service.RegisterUser(RegisterUserRequest{
		Username:  "janedoe",
		Email:     "jane@example.org",
		PublicKey: "some_public_key",
	})

	require.NoError(t, err, "should not return error")
	require.Equal(t, &RegisterUserResponse{
		Id:        23,
		Username:  "janedoe",
		Email:     "jane@example.org",
		CreatedAt: createdAt,
		PublicKey: "some_public_key",
	}, registeredUser, "should return user details response")

	mockAppStoreInsertUser.Unset()
	mockAppStoreInsertKey.Unset()

	if expected := mockAppStore.AssertExpectations(t); !expected {
		t.Error("did not call store as expected")
	}
}

func testAppServiceRegisterUserFieldValidationError(t *testing.T, service AppService) {
	var err error

	usernameMinLength, err := service.RegisterUser(RegisterUserRequest{
		Username:  "ja",
		Email:     "jane@example.org",
		PublicKey: "some_public_key",
	})
	require.Empty(t, usernameMinLength)
	require.Error(t, err, "should return validation error")

	usernameMaxLength, err := service.RegisterUser(RegisterUserRequest{
		Username:  "janedoejanedoejanedoejanedoejanedoe",
		Email:     "jane@example.org",
		PublicKey: "some_public_key",
	})
	require.Empty(t, usernameMaxLength)
	require.Error(t, err, "should return validation error")

	emailInvalid, err := service.RegisterUser(RegisterUserRequest{
		Username:  "janedoe",
		Email:     "janeexampleorg",
		PublicKey: "some_public_key",
	})
	require.Empty(t, emailInvalid)
	require.Error(t, err, "should return validation error")

	missingUsername, err := service.RegisterUser(RegisterUserRequest{
		Email:     "jane@example.org",
		PublicKey: "some_public_key",
	})
	require.Empty(t, missingUsername)
	require.Error(t, err, "should return validation error")

	missingEmail, err := service.RegisterUser(RegisterUserRequest{
		Username:  "janedoe",
		PublicKey: "some_public_key",
	})
	require.Empty(t, missingEmail)
	require.Error(t, err, "should return validation error")

	missingPublicKey, err := service.RegisterUser(RegisterUserRequest{
		Username: "janedoe",
		Email:    "jane@example.org",
	})
	require.Empty(t, missingPublicKey)
	require.Error(t, err, "should return validation error")

	if mockAppStoreNotCalled := mockAppStore.AssertNotCalled(t, "Insert"); !mockAppStoreNotCalled {
		t.Error("expected user store not to be called")
	}
}

func testAppServiceRegisterUserInsertUserStoreError(t *testing.T, service AppService) {
	mockAppStoreInsert := mockAppStore.
		On("InsertUser", "janedoe", "jane@example.org", "active").
		Return(&models.User{}, errors.New("store_insert_error"))

	registeredUser, err := service.RegisterUser(RegisterUserRequest{
		Username:  "janedoe",
		Email:     "jane@example.org",
		PublicKey: "some_public_key",
	})

	require.Empty(t, registeredUser)
	require.EqualError(t, err, "store_insert_error")

	mockAppStoreInsert.Unset()
	if expected := mockAppStore.AssertExpectations(t); !expected {
		t.Error("did not call user store as expected")
	}
}

func testAppServiceRegisterUserInsertKeyStoreError(t *testing.T, service AppService) {
	createdAt := "2024-06-05 05:29:16"

	mockAppStoreInsertUser := mockAppStore.
		On("InsertUser", "janedoe", "jane@example.org", "active").
		Return(&models.User{
			Id:        23,
			Username:  "janedoe",
			Email:     "jane@example.org",
			CreatedAt: createdAt,
		}, nil)

	mockAppStoreInsertKey := mockAppStore.
		On("InsertKey", 23, "some_public_key").
		Return(&models.Key{}, errors.New("key_store_insert_error"))

	registeredUser, err := service.RegisterUser(RegisterUserRequest{
		Username:  "janedoe",
		Email:     "jane@example.org",
		PublicKey: "some_public_key",
	})

	require.EqualError(t, err, "key_store_insert_error", "should return key store error")
	require.Empty(t, registeredUser, "should not return user details")

	mockAppStoreInsertUser.Unset()
	mockAppStoreInsertKey.Unset()

	if expected := mockAppStore.AssertExpectations(t); !expected {
		t.Error("did not call store as expected")
	}
}

func testAppServiceAddPublicKeySuccess(t *testing.T, service AppService) {
	createdAt := "2024-06-05 05:29:16"

	mockAppStoreInsertKey := mockAppStore.
		On("InsertKey", 23, "some_public_key").
		Return(&models.Key{
			Id:        42,
			PublicKey: "some_public_key",
			UserId:    23,
			CreatedAt: createdAt,
		}, nil)

	addedPublicKey, err := service.AddPublicKey(AddPublicKeyRequest{
		UserId:    23,
		PublicKey: "some_public_key",
	})

	require.NoError(t, err, "should not return error")
	require.Equal(t, &AddPublicKeyResponse{
		Id:        42,
		UserId:    23,
		CreatedAt: createdAt,
		PublicKey: "some_public_key",
	}, addedPublicKey, "should return added key details response")

	mockAppStoreInsertKey.Unset()

	if expected := mockAppStore.AssertExpectations(t); !expected {
		t.Error("did not call store as expected")
	}
}

func testAppServiceAddPublicKeyFieldValidationError(t *testing.T, service AppService) {
	missingUserId, err := service.AddPublicKey(AddPublicKeyRequest{
		PublicKey: "some_public_key",
	})
	require.Empty(t, missingUserId, "should return empty result")
	require.Error(t, err, "should return validation error")

	missingPublicKey, err := service.AddPublicKey(AddPublicKeyRequest{
		UserId: 23,
	})
	require.Empty(t, missingPublicKey, "should return empty result")
	require.Error(t, err, "should return validation error")
}

func testAppServiceAddPublicKeyKeyStoreError(t *testing.T, service AppService) {
	mockAppStoreInsertKey := mockAppStore.
		On("InsertKey", 23, "some_public_key").
		Return(&models.Key{}, errors.New("key_store_error"))

	addedPublicKey, err := service.AddPublicKey(AddPublicKeyRequest{
		UserId:    23,
		PublicKey: "some_public_key",
	})

	require.EqualError(t, err, "key_store_error", "should return key store error")
	require.Empty(t, addedPublicKey, "should return empty key details response")

	mockAppStoreInsertKey.Unset()

	if expected := mockAppStore.AssertExpectations(t); !expected {
		t.Error("did not call store as expected")
	}
}
