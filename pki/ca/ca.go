package ca

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"sort"
	"time"

	"github.com/op/go-logging"
)

var (
	ErrCertRevoked   = errors.New("certificate revoked")
	ErrCorruptWrite  = errors.New("bytes written don't match data length")
	ErrCertNotFound  = errors.New("certificate not found")
	ErrInvalidConfig = errors.New("invalid CA configuration")
)

type CertificateAuthority interface {
	Init() error
	IsEstablished() bool
	IssueCertificate(request CertificateRequest) ([]byte, error)
	CreateCSR(email string, request CertificateRequest) ([]byte, error)
	SignCSR(csrBytes []byte, request CertificateRequest) ([]byte, error)
	Verify(certBytes []byte) (bool, error)
	DecodeCSR(bytes []byte) (*x509.CertificateRequest, error)
	DecodePEM(bytes []byte) (*x509.Certificate, error)
	PEM(cn string) ([]byte, error)
	Revoke(cn string) (crl []byte, err error)
	PrivateKey(cn string) (*rsa.PrivateKey, error)
	PublicKey(cn string) (*rsa.PublicKey, error)
	TrustStore() TrustStore
	IssuedCertificates() ([]string, error)
}

type CAService struct {
	logger              *logging.Logger
	config              *Config
	certDir             string
	identity            *x509.Certificate
	privateKey          *rsa.PrivateKey
	publicKey           *rsa.PublicKey
	revokedCertificates []x509.RevocationListEntry
	revocationList      *x509.RevocationList
	trustStore          TrustStore
	CertificateAuthority
}

// Createa a new x509 Certificate Authority
func NewCertificateAuthority(
	logger *logging.Logger,
	certDir string,
	config *Config) (CertificateAuthority, error) {

	if config == nil {
		return nil, ErrInvalidConfig
	}

	return &CAService{
		logger:     logger,
		certDir:    certDir,
		config:     config,
		trustStore: NewDebianTrustStore(logger, certDir)}, nil
}

// Returns the operating system's CA trusted certificates store provider
func (ca *CAService) TrustStore() TrustStore {
	return ca.trustStore
}

// Returns true if the CA has already been initialized
// and established, false otherwise.
func (ca *CAService) IsEstablished() bool {
	_, err := ca.PEM("ca")
	if err != nil {
		if err == ErrCertNotFound || os.IsNotExist(err) {
			return false
		}
		ca.logger.Fatal(err)
	}
	return true
}

// Load CA private and public keys from the cert store, decode the
// public key to x509 certificate and set it as the CA identity /
// signing certificate.
func (ca *CAService) loadCA() error {

	ca.logger.Debug("Loading Certificate Authority")

	privateKey, err := ca.PrivateKey("ca")
	if err != nil {
		ca.logger.Fatal(err)
		return err
	}

	ca.privateKey = privateKey
	ca.publicKey = &privateKey.PublicKey

	publicPEM, err := ca.PEM("ca")
	if err != nil {
		ca.logger.Error(err)
		return err
	}

	cert, err := ca.DecodePEM(publicPEM)
	if err != nil {
		ca.logger.Error(err)
		return err
	}

	ca.identity = cert

	return nil
}

