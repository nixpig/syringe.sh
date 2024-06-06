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
		"test user service create user (success)":                testAppServiceCreateUserSuccess,
		"test user service create user (field validation error)": testAppServiceCreateUserFieldValidationError,
		"test user service create user (store error)":            testAppServiceCreateAppStoreError,
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

func testAppServiceCreateUserSuccess(t *testing.T, service AppService) {
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
		On("InsertKey", 23, "p4ssw0rd").
		Return(&models.Key{
			Id:        42,
			PublicKey: "p4ssw0rd",
			UserId:    23,
			CreatedAt: createdAt,
		}, nil)

	createdUser, err := service.RegisterUser(RegisterUserRequestDto{
		Username:  "janedoe",
		Email:     "jane@example.org",
		PublicKey: "p4ssw0rd",
	})

	require.NoError(t, err, "should not return error")
	require.Equal(t, &RegisterUserResponseDto{
		Id:        23,
		Username:  "janedoe",
		Email:     "jane@example.org",
		CreatedAt: createdAt,
		PublicKey: "p4ssw0rd",
	}, createdUser, "should return user details response")

	mockAppStoreInsertUser.Unset()
	mockAppStoreInsertKey.Unset()

	if expected := mockAppStore.AssertExpectations(t); !expected {
		t.Error("did not call store as expected")
	}
}

func testAppServiceCreateUserFieldValidationError(t *testing.T, service AppService) {
	var err error

	usernameMinLength, err := service.RegisterUser(RegisterUserRequestDto{
		Username:  "ja",
		Email:     "jane@example.org",
		PublicKey: "p4ssw0rd",
	})
	require.Empty(t, usernameMinLength)
	require.Error(t, err, "should return validation error")

	usernameMaxLength, err := service.RegisterUser(RegisterUserRequestDto{
		Username:  "janedoejanedoejanedoejanedoejanedoe",
		Email:     "jane@example.org",
		PublicKey: "p4ssw0rd",
	})
	require.Empty(t, usernameMaxLength)
	require.Error(t, err, "should return validation error")

	emailInvalid, err := service.RegisterUser(RegisterUserRequestDto{
		Username:  "janedoe",
		Email:     "janeexampleorg",
		PublicKey: "p4ssw0rd",
	})
	require.Empty(t, emailInvalid)
	require.Error(t, err, "should return validation error")

	missingUsername, err := service.RegisterUser(RegisterUserRequestDto{
		Email:     "jane@example.org",
		PublicKey: "p4ssw0rd",
	})
	require.Empty(t, missingUsername)
	require.Error(t, err, "should return validation error")

	missingEmail, err := service.RegisterUser(RegisterUserRequestDto{
		Username:  "janedoe",
		PublicKey: "p4ssw0rd",
	})
	require.Empty(t, missingEmail)
	require.Error(t, err, "should return validation error")

	missingPublicKey, err := service.RegisterUser(RegisterUserRequestDto{
		Username: "janedoe",
		Email:    "jane@example.org",
	})
	require.Empty(t, missingPublicKey)
	require.Error(t, err, "should return validation error")

	if mockAppStoreNotCalled := mockAppStore.AssertNotCalled(t, "Insert"); !mockAppStoreNotCalled {
		t.Error("expected user store not to be called")
	}
}

func testAppServiceCreateAppStoreError(t *testing.T, service AppService) {
	mockAppStoreInsert := mockAppStore.
		On("InsertUser", "janedoe", "jane@example.org", "active").
		Return(&models.User{}, errors.New("store_insert_error"))

	createdUser, err := service.RegisterUser(RegisterUserRequestDto{
		Username:  "janedoe",
		Email:     "jane@example.org",
		PublicKey: "p4ssw0rd",
	})

	require.Empty(t, createdUser)
	require.EqualError(t, err, "store_insert_error")

	mockAppStoreInsert.Unset()
	if expected := mockAppStore.AssertExpectations(t); !expected {
		t.Error("did not call user store as expected")
	}
}
