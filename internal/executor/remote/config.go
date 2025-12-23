package remote

type SSHConfig struct {
	User     string
	Host     string
	Port     int
	KeyPath  string
	Password string
}
