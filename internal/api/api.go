package api

import (
	"fmt"
	"io"

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

type HostAPI struct {
	client *ssh.SSHClient
	out    io.Writer
}

func New(client *ssh.SSHClient, out io.Writer) *HostAPI {
	return &HostAPI{
		client: client,
		out:    out,
	}
}

func (l *HostAPI) SetOut(w io.Writer) {
	l.out = w
}

func (l *HostAPI) Register() error {
	return l.client.Run("register", l.out)
}

func (l *HostAPI) Set(key, value string) error {
	return l.client.Run(fmt.Sprintf("set %s %s", key, value), l.out)
}

func (l *HostAPI) Get(key string) error {
	return l.client.Run(fmt.Sprintf("get %s", key), l.out)
}

func (l *HostAPI) List() error {
	return l.client.Run("list", l.out)
}

func (l *HostAPI) Remove(key string) error {
	return l.client.Run(fmt.Sprintf("remove %s", key), l.out)
}

func (l *HostAPI) Close() error {
	l.client.Close()
	return nil
}
