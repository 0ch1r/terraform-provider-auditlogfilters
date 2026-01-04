package provider

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-sql-driver/mysql"
)

func registerTLSConfig(caFile, certFile, keyFile, serverName string, skipVerify bool) (string, error) {
	if (certFile == "") != (keyFile == "") {
		return "", fmt.Errorf("tls_cert_file and tls_key_file must be set together")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: skipVerify, //nolint:gosec // G402: explicitly user-configurable (dev/test); disabling TLS verification is insecure.
	}

	if serverName != "" {
		tlsConfig.ServerName = serverName
	}

	if caFile != "" {
		cleanCAFile := filepath.Clean(caFile)
		fi, err := os.Stat(cleanCAFile)
		if err != nil {
			return "", fmt.Errorf("stat TLS CA file: %w", err)
		}
		if !fi.Mode().IsRegular() {
			return "", fmt.Errorf("TLS CA file must be a regular file")
		}

		// CA path is user-supplied (provider config) and intentionally read after validation.
		caPEM, err := os.ReadFile(cleanCAFile)
		if err != nil {
			return "", fmt.Errorf("read TLS CA file: %w", err)
		}
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM(caPEM); !ok {
			return "", fmt.Errorf("failed to parse TLS CA file")
		}
		tlsConfig.RootCAs = pool
	}

	if certFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return "", fmt.Errorf("load TLS client cert/key: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	name := fmt.Sprintf("auditlogfilters-%d", time.Now().UnixNano())
	if err := mysql.RegisterTLSConfig(name, tlsConfig); err != nil {
		return "", fmt.Errorf("register TLS config: %w", err)
	}

	return name, nil
}
