package cli_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/nixpig/syringe.sh/internal/cli"
	"github.com/nixpig/syringe.sh/test"
	"github.com/stretchr/testify/require"
)

var mockCrypt = new(test.MockCrypt)

func TestResponseParsers(t *testing.T) {
	scenarios := map[string]func(t *testing.T){
		"test list response parser single secret happy path":    testListResponseParserSingleSecretHappyPath,
		"test list response parser multiple secrets happy path": testListResponseParserMultipleSecretsHappyPath,
		"test list response parser decrypt error":               testListResponseParserDecryptError,

		"test get response parser happy path":    testGetResponseParserHappyPath,
		"test get response parser decrypt error": testGetResponseParserDecryptError,

		"test inject response parser single secret happy path":    testInjectResponseParserSingleSecretHappyPath,
		"test inject response parser multiple secrets happy path": testInjectResponseParserMultipleSecretsHappyPath,
		"test inject response parser decrypt error":               testInjectResponseParserDecryptError,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}
}

func testListResponseParserSingleSecretHappyPath(t *testing.T) {
	w := bytes.NewBufferString("")
	_, privateKey, err := test.GenerateKeyPair()
	require.NoError(t, err)

	parser := cli.NewListResponseParser(
		w,
		privateKey,
		mockCrypt.Decrypt,
	)

	mockCrypt.
		On("Decrypt", "mock_encrypted_value", privateKey).
		Return("mock_decrypted_value", nil)

	b, err := parser.Write([]byte("mock_key=mock_encrypted_value"))
	require.NoError(t, err)
	require.Equal(t, 29, b)

	written, err := io.ReadAll(w)
	require.NoError(t, err)

	require.Equal(t, "mock_key=mock_decrypted_value", string(written))
}

func testListResponseParserMultipleSecretsHappyPath(t *testing.T) {
	w := bytes.NewBufferString("")
	_, privateKey, err := test.GenerateKeyPair()
	require.NoError(t, err)

	parser := cli.NewListResponseParser(
		w,
		privateKey,
		mockCrypt.Decrypt,
	)

	mockCrypt.
		On("Decrypt", "mock_encrypted_value_1", privateKey).
		Return("mock_decrypted_value_1", nil).
		On("Decrypt", "mock_encrypted_value_2", privateKey).
		Return("mock_decrypted_value_2", nil)

	b, err := parser.Write([]byte("mock_key_1=mock_encrypted_value_1\nmock_key_2=mock_encrypted_value_2"))
	require.NoError(t, err)
	require.Equal(t, 67, b)

	written, err := io.ReadAll(w)
	require.NoError(t, err)

	require.Equal(t, "mock_key_1=mock_decrypted_value_1\nmock_key_2=mock_decrypted_value_2", string(written))
}

func testListResponseParserDecryptError(t *testing.T) {
	w := bytes.NewBufferString("")
	_, privateKey, err := test.GenerateKeyPair()
	require.NoError(t, err)

	parser := cli.NewListResponseParser(
		w,
		privateKey,
		mockCrypt.Decrypt,
	)

	mockCrypt.
		On("Decrypt", "mock_encrypted_value", privateKey).
		Return("", errors.New("decrypt_error"))

	b, err := parser.Write([]byte("mock_key=mock_encrypted_value"))
	require.EqualError(t, err, "decrypt_error")
	require.Equal(t, 0, b)

	written, err := io.ReadAll(w)
	require.NoError(t, err)
	require.Equal(t, "", string(written))
}

func testGetResponseParserHappyPath(t *testing.T) {
	w := bytes.NewBufferString("")
	_, privateKey, err := test.GenerateKeyPair()
	require.NoError(t, err)

	parser := cli.NewGetResponseParser(
		w,
		privateKey,
		mockCrypt.Decrypt,
	)

	mockCrypt.
		On("Decrypt", "mock_encrypted_value", privateKey).
		Return("mock_decrypted_value", nil)

	b, err := parser.Write([]byte("mock_encrypted_value"))
	require.NoError(t, err)
	require.Equal(t, 20, b)

	written, err := io.ReadAll(w)
	require.NoError(t, err)

	require.Equal(t, "mock_decrypted_value", string(written))
}

func testGetResponseParserDecryptError(t *testing.T) {
	w := bytes.NewBufferString("")
	_, privateKey, err := test.GenerateKeyPair()
	require.NoError(t, err)

	parser := cli.NewGetResponseParser(
		w,
		privateKey,
		mockCrypt.Decrypt,
	)

	mockCrypt.
		On("Decrypt", "mock_encrypted_value", privateKey).
		Return("", errors.New("decrypt_error"))

	b, err := parser.Write([]byte("mock_encrypted_value"))
	require.EqualError(t, err, "decrypt_error")
	require.Equal(t, 0, b)

	written, err := io.ReadAll(w)
	require.NoError(t, err)

	require.Equal(t, "", string(written))
}

func testInjectResponseParserSingleSecretHappyPath(t *testing.T) {
	w := bytes.NewBufferString("")
	_, privateKey, err := test.GenerateKeyPair()
	require.NoError(t, err)

	parser := cli.NewInjectResponseParser(
		w,
		privateKey,
		mockCrypt.Decrypt,
	)

	mockCrypt.
		On("Decrypt", "mock_encrypted_value", privateKey).
		Return("mock_decrypted_value", nil)

	b, err := parser.Write([]byte("mock_key=mock_encrypted_value"))
	require.NoError(t, err)
	require.Equal(t, 29, b)

	written, err := io.ReadAll(w)
	require.NoError(t, err)

	require.Equal(t, "mock_key=mock_decrypted_value", string(written))
}

func testInjectResponseParserMultipleSecretsHappyPath(t *testing.T) {
	w := bytes.NewBufferString("")
	_, privateKey, err := test.GenerateKeyPair()
	require.NoError(t, err)

	parser := cli.NewInjectResponseParser(
		w,
		privateKey,
		mockCrypt.Decrypt,
	)

	mockCrypt.
		On("Decrypt", "mock_encrypted_value_1", privateKey).
		Return("mock_decrypted_value_1", nil).
		On("Decrypt", "mock_encrypted_value_2", privateKey).
		Return("mock_decrypted_value_2", nil)

	b, err := parser.Write([]byte("mock_key_1=mock_encrypted_value_1 mock_key_2=mock_encrypted_value_2"))
	require.NoError(t, err)
	require.Equal(t, 67, b)

	written, err := io.ReadAll(w)
	require.NoError(t, err)

	require.Equal(t, "mock_key_1=mock_decrypted_value_1 mock_key_2=mock_decrypted_value_2", string(written))
}

func testInjectResponseParserDecryptError(t *testing.T) {
	w := bytes.NewBufferString("")
	_, privateKey, err := test.GenerateKeyPair()
	require.NoError(t, err)

	parser := cli.NewInjectResponseParser(
		w,
		privateKey,
		mockCrypt.Decrypt,
	)

	mockCrypt.
		On("Decrypt", "mock_encrypted_value", privateKey).
		Return("", errors.New("decrypt_error"))

	b, err := parser.Write([]byte("mock_key=mock_encrypted_value"))
	require.EqualError(t, err, "decrypt_error")
	require.Equal(t, 0, b)

	written, err := io.ReadAll(w)
	require.NoError(t, err)
	require.Equal(t, "", string(written))
}
