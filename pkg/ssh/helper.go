package ssh

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
	kh "golang.org/x/crypto/ssh/knownhosts"
)

// GetConfigPEM builds an [ssh.ClientConfig] from the given [ClientConfig].
//
// It reads the unencrypted PEM-encoded private key at cfg.PrivateKeyPath,
// parses it into a signer, and loads cfg.KnownHostsPath for strict host-key
// validation. The returned config is ready to be passed directly to [ssh.Dial].
func GetConfigPEM(cfg *ClientConfig) (*ssh.ClientConfig, error) {
	key, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	hostKeyCallback, err := kh.New(cfg.KnownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("known_hosts load: %w", err)
	}

	return &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: hostKeyCallback,
		Timeout:         cfg.Timeout,
	}, nil
}
