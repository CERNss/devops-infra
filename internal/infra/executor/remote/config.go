package remote

type SSHConfig struct {
	User     string
	Host     string
	Port     int
	KeyPath  string
	Password string
	// KnownHostsPath defaults to ~/.ssh/known_hosts when empty.
	KnownHostsPath string
	// InsecureIgnoreHostKey skips host key verification (use only when you accept the risk).
	InsecureIgnoreHostKey bool
}
