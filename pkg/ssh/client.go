package ssh

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

// NewClient dials the remote host described by config and returns an
// authenticated [ssh.Client].
//
// It calls [GetConfigPEM] to build the underlying [ssh.ClientConfig], then
// opens a TCP connection to config.Addr. The caller is responsible for closing
// the returned client when it is no longer needed.
//
// Example:
//
//	client, err := ssh.NewClient(&ssh.ClientConfig{
//	    Addr:           "bastion.example.com:22",
//	    User:           "deploy",
//	    Timeout:        10 * time.Second,
//	    PrivateKeyPath: "/home/user/.ssh/id_ed25519",
//	    KnownHostsPath: "/home/user/.ssh/known_hosts",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
func NewClient(config *ClientConfig) (*ssh.Client, error) {
	sshConfig, err := GetConfigPEM(config)
	if err != nil {
		return nil, err
	}

	client, err := ssh.Dial("tcp", config.Addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to connect: %w", err)
	}

	return client, nil
}
