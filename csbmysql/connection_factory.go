package csbmysql

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const customCaConfigName = "custom-ca"

type connectionFactory struct {
	host          string
	port          int
	username      string
	password      string
	database      string
	caCertificate []byte
}

func (c connectionFactory) ConnectAsAdmin() (*sql.DB, error) {
	if len(c.caCertificate) > 0 {
		err := c.registerCustomCA()
		if err != nil {
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
	if len(c.caCertificate) > 0 {
		return customCaConfigName
	}
	return "true"
}

func (c connectionFactory) registerCustomCA() error {
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(c.caCertificate); !ok {
		return fmt.Errorf("unable to append CA cert:\n[ %v ]", c.caCertificate)
	}
	tlsConfig := &tls.Config{}
	tlsConfig.RootCAs = certPool
	err := mysql.RegisterTLSConfig(customCaConfigName, tlsConfig)
	if err != nil {
		return fmt.Errorf("unable to register custom-ca mysql config: %s", err.Error())
	}
	return nil
}

func (c connectionFactory) uri() string {
	return c.uriWithCreds(c.username, c.password)
}

func (c connectionFactory) uriRedacted() string {
	return strings.ReplaceAll(c.uri(), c.password, "REDACTED")
}
