package api

import (
	"fmt"
	"io"
	"os"

	"github.com/nixpig/syringe.sh/pkg/ssh"
)

type API interface {
	Register() error
	Set(key, value string) error
	Get(key string) error
	List() error
	Remove(key string) error
	SetOut(w io.Writer)
	Close() error
}

type hostAPI struct {
	// calls remote API over SSH
	client *ssh.SSHClient
	out    io.Writer
}

func New(client *ssh.SSHClient, out io.Writer) *hostAPI {
	return &hostAPI{
		client: client,
		out:    out,
	}
}

func (l *hostAPI) SetOut(w io.Writer) {
	l.out = w
}

func (l *hostAPI) Register() error {
	return l.client.Run("register", os.Stdout)
}

func (l *hostAPI) Set(key, value string) error {
	return l.client.Run(fmt.Sprintf("set %s %s", key, value), l.out)
}

func (l *hostAPI) Get(key string) error {
	return l.client.Run(fmt.Sprintf("get %s", key), l.out)
}

func (l *hostAPI) List() error {
	return l.client.Run("list", l.out)
}

func (l *hostAPI) Remove(key string) error {
	return l.client.Run(fmt.Sprintf("remove %s", key), l.out)
}

func (l *hostAPI) Close() error {
	l.client.Close()
	return nil
}
