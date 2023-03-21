package csbmysql

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

const customCaConfigName = "custom-ca"

type connectionFactory struct {
	host                        string
	port                        int
	username                    string
	password                    string
	database                    string
	caCertificate               []byte
	clientCertificate           []byte
	clientCertificatePrivateKey []byte
	// skipVerify controls whether a client verifies the server's
	// certificate chain and host name. If skipVerify is true, crypto/tls
	// accepts any certificate presented by the server and any host name in that
	// certificate.
	skipVerify bool
}

func (c connectionFactory) ConnectAsAdmin() (*sql.DB, error) {
	if len(c.caCertificate) > 0 {
		if err := c.registerCustomCA(); err != nil {
			return nil, err
		}
	}
	return c.connect(c.uri())
}

func (c connectionFactory) connect(uri string) (*sql.DB, error) {
	db, err := sql.Open("mysql", uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL %q: %w", c.uriRedacted(), err)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)

	return db, nil
}
func (c connectionFactory) uriWithCreds(username, password string) string {
	uri := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=%s", username, password, c.host, c.port, c.database, c.tlsMode())
	return uri
}

func (c connectionFactory) tlsMode() string {
	if c.hasCACertificate() {
		return customCaConfigName
	}

	if c.skipVerify {
		return "skip-verify"
	}

	return "true"
}

func (c connectionFactory) registerCustomCA() error {
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(c.caCertificate); !ok {
		return fmt.Errorf("unable to append CA cert:\n[ %v ]", c.caCertificate)
	}

	tlsConfig := &tls.Config{RootCAs: certPool}

	if c.hasClientCertificate() {
		clientCert := make([]tls.Certificate, 0, 1)
		certs, err := tls.X509KeyPair(c.clientCertificate, c.clientCertificatePrivateKey)
		if err != nil {
			return fmt.Errorf("unable to parse a public/private key pair from a pair of PEM encoded data CA cert:\n[ %s ]", err)
		}

		tlsConfig.Certificates = append(clientCert, certs)
	}

	if c.skipVerify {
		// We can't perform a verify-full with GCP certs.
		// TODO: Research on creating a custom verification using `tls.Config.VerifyConnection` or `tls.Config.VerifyPeerCertificate`.
		tlsConfig.InsecureSkipVerify = true
	}

	err := mysql.RegisterTLSConfig(customCaConfigName, tlsConfig)
	if err != nil {
		return fmt.Errorf("unable to register custom-ca mysql config: %s", err.Error())
	}
	return nil
}

func (c connectionFactory) hasCACertificate() bool {
	return len(c.caCertificate) > 0
}

func (c connectionFactory) hasClientCertificate() bool {
	return len(c.clientCertificate) > 0
}

func (c connectionFactory) uri() string {
	return c.uriWithCreds(c.username, c.password)
}

func (c connectionFactory) uriRedacted() string {
	return strings.ReplaceAll(c.uri(), c.password, "REDACTED")
}