// The first time the CA is instantiated, a new RSA private / public key and
// x509 signing certificate is generated and saved to the cert store. Subsequent
// initializations load the generated keys and signing certificate so the CA is
// ready to start servicing requests. Certificates are saved to the cert store
// in PEM format. This methos returns the raw DER encoded []byte array as returned
// from x509.CreateCertificate.
func (ca *CAService) Init() error {

	if ca.IsEstablished() {
		return ca.loadCA()
	}

	var privateKey *rsa.PrivateKey

	ca.logger.Debug("Initializing new Certificate Authority")

	if err := os.MkdirAll(fmt.Sprintf("%s/revoked", ca.certDir), 0700); err != nil {
		ca.logger.Error(err)
		return err
	}

	ipAddresses, dnsNames, emailAddresses, err := parseSANS(ca.config.Identity.SANS)
	if err != nil {
		ca.logger.Error(err)
		return err
	}

	serialNumber, err := newSerialNumber()
	if err != nil {
		ca.logger.Error(err)
		return err
	}

	// Generate new CA private key
	privateKey, err = rsa.GenerateKey(rand.Reader, ca.config.Identity.KeySize)
	if err != nil {
		ca.logger.Error(err)
		return err
	}

	ca.privateKey = privateKey
	ca.publicKey = &privateKey.PublicKey

	subjectKeyID, err := ca.createSubjectKeyIdentifier()
	if err != nil {
		ca.logger.Error(err)
		return err
	}

	// Create the CA signing certficate
	ca.identity = &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         ca.config.Identity.Subject.CommonName,
			Organization:       []string{ca.config.Identity.Subject.Organization},
			OrganizationalUnit: []string{ca.config.Identity.Subject.OrganizationalUnit},
			Country:            []string{ca.config.Identity.Subject.Country},
			Province:           []string{ca.config.Identity.Subject.Province},
			Locality:           []string{ca.config.Identity.Subject.Locality},
			StreetAddress:      []string{ca.config.Identity.Subject.Address},
			PostalCode:         []string{ca.config.Identity.Subject.PostalCode}},
		SubjectKeyId:          subjectKeyID,
		AuthorityKeyId:        subjectKeyID,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(ca.config.Identity.Valid, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		DNSNames:              dnsNames,
		IPAddresses:           ipAddresses,
		EmailAddresses:        emailAddresses}

	caDerCert, err := x509.CreateCertificate(rand.Reader,
		ca.identity, ca.identity, &ca.privateKey.PublicKey, ca.privateKey)

	if err != nil {
		ca.logger.Error(err)
		return err
	}

	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caDerCert,
	})

	err = ca.savePrivateKey("ca", privateKey)
	if err != nil {
		ca.logger.Error(err)
		return err
	}

	if err := ca.save("ca.crt", caPEM.Bytes()); err != nil {
		ca.logger.Error(err)
		return err
	}

	// if err := ca.trustStore.Install("ca"); err != nil {
	// 	ca.logger.Error(err)
	// 	return err
	// }

	ca.logger.Debug(caPEM.String())

	return nil
}

// Creates a new Certificate Signing Request (CSR)
func (ca *CAService) CreateCSR(email string, request CertificateRequest) ([]byte, error) {

	var oidEmailAddress = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 1}

	ipAddresses, dnsNames, emailAddresses, err := parseSANS(request.SANS)
	if err != nil {
		ca.logger.Error(err)
		return nil, err
	}
	emailAddresses = append(emailAddresses, email)

	_subject := pkix.Name{
		CommonName:         request.Subject.CommonName,
		Organization:       []string{request.Subject.Organization},
		OrganizationalUnit: []string{request.Subject.OrganizationalUnit},
		Country:            []string{request.Subject.Country},
		Province:           []string{request.Subject.Province},
		Locality:           []string{request.Subject.Locality},
		StreetAddress:      []string{request.Subject.Address},
		PostalCode:         []string{request.Subject.PostalCode},
	}

	rawSubj := _subject.ToRDNSequence()
	rawSubj = append(rawSubj, []pkix.AttributeTypeAndValue{
		{Type: oidEmailAddress, Value: emailAddresses},
	})

	asn1Subj, _ := asn1.Marshal(rawSubj)

	template := x509.CertificateRequest{
		RawSubject:         asn1Subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
		DNSNames:           dnsNames,
		IPAddresses:        ipAddresses,
		EmailAddresses:     emailAddresses,
	}

	dnsNames = append(dnsNames, request.Subject.CommonName)
	template.DNSNames = dnsNames

	csrPEM := new(bytes.Buffer)
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, ca.privateKey)
	if err != nil {
		ca.logger.Error(err)
		return nil, err
	}

	csrBlock := &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes}

	if err := pem.Encode(csrPEM, csrBlock); err != nil {
		ca.logger.Error(err)
		return nil, err
	}

	// if err := pem.Encode(os.Stdout, csrBlock); err != nil {
	// 	return err
	// }

	pemBytes := csrPEM.Bytes()
	filename := fmt.Sprintf("%s.csr", request.Subject.CommonName)
	if err := ca.save(filename, pemBytes); err != nil {
		ca.logger.Error(err)
		return nil, err
	}

	ca.logger.Debug(string(pemBytes))

	return pemBytes, nil
}

