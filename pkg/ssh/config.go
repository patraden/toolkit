package ssh

import "time"

// ClientConfig holds the parameters required to establish an SSH connection.
type ClientConfig struct {
	// Addr is the remote host address in "host:port" form.
	Addr string
	// User is the username to authenticate as.
	User string
	// Timeout is the maximum duration allowed for the TCP connection to be
	// established. Zero means no timeout.
	Timeout time.Duration
	// PrivateKeyPath is the path to an unencrypted PEM-encoded private key file
	// used for public-key authentication.
	PrivateKeyPath string
	// KnownHostsPath is the path to a known_hosts file used to validate the
	// remote host key. The file must exist and contain an entry for the target
	// host before a connection is attempted.
	KnownHostsPath string
}
