package config

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// DefaultAddr is the loopback listen address when [server] addr is empty.
const DefaultAddr = "127.0.0.1:9200"

var (
	errNeedHostPort    = errors.New("无效 addr，需要 host:port")
	errInvalidPort     = errors.New("无效端口")
	errUnspecifiedBind = errors.New("拒绝绑定，请使用 127.0.0.1 或具体地址")
	errTokenRequired   = errors.New("必须设置 token 或 NSK_SERVER_TOKEN")
)

func parseAddr(addr string) (string, int, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		addr = DefaultAddr
	}

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, fmt.Errorf("%w: %s", errNeedHostPort, addr)
	}

	host, err = normalizeHost(host)
	if err != nil {
		return "", 0, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("%w: %s", errInvalidPort, portStr)
	}

	return host, port, nil
}

func normalizeHost(host string) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "127.0.0.1", nil
	}

	if host == "0.0.0.0" || host == "::" || host == "[::]" {
		return "", fmt.Errorf("%w: %s", errUnspecifiedBind, host)
	}

	return host, nil
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}

	ip := net.ParseIP(host)

	return ip != nil && ip.IsLoopback()
}

func validateListen(host string, port int, token string) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("%w: %d", errInvalidPort, port)
	}

	if !isLoopbackHost(host) && strings.TrimSpace(token) == "" {
		return fmt.Errorf("%w: %s", errTokenRequired, host)
	}

	return nil
}
