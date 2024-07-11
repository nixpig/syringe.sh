package cli

import (
	"crypto/rsa"
	"io"
	"strings"

	"github.com/nixpig/syringe.sh/pkg/ssh"
)

type ListResponseParser struct {
	w          io.Writer
	privateKey *rsa.PrivateKey
}

func (lrp ListResponseParser) Write(p []byte) (int, error) {
	cypherText := string(p)
	var err error

	lines := strings.Split(cypherText, "\n")
	for i, l := range lines {
		parts := strings.SplitN(l, "=", 2)
		parts[1], err = ssh.Decrypt(parts[1], lrp.privateKey)
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
}

func (grp GetResponseParser) Write(p []byte) (int, error) {
	decrypted, err := ssh.Decrypt(string(p), grp.privateKey)
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
