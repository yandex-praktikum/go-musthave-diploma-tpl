package config

import (
	"errors"
	"fmt"
	"net"
	"strconv"
)

// Для сервера (:8080 или localhost:8080)
func ValidateAddr(addr string) error {
	if addr == "" {
		return errors.New("address cannot be empty")
	}

	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("address must be in format host:port or :port: %w", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("port must be numeric: %w", err)
	}

	if port < 1 || port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}

	return nil
}