// Signs a Certificate Signing Request (CSR) and stores it in the cert store
// in PEM format. This method returns the raw DER encoded []byte array as
// returned from x509.CreateCertificate.
func (ca *CAService) SignCSR(csrBytes []byte, request CertificateRequest) ([]byte, error) {

	csr, err := ca.DecodeCSR(csrBytes)
	if err != nil {
		ca.logger.Error(err)
		return nil, err
	}

	serialNumber, err := newSerialNumber()
	if err != nil {
		ca.logger.Error(err)
		return nil, err
	}

	template := x509.Certificate{
		Signature:          csr.Signature,
		SignatureAlgorithm: csr.SignatureAlgorithm,

		PublicKeyAlgorithm: csr.PublicKeyAlgorithm,
		PublicKey:          csr.PublicKey,

		SerialNumber:   serialNumber,
		Issuer:         ca.identity.Subject,
		Subject:        csr.Subject,
		AuthorityKeyId: ca.identity.SubjectKeyId,
		SubjectKeyId:   ca.identity.SubjectKeyId,

		NotBefore:      time.Now(),
		NotAfter:       time.Now().AddDate(0, 0, request.Valid),
		KeyUsage:       x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		IPAddresses:    csr.IPAddresses,
		DNSNames:       csr.DNSNames,
		EmailAddresses: csr.EmailAddresses,
	}
	template.DNSNames = csr.DNSNames

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, ca.identity, template.PublicKey, ca.privateKey)
	if err != nil {
		ca.logger.Error(err)
		return nil, err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	pemBytes := certPEM.Bytes()
	if err := ca.save(fmt.Sprintf("%s.crt", csr.Subject.CommonName), pemBytes); err != nil {
		ca.logger.Error(err)
		return nil, err
	}

	ca.logger.Debug(string(pemBytes))

	return pemBytes, nil
}

// Create a new private / public key pair and save it to the cert store
// in PEM format. This method returns the raw DER encoded []byte array as
// returned from x509.CreateCertificate.
func (ca *CAService) IssueCertificate(request CertificateRequest) ([]byte, error) {
	ipAddresses, dnsNames, emailAddresses, err := parseSANS(request.SANS)
	if err != nil {
		ca.logger.Error(err)
		return nil, err
	}
	dnsNames = append(dnsNames, request.Subject.CommonName)
	serialNumber, err := newSerialNumber()
	if err != nil {
		ca.logger.Error(err)
		return nil, err
	}
	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Issuer:       ca.identity.Subject,
		Subject: pkix.Name{
			CommonName:    request.Subject.CommonName,
			Organization:  []string{request.Subject.Organization},
			Country:       []string{request.Subject.Country},
			Province:      []string{request.Subject.Province},
			Locality:      []string{request.Subject.Locality},
			StreetAddress: []string{request.Subject.Address},
			PostalCode:    []string{request.Subject.PostalCode},
		},
		// IPAddresses:    []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:      time.Now(),
		NotAfter:       time.Now().AddDate(0, 0, request.Valid),
		AuthorityKeyId: ca.identity.SubjectKeyId,
		SubjectKeyId:   ca.identity.SubjectKeyId,
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:       x509.KeyUsageDigitalSignature,
		DNSNames:       dnsNames,
		IPAddresses:    ipAddresses,
		EmailAddresses: emailAddresses}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, ca.config.Identity.KeySize)
	if err != nil {
		ca.logger.Error(err)
		return nil, err
	}

	certDerBytes, err := x509.CreateCertificate(rand.Reader, cert, ca.identity, &certPrivKey.PublicKey, ca.privateKey)
	if err != nil {
		ca.logger.Error(err)
		return nil, err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDerBytes,
	})

	certPEMBytes := certPEM.Bytes()
	certFile := fmt.Sprintf("%s.crt", request.Subject.CommonName)
	if err := ca.save(certFile, certPEMBytes); err != nil {
		ca.logger.Error(err)
		return nil, err
	}

	err = ca.savePrivateKey(request.Subject.CommonName, certPrivKey)
	if err != nil {
		ca.logger.Error(err)
		return nil, err
	}

	ca.logger.Debug(string(certPEMBytes))

	return certDerBytes, nil
}

