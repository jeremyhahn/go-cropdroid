package ca

import (
	"os"
	"testing"

	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
)

var CERTS_DIR = "./certs"

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func teardown() {
	os.RemoveAll(CERTS_DIR)
}

func setup() {
	// The CA creates the certs dir on first start
	//os.MkdirAll(CERTS_DIR, os.ModePerm)
}

func TestCA(t *testing.T) {

	ca, err := createService()
	assert.Nil(t, err)

	// Get the CA cert *rsa.PublicKey
	publicKey, err := ca.PublicKey("ca")
	assert.Nil(t, err)
	assert.NotNil(t, publicKey)

	// Get the CA cert *rsa.PrivateKey
	privateKey, err := ca.PrivateKey("ca")
	assert.Nil(t, err)
	assert.NotNil(t, privateKey)

	// openssl rsa -in testorg.cropdroid.com.key -check
	// openssl x509 -in testorg.cropdroid.com.crt -text -noout
	keypair, err := ca.IssueCertificate(
		CertificateRequest{
			Valid: 365, // 1 days
			Subject: Subject{
				CommonName:         "testorg.cropdroid.com",
				Organization:       "Test Organization",
				OrganizationalUnit: "Web Services",
				Country:            "US",
				Locality:           "New York",
				Address:            "123 anywhere street",
				PostalCode:         "54321"},
			SANS: &SubjectAlternativeNames{
				DNS: []string{
					"localhost",
					"localhost.localdomain",
				},
				IPs: []string{
					"127.0.0.1",
				},
				Email: []string{
					"user@testorg.com",
					"root@test.com",
				},
			},
		},
	)
	assert.Nil(t, err)
	assert.NotNil(t, keypair)

	// openssl req -in testme.cropdroid.com.csr -noout -text
	csrBytes, err := ca.CreateCSR(
		"me@mydomain.com",
		CertificateRequest{
			Valid: 365, // 1 days
			Subject: Subject{
				CommonName:         "testme.cropdroid.com",
				Organization:       "Customer Organization",
				OrganizationalUnit: "Farming",
				Country:            "US",
				Locality:           "California",
				Address:            "123 farming street",
				PostalCode:         "01210",
			},
			SANS: &SubjectAlternativeNames{
				DNS: []string{
					"localhost",
					"localhost.localdomain",
					"localhost.testme",
				},
				IPs: []string{
					"127.0.0.1",
					"192.168.1.10",
				},
				Email: []string{
					"user@testme.com",
					"info@testme.com",
				},
			},
		})
	assert.Nil(t, err)
	assert.NotNil(t, csrBytes)

	// openssl x509 -in testme.cropdroid.com.crt -text -noout
	certBytes, err := ca.SignCSR(
		csrBytes,
		CertificateRequest{
			Valid: 365, // 1 days
			Subject: Subject{
				CommonName:         "testme.cropdroid.com",
				Organization:       "Customer Organization",
				OrganizationalUnit: "Farming",
				Country:            "US",
				Locality:           "California",
				Address:            "123 farming street",
				PostalCode:         "01210",
			},
			SANS: &SubjectAlternativeNames{
				DNS: []string{
					"localhost",
					"localhost.localdomain",
					"localhost.testme",
				},
				IPs: []string{
					"127.0.0.1",
					"192.168.1.10",
				},
				Email: []string{
					"user@testme.com",
					"info@testme.com",
				},
			},
		})
	assert.Nil(t, err)
	assert.NotNil(t, certBytes)

	cert, err := ca.DecodePEM(certBytes)
	assert.Nil(t, err)
	assert.NotNil(t, cert)

	// Make sure the cert is valid
	valid, err := ca.Verify(certBytes)
	assert.Nil(t, err)
	assert.True(t, valid)

	// Get the cert *rsa.PublicKey
	publicKey, err = ca.PublicKey("testme.cropdroid.com")
	assert.Nil(t, err)
	assert.NotNil(t, publicKey)

	// Removke the cert
	crlBytes, err := ca.Revoke("testme.cropdroid.com")
	assert.Nil(t, err)
	assert.NotNil(t, crlBytes)

	// Revoke the certificate again to ensure it errors
	_, err = ca.Revoke("testme.cropdroid.com")
	assert.Equal(t, ErrCertRevoked, err)

	// Make sure the cert is no longer valid
	valid, err = ca.Verify(certBytes)
	assert.NotNil(t, err)
	assert.Equal(t, ErrCertRevoked, err)
	assert.False(t, valid)

	// Test the web service
	// openssl s_client -connect localhost:8443 -servername localhost  | openssl x509 -noout -text
}

func createService() (CertificateAuthority, error) {
	caConfig := &Config{
		Identity: Identity{
			KeySize: 1024, // bits
			Valid:   10,   // years
			Subject: Subject{
				Organization: "Automate The Things, LLC",
				Country:      "US",
				Locality:     "Miami",
				Address:      "123 test street",
				PostalCode:   "12345"},
			SANS: &SubjectAlternativeNames{
				DNS: []string{
					"localhost",
					"localhost.localdomain",
				},
				IPs: []string{
					"127.0.0.1",
				},
				Email: []string{
					"root@localhost",
					"root@test.com",
				},
			}},
	}

	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	logging.SetBackend(stdout)
	logger := logging.MustGetLogger("certificate-authority")

	ca, err := NewCertificateAuthority(logger, CERTS_DIR, caConfig)
	if err != nil {
		logger.Fatal(err)
	}

	// openssl rsa -in ca.key -check
	if err := ca.Init(); err != nil {
		logger.Fatal(err)
	}

	return ca, nil
}
