package services

import (
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var mockUserStore = new(MockUserStore)

func TestJsonUserService(t *testing.T) {
	scenarios := map[string]func(t *testing.T, service UserService){
		"test json user service create user (success)":                testJsonUserServiceCreateUserSuccess,
		"test json user service create user (field validation error)": testJsonUserServiceCreateUserFieldValidationError,
		"test json user service create user (store error)":            testJsonUserServiceCreateUserStoreError,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			validate := validator.New()
			service := NewJsonUserService(mockUserStore, validate)

			fn(t, service)
		})
	}
}

type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) Insert(username, email, publicKey string) (*User, error) {
	args := m.Called(username, email, publicKey)

	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserStore) Update(user User) (*User, error) {
	args := m.Called(user)

	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserStore) DeleteByUsername(username string) error {
	args := m.Called(username)

	return args.Error(0)
}

func (m *MockUserStore) GetByUsername(username string) (*User, error) {
	args := m.Called(username)

	return args.Get(0).(*User), args.Error(1)
}

func testJsonUserServiceCreateUserSuccess(t *testing.T, service UserService) {
	createdAt := time.Now()

	mockUserStoreCreate := mockUserStore.
		On("Insert", "janedoe", "jane@example.org", "p4ssw0rd").
		Return(&User{
			Id:        23,
			Username:  "janedoe",
			Email:     "jane@example.org",
			CreatedAt: createdAt,
		}, nil)

	createdUser, err := service.Create(RegisterUserRequestDto{
		Username:  "janedoe",
		Email:     "jane@example.org",
		PublicKey: "p4ssw0rd",
	})

	require.NoError(t, err, "should not return error")
	require.Equal(t, &UserDetailsResponseDto{
		Id:        23,
		Username:  "janedoe",
		Email:     "jane@example.org",
		CreatedAt: createdAt,
	}, createdUser, "should return user details response")

	mockUserStoreCreate.Unset()
	if expected := mockUserStore.AssertExpectations(t); !expected {
		t.Error("did not call store as expected")
	}
}

func testJsonUserServiceCreateUserFieldValidationError(t *testing.T, service UserService) {
	var err error

	usernameMinLength, err := service.Create(RegisterUserRequestDto{
		Username:  "ja",
		Email:     "jane@example.org",
		PublicKey: "p4ssw0rd",
	})
	require.Empty(t, usernameMinLength)
	require.Error(t, err, "should return validation error")

	usernameMaxLength, err := service.Create(RegisterUserRequestDto{
		Username:  "janedoejanedoejanedoejanedoejanedoe",
		Email:     "jane@example.org",
		PublicKey: "p4ssw0rd",
	})
	require.Empty(t, usernameMaxLength)
	require.Error(t, err, "should return validation error")

	emailInvalid, err := service.Create(RegisterUserRequestDto{
		Username:  "janedoe",
		Email:     "janeexampleorg",
		PublicKey: "p4ssw0rd",
	})
	require.Empty(t, emailInvalid)
	require.Error(t, err, "should return validation error")

	missingUsername, err := service.Create(RegisterUserRequestDto{
		Email:     "jane@example.org",
		PublicKey: "p4ssw0rd",
	})
	require.Empty(t, missingUsername)
	require.Error(t, err, "should return validation error")

	missingEmail, err := service.Create(RegisterUserRequestDto{
		Username:  "janedoe",
		PublicKey: "p4ssw0rd",
	})
	require.Empty(t, missingEmail)
	require.Error(t, err, "should return validation error")

	missingPublicKey, err := service.Create(RegisterUserRequestDto{
		Username: "janedoe",
		Email:    "jane@example.org",
	})
	require.Empty(t, missingPublicKey)
	require.Error(t, err, "should return validation error")

	if mockUserStoreNotCalled := mockUserStore.AssertNotCalled(t, "Insert"); !mockUserStoreNotCalled {
		t.Error("expected user store not to be called")
	}
}

func testJsonUserServiceCreateUserStoreError(t *testing.T, service UserService) {
	mockUserStoreInsert := mockUserStore.
		On("Insert", "janedoe", "jane@example.org", "p4ssw0rd").
		Return(&User{}, errors.New("store_insert_error"))

	createdUser, err := service.Create(RegisterUserRequestDto{
		Username:  "janedoe",
		Email:     "jane@example.org",
		PublicKey: "p4ssw0rd",
	})

	require.Empty(t, createdUser)
	require.EqualError(t, err, "store_insert_error")

	mockUserStoreInsert.Unset()
	if expected := mockUserStore.AssertExpectations(t); !expected {
		t.Error("did not call user store as expected")
	}
}