// Revokes a certificate
func (ca *CAService) Revoke(cn string) ([]byte, error) {

	if _, err := os.Stat(fmt.Sprintf("%s/revoked/%s.crt", ca.certDir, cn)); err == nil {
		return nil, ErrCertRevoked
	}

	// Load the requested cert
	certPEM, err := ca.PEM(cn)
	if err != nil {
		return nil, err
	}

	// Decode the PEM to a *x509.Certificate
	certificate, err := ca.DecodePEM(certPEM)
	if err != nil {
		ca.logger.Error(err)
		return nil, err
	}

	// Check to see if the certificate is already revoked
	if ca.revocationList != nil {
		for _, serialNumber := range ca.revocationList.RevokedCertificateEntries {
			if serialNumber.SerialNumber.String() == certificate.SerialNumber.String() {
				return nil, ErrCertRevoked
			}
		}
		ca.revokedCertificates = ca.revocationList.RevokedCertificateEntries
	} else {
		ca.revokedCertificates = make([]x509.RevocationListEntry, 0)
	}

	// Create a new revocation entry
	ca.revokedCertificates = append(ca.revokedCertificates,
		x509.RevocationListEntry{
			SerialNumber:   certificate.SerialNumber,
			RevocationTime: time.Now()})

	// Create a new revocation list serial number and template with the updated revokedCertificates list
	serialNumber, err := newSerialNumber()
	if err != nil {
		return nil, err
	}
	template := x509.RevocationList{
		SignatureAlgorithm:        ca.identity.SignatureAlgorithm,
		RevokedCertificateEntries: ca.revokedCertificates,
		Number:                    serialNumber,
		ThisUpdate:                time.Now(),
		NextUpdate:                time.Now().AddDate(0, 0, 1),
	}

	// Create the new revocation list
	crlBytes, err := x509.CreateRevocationList(rand.Reader, &template, ca.identity, ca.privateKey)
	if err != nil {
		return nil, err
	}

	// Update the revocation list database
	if err := ca.save("ca.crl", crlBytes); err != nil {
		ca.logger.Error(err)
		return nil, err
	}

	revocationList, err := x509.ParseRevocationList(crlBytes)
	if err != nil {
		ca.logger.Error(err)
		return nil, err
	}
	ca.revocationList = revocationList

	// Move the certs to the revoked folder
	os.Rename(fmt.Sprintf("%s/%s.csr", ca.certDir, cn), fmt.Sprintf("%s/revoked/%s.csr", ca.certDir, cn))
	os.Rename(fmt.Sprintf("%s/%s.crt", ca.certDir, cn), fmt.Sprintf("%s/revoked/%s.crt", ca.certDir, cn))
	os.Rename(fmt.Sprintf("%s/%s.key", ca.certDir, cn), fmt.Sprintf("%s/revoked/%s.key", ca.certDir, cn))
	os.Rename(fmt.Sprintf("%s/%s.key.pkcs8", ca.certDir, cn), fmt.Sprintf("%s/revoked/%s.key.pkcs8", ca.certDir, cn))

	//ca.logger.Debug(string(crlBytes))

	return crlBytes, err
}

