package csbmysql

import (
	"database/sql"
	"fmt"
	"strings"
)

type connectionFactory struct {
	host      string
	port      int
	username  string
	password  string
	database  string
	verifyTLS bool
}

func (c connectionFactory) ConnectAsAdmin() (*sql.DB, error) {
	return c.connect(c.uri())
}

func (c connectionFactory) connect(uri string) (*sql.DB, error) {
	db, err := sql.Open("mysql", uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL %q: %w", c.uriRedacted(), err)
	}

	return db, nil
}
func (c connectionFactory) uriWithCreds(username, password string) string {
	uri := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=%s", username, password, c.host, c.port, c.database, c.tlsMode())
	return uri
}

func (c connectionFactory) tlsMode() string {
	if c.verifyTLS {
		return "true"
	}
	return "skip-verify"
}

func (c connectionFactory) uri() string {
	return c.uriWithCreds(c.username, c.password)
}

func (c connectionFactory) uriRedacted() string {
	return strings.ReplaceAll(c.uri(), c.password, "REDACTED")
}
