package ssh

import (
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type SshClient struct {
	client *ssh.Client
}

type NewOptions struct {
	Addr       string
	User       string
	Password   string
	Key        string
	KeyPath    string
	KnownHosts []string
	Timeout    int
	Insecure   bool
}

func NewSshClient(options *NewOptions) (*SshClient, error) {
	auth := []ssh.AuthMethod{}
	if len(options.Key) > 0 {
		signer, err := ssh.ParsePrivateKey([]byte(options.Key))
		if err != nil {
			return nil, err
		}

		auth = append(auth, ssh.PublicKeys(signer))
	} else if len(options.KeyPath) > 0 {
		key, err := os.ReadFile(options.KeyPath)
		if err != nil {
			return nil, err
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}

		auth = append(auth, ssh.PublicKeys(signer))
	}
	if len(options.Password) > 0 {
		auth = append(auth, ssh.Password(options.Password))
	}

	client, err := ssh.Dial("tcp", options.Addr, &ssh.ClientConfig{
		User:    options.User,
		Auth:    auth,
		Timeout: time.Duration(options.Timeout) * time.Second,
		HostKeyCallback: (func() ssh.HostKeyCallback {
			if options.Insecure {
				return ssh.InsecureIgnoreHostKey()
			}
			if callback, err := knownhosts.New(options.KnownHosts...); err == nil {
				return callback
			} else {
				return nil
			}
		})(),
	})
	if err != nil {
		return nil, err
	}

	return &SshClient{client}, nil
}

type RunOptions struct {
	Command string
	Env     map[string]string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

func (s *SshClient) Run(options *RunOptions) error {
	session, err := s.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	for name, value := range options.Env {
		if err := session.Setenv(name, value); err != nil {
			return err
		}
	}

	session.Stdout = options.Stdout
	session.Stderr = options.Stderr

	writer, err := session.StdinPipe()
	if err != nil {
		return err
	}
	defer writer.Close()

	err = session.Shell()
	if err != nil {
		return err
	}

	cmds := []string{options.Command, "exit", "\x00"}
	for _, cmd := range cmds {
		fmt.Fprintf(writer, "%s\n", cmd)
	}

	return session.Wait()
}

func (s *SshClient) Close() error {
	return s.client.Close()
}
