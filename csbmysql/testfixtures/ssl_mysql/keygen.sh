#!/usr/bin/env bash
set -e 

# Generate the self-signed CA certificate and key
openssl req -new -x509 -sha256 -days 3650 -nodes -out certs/ca.crt \
  -keyout keys/ca.key -subj "/CN=root-ca"

# Create the server key and a certificate sign request (CSR) for signing
openssl req -new -nodes -out server.csr \
  -keyout keys/server.key -subj "/CN=localhost"

# Sign the CSR with the root CA key, producing a CA-signed certificate
openssl x509 -req -in server.csr -sha256 -days 3650 \
    -CA certs/ca.crt -CAkey keys/ca.key -CAcreateserial \
    -extfile <( echo "subjectAltName = DNS:localhost" ) \
    -out certs/server.crt

# The CSR is no longer needed
rm server.csr

openssl ecparam -name prime256v1 -genkey -noout -out keys/client.key

openssl req -new -sha256 -key keys/client.key -out keys/client.csr -subj "/CN=mysql"

openssl x509 -req -in keys/client.csr -CA certs/ca.crt -CAkey keys/ca.key -CAcreateserial -out certs/client.crt -days 3650 -sha256
