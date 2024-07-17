package test

import (
	"crypto/rsa"

	"github.com/stretchr/testify/mock"
)

type MockCrypt struct {
	mock.Mock
}

func (m *MockCrypt) Decrypt(cypherText string, privateKey *rsa.PrivateKey) (string, error) {
	args := m.Called(cypherText, privateKey)

	return args.String(0), args.Error(1)
}

type MockTerm struct {
	mock.Mock
}

func (mt *MockTerm) ReadPassword(fd int) ([]byte, error) {
	args := mt.Called(fd)

	return args.Get(0).([]byte), args.Error(1)
}
