package config

// LoadDir reads config.toml from dir without requiring the XDG config path.
func LoadDir(dir string) (*Config, error) {
	return loadDir(dir)
}

// ParseAddr is the host:port test seam.
func ParseAddr(addr string) (string, int, error) {
	return parseAddr(addr)
}

// NormalizeHost is the unspecified-bind test seam.
func NormalizeHost(host string) (string, error) {
	return normalizeHost(host)
}

// ValidateListen is the port/token test seam.
func ValidateListen(host string, port int, token string) error {
	return validateListen(host, port, token)
}

// ResolvePath is the cookie path test seam.
func ResolvePath(dir, raw string) string {
	return resolvePath(dir, raw)
}

// FileExists is the regular-file test seam.
func FileExists(path string) bool {
	return fileExists(path)
}

// StubUserHomeDir replaces HOME lookup for missing-HOME tests.
func StubUserHomeDir(homeDir func() (string, error)) func() {
	orig := userHomeDir
	userHomeDir = homeDir

	return func() { userHomeDir = orig }
}

// StubReadFile replaces config file reads for permission-error tests.
func StubReadFile(read func(string) ([]byte, error)) func() {
	orig := readFile
	readFile = read

	return func() { readFile = orig }
}

// StubAbsPath replaces filepath.Abs for --config tests.
func StubAbsPath(abs func(string) (string, error)) func() {
	orig := absPath
	absPath = abs

	return func() { absPath = orig }
}