// Verifies a certificate is valid
func (ca *CAService) Verify(certBytes []byte) (bool, error) {

	// Load the requested certificate
	certificate, err := ca.DecodePEM(certBytes)
	if err != nil {
		ca.logger.Error(err)
		return false, err
	}

	// Check to see if the certificate is revoked
	if ca.revocationList != nil {
		for _, serialNumber := range ca.revocationList.RevokedCertificateEntries {
			if serialNumber.SerialNumber.String() == certificate.SerialNumber.String() {
				return false, ErrCertRevoked
			}
		}
		ca.revokedCertificates = ca.revocationList.RevokedCertificateEntries
	} else {
		ca.revokedCertificates = make([]x509.RevocationListEntry, 0)
	}

	// Load the CA public key in PEM format
	pubBytes, err := ca.PEM("ca")
	if err != nil {
		ca.logger.Error(err)
		return false, err
	}

	// Load the CA public key into the root certificate pool
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(pubBytes))
	if !ok {
		ca.logger.Error(err)
		return false, err
	}

	// Set the verify options containing the known root CA
	// certificates and the common name of this CA.
	opts := x509.VerifyOptions{
		Roots:         roots,
		DNSName:       ca.identity.Subject.CommonName,
		Intermediates: x509.NewCertPool(),
	}

	if _, err := certificate.Verify(opts); err != nil {
		return false, err
	}

	// The cert is valid
	return true, nil
}

// Returns the operating system's CA trusted certificates store provider
func (ca *CAService) IssuedCertificates() ([]string, error) {
	certs := make(map[string]bool, 0)
	files, err := os.ReadDir(ca.certDir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if file.Name() == "ca.crl" {
			continue
		}
		// pieces := strings.Split(file.Name(), ".")
		// certs[pieces[0]] = true
		certs[file.Name()] = true
	}
	names := make([]string, len(certs))
	i := 0
	for k := range certs {
		names[i] = k
		i++
	}
	sort.Strings(names)
	return names, nil
}

// Saves a private key in PEM and PCKS8 format
func (ca *CAService) savePrivateKey(cn string, privateKey *rsa.PrivateKey) error {
	// PKCS8
	pkcs8PrivKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		ca.logger.Error(err)
		return err
	}
	pkcs8File := fmt.Sprintf("%s.key.pkcs8", cn)
	if err := ca.save(pkcs8File, pkcs8PrivKeyBytes); err != nil {
		ca.logger.Error(err)
		return err
	}
	// PEM
	caPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		ca.logger.Error(err)
		return err
	}
	pemFile := fmt.Sprintf("%s.key", cn)
	if err := ca.save(pemFile, caPrivKeyPEM.Bytes()); err != nil {
		ca.logger.Error(err)
		return err
	}
	return nil
}

// Saves the passed data to the specified filename in the CA data directory
func (ca *CAService) save(filename string, data []byte) error {
	fo, err := os.Create(fmt.Sprintf("%s/%s", ca.certDir, filename))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fo.Close(); err != nil {
			ca.logger.Error(err)
		}
	}()
	bytesWritten, err := fo.Write(data)
	if bytesWritten != len(data) {
		return ErrCorruptWrite
	}
	if err != nil {
		ca.logger.Error(err)
		return err
	}
	return nil
}

// Generates a new certificate serial number
func newSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	return rand.Int(rand.Reader, serialNumberLimit)
}

// Parses extended IP, DNS, and Email addresses SubjectAlternativeNames (SANS)
func parseSANS(sans *SubjectAlternativeNames) ([]net.IP, []string, []string, error) {
	var ipAddresses []net.IP
	var dnsNames []string
	var emailAddresses []string
	if sans != nil {
		ipAddresses = make([]net.IP, len(sans.IPs))
		for i, ip := range sans.IPs {
			ip, _, err := net.ParseCIDR(fmt.Sprintf("%s/32", ip)) // ip, ipnet, err
			if err != nil {
				return nil, nil, nil, err
			}
			ipAddresses[i] = ip
		}
		dnsNames = make([]string, len(sans.DNS))
		copy(dnsNames, sans.DNS)
		emailAddresses = make([]string, len(sans.Email))
		copy(emailAddresses, sans.Email)
	}
	return ipAddresses, dnsNames, emailAddresses, nil
}

