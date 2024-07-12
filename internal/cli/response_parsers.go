package cli

import (
	"crypto/rsa"
	"io"
	"strings"

	"github.com/nixpig/syringe.sh/pkg/ssh"
)

type Decryptor func(cypherText string, privateKey *rsa.PrivateKey) (string, error)

type ListResponseParser struct {
	w          io.Writer
	privateKey *rsa.PrivateKey
	decrypt    Decryptor
}

func NewListResponseParser(w io.Writer, privateKey *rsa.PrivateKey, decrypt Decryptor) ListResponseParser {
	return ListResponseParser{
		w:          w,
		privateKey: privateKey,
		decrypt:    decrypt,
	}
}

func (lrp ListResponseParser) Write(p []byte) (int, error) {
	cypherText := string(p)
	var err error

	lines := strings.Split(cypherText, "\n")
	for i, l := range lines {
		parts := strings.SplitN(l, "=", 2)
		parts[1], err = lrp.decrypt(parts[1], lrp.privateKey)
		if err != nil {
			return 0, err
		}

		lines[i] = strings.Join(parts, "=")
	}

	return lrp.w.Write([]byte(strings.Join(lines, "\n")))
}

type GetResponseParser struct {
	w          io.Writer
	privateKey *rsa.PrivateKey
	decrypt    Decryptor
}

func NewGetResponseParser(w io.Writer, privateKey *rsa.PrivateKey, decrypt Decryptor) GetResponseParser {
	return GetResponseParser{
		w:          w,
		privateKey: privateKey,
		decrypt:    decrypt,
	}
}

func (grp GetResponseParser) Write(p []byte) (int, error) {
	decrypted, err := grp.decrypt(string(p), grp.privateKey)
	if err != nil {
		return 0, err
	}

	return grp.w.Write([]byte(decrypted))
}

type InjectResponseParser struct {
	w          io.Writer
	privateKey *rsa.PrivateKey
}

func (irp InjectResponseParser) Write(p []byte) (int, error) {
	var err error

	cypherText := string(p)

	lines := strings.Split(cypherText, " ")
	for i, l := range lines {
		parts := strings.SplitN(l, "=", 2)
		parts[1], err = ssh.Decrypt(parts[1], irp.privateKey)
		if err != nil {
			return 0, err
		}

		lines[i] = strings.Join(parts, "=")
	}

	return irp.w.Write([]byte(strings.Join(lines, " ")))
}
