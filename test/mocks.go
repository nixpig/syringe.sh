package test

import (
	"crypto/rsa"
	"net/http"

	"github.com/nixpig/syringe.sh/pkg/turso"
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

type MockTursoClient struct {
	mock.Mock
}

func (m *MockTursoClient) CreateDatabase(name, group string) (*turso.TursoDatabaseResponse, error) {
	args := m.Called(name, group)

	return args.Get(0).(*turso.TursoDatabaseResponse), args.Error(1)
}

func (m *MockTursoClient) ListDatabases() (*turso.TursoDatabases, error) {
	args := m.Called()

	return args.Get(0).(*turso.TursoDatabases), args.Error(1)
}

func (m *MockTursoClient) CreateToken(name, expiration string) (*turso.TursoToken, error) {
	args := m.Called(name, expiration)

	return args.Get(0).(*turso.TursoToken), args.Error(1)
}

func (m *MockTursoClient) New(organization, apiToken, baseURL string, httpClient http.Client) turso.TursoDatabaseAPI {
	args := m.Called(organization, apiToken, httpClient)

	return args.Get(0).(*MockTursoClient)
}