// Build Subject Key Identifier
func (ca *CAService) createSubjectKeyIdentifier() ([]byte, error) {
	var spki struct {
		Algorithm        pkix.AlgorithmIdentifier
		SubjectPublicKey asn1.BitString
	}
	spkiASN1, err := x509.MarshalPKIXPublicKey(ca.privateKey.Public())
	if err != nil {
		ca.logger.Error(err)
		return nil, err
	}
	_, err = asn1.Unmarshal(spkiASN1, &spki)
	if err != nil {
		ca.logger.Error(err)
		return nil, err
	}
	skid := sha1.Sum(spki.SubjectPublicKey.Bytes)
	return skid[:], nil
}

// Decodes a CSR byte array to a x509.CertificateRequest
func (ca *CAService) DecodeCSR(bytes []byte) (*x509.CertificateRequest, error) {
	var block *pem.Block
	if block, _ = pem.Decode(bytes); block == nil {
		return nil, ErrInvalidEncoding
	}
	return x509.ParseCertificateRequest(block.Bytes)
}

// Decodes a PEM byte array returning a pointer to its decoded x509.Certificate
func (ca *CAService) DecodePEM(bytes []byte) (*x509.Certificate, error) {
	var block *pem.Block
	if block, _ = pem.Decode(bytes); block == nil {
		return nil, ErrInvalidEncoding
	}
	return x509.ParseCertificate(block.Bytes)
}

// Returns a PEM certifcate from the cert store as a []byte array or
// ErrCertNotFound if the certificate does not exist.
func (ca *CAService) PEM(cn string) ([]byte, error) {
	pem, err := os.ReadFile(fmt.Sprintf("%s/%s.crt", ca.certDir, cn))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrCertNotFound
		}
		return nil, err
	}
	return pem, nil
}

// Parses a PKCS8 formatted RSA private key
func (ca *CAService) PrivateKey(cn string) (*rsa.PrivateKey, error) {
	bytes, err := os.ReadFile(fmt.Sprintf("%s/%s.key.pkcs8", ca.certDir, cn))
	if err != nil {
		return nil, err
	}
	key, err := x509.ParsePKCS8PrivateKey(bytes)
	if err != nil {
		return nil, err
	}
	return key.(*rsa.PrivateKey), nil
}

// Parses a PEM formatted RSA public key
func (ca *CAService) PublicKey(cn string) (*rsa.PublicKey, error) {
	pem, err := ca.PEM(cn)
	if err != nil {
		return nil, err
	}
	return ParseRSAPublicKeyFromPEM(pem)
}

// func loadCRL(bytes []byte) (*x509.RevocationList, error) {
// 	return x509.ParseRevocationList(bytes)
// }

// func (ca *CAService) loadPrivateKey(bytes []byte) (*rsa.PrivateKey, error) {
// 	block, _ := pem.Decode(bytes)
// 	return x509.ParsePKCS1PrivateKey(block.Bytes)
// }

// func (ca *CAService) loadPublicKey(bytes []byte) (*rsa.PublicKey, error) {
// 	block, _ := pem.Decode(bytes)
// 	return x509.ParsePKCS1PublicKey(block.Bytes)
// }

// // Loads the CA private key from the cert store
// func (ca *CAService) loadPrivateKey() ([]byte, error) {
// 	return os.ReadFile(fmt.Sprintf("%s/ca.key", ca.certDir))
// }

// // Loads the CA public key from the cert store
// func (ca *CAService) loadPublicKey() ([]byte, error) {
// 	return os.ReadFile(fmt.Sprintf("%s/ca.crt", ca.certDir))
// }
