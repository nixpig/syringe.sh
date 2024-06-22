package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func main() {
	keyFile, err := os.Open("/home/nixpig/.ssh/id_rsa")
	if err != nil {
		fmt.Println("unable to open file: ", err)
		os.Exit(1)
	}

	keyContents, err := io.ReadAll(keyFile)
	if err != nil {
		fmt.Println("unable to read file contents: ", err)
		os.Exit(1)
	}

	fmt.Print("Enter password for private key: ")

	pass, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("unable to read in password from stdin: ", err)
		os.Exit(1)
	}

	signer, err := gossh.ParsePrivateKeyWithPassphrase(keyContents, pass)
	if err != nil {
		fmt.Println("unable to parse private key: ", err)
		os.Exit(1)
	}

	config := &gossh.ClientConfig{
		User: "nixpig",
		Auth: []gossh.AuthMethod{
			gossh.PublicKeys(signer),
		},
		HostKeyCallback: gossh.HostKeyCallback(func(hostname string, remote net.Addr, key gossh.PublicKey) error {
			fmt.Println("in callback...")
			fmt.Println("hostname: ", hostname)
			fmt.Println("remote: ", remote)

			// publicKey, err := gossh.ParsePublicKey(key.Marshal())
			// if err != nil {
			// 	fmt.Println("unable to parse public key: ", err)
			// 	os.Exit(1)
			// }

			fmt.Println("key: ", string(key.Marshal()))
			return nil
		}),
	}

	conn, err := gossh.Dial("tcp", "localhost:23234", config)
	if err != nil {
		fmt.Println("unable to dial: ", err)
		os.Exit(1)
	}

	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		fmt.Println("unable to create new session: ", err)
		os.Exit(1)
	}

	defer session.Close()

	// stdin, err := session.StdinPipe()
	// if err != nil {
	// 	fmt.Println("unable to get stdin pipe: ", err)
	// 	os.Exit(1)
	// }

	stdout, err := session.StdoutPipe()
	if err != nil {
		fmt.Println("unable to get stdout pipe: ", err)
		os.Exit(1)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		fmt.Println("unable to get stderr pipe: ", err)
		os.Exit(1)
	}

	go func() {
		scanner := bufio.NewScanner(stdout)

		for {
			if tkn := scanner.Scan(); tkn {
				rcv := scanner.Bytes()

				raw := make([]byte, len(rcv))

				copy(raw, rcv)

				fmt.Println("raw: ", string(raw))
			} else {
				if scanner.Err() != nil {
					fmt.Println("err: ", scanner.Err())
				} else {
					fmt.Println("io.EOF")
				}

				return
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)

		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	if err := session.Run("project -meh"); err != nil {
		fmt.Println("unable to run command on remote connection: ", err)
		os.Exit(1)
	}
}
