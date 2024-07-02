package test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

var (
	organization  = []string{"domain.tld", "My Name"}
	streetaddress = []string{"123 anywhere st"}
	postalcode    = []string{"12345"}
	province      = []string{"Province"}
	locality      = []string{"City"}
	country       = []string{"US"}
)

var (
	rcat  = rootCATemplate()
	icat  = intermediateCATemplate()
	rpriv *rsa.PrivateKey
	ipriv *rsa.PrivateKey
)

func main() {
	rootCA()
	intermediateCA()
	serverCert("server1.name", "1.2.3.4", "2a01::2")
	serverCert("server2.name", "2.3.4.5", "2a02::2")
}

func rootCATemplate() x509.Certificate {
	template := x509.Certificate{}
	template.Subject = pkix.Name{
		Organization:  organization,
		StreetAddress: streetaddress,
		PostalCode:    postalcode,
		Province:      province,
		Locality:      locality,
		Country:       country,
		CommonName:    "localhost",
	}

	template.NotBefore = time.Now()
	template.NotAfter = template.NotBefore.Add(172800 * time.Hour)
	template.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCRLSign
	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	template.IsCA = true
	template.BasicConstraintsValid = true
	extSubjectAltName := pkix.Extension{}
	extSubjectAltName.Id = asn1.ObjectIdentifier{2, 5, 29, 17}
	extSubjectAltName.Critical = false
	extSubjectAltName.Value = []byte(`email:user@domain1.com, URI:http://host.domain1.com/`)
	template.ExtraExtensions = []pkix.Extension{extSubjectAltName}
	return template
}

func intermediateCATemplate() x509.Certificate {
	template := x509.Certificate{}
	template.Subject = pkix.Name{
		Organization:  organization,
		StreetAddress: streetaddress,
		PostalCode:    postalcode,
		Province:      province,
		Locality:      locality,
		Country:       country,
		CommonName:    "SelfTLS CA",
	}

	template.NotBefore = time.Now()
	template.NotAfter = template.NotBefore.Add(172800 * time.Hour)
	template.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCRLSign
	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	template.IsCA = true
	template.BasicConstraintsValid = true
	extSubjectAltName := pkix.Extension{}
	extSubjectAltName.Id = asn1.ObjectIdentifier{2, 5, 29, 17}
	extSubjectAltName.Critical = false
	extSubjectAltName.Value = []byte(`email:user@domain2.com, URI:http://www.domain2.com/`)
	template.ExtraExtensions = []pkix.Extension{extSubjectAltName}
	return template
}

// hosts: []string{"hostname","ipv4addr","ipv6addr"}
func serverTemplate(hosts []string) x509.Certificate {
	template := x509.Certificate{}
	template.Subject = pkix.Name{
		Organization:  organization,
		StreetAddress: streetaddress,
		PostalCode:    postalcode,
		Province:      province,
		Locality:      locality,
		Country:       country,
	}

	template.NotBefore = time.Now()
	template.NotAfter = template.NotBefore.Add(86400 * time.Hour)
	template.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	template.BasicConstraintsValid = true

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
			template.Subject.CommonName = h
		}
	}
	return template
}

func serverCert(host, ipv4, ipv6 string) {
	tpl := serverTemplate([]string{host, ipv4, ipv6})
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		fmt.Println("Failed to generate private key:", err)
		os.Exit(1)
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	tpl.SerialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		fmt.Println("Failed to generate serial number:", err)
		os.Exit(1)
	}
	h := sha512.New()
	pb, e := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if e != nil {
		fmt.Println(e.Error())
		return
	}
	h.Write(pb)
	tpl.SubjectKeyId = h.Sum(nil)
	derBytes, err := x509.CreateCertificate(rand.Reader, &tpl, &icat, &priv.PublicKey, ipriv)
	if err != nil {
		fmt.Println("Failed to create certificate:", err)
		os.Exit(1)
	}
	certOut, err := os.Create(host + ".crt")
	if err != nil {
		fmt.Println("Failed to open "+host+".crt for writing:", err)
		os.Exit(1)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	keyOut, err := os.OpenFile(host+".key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Println("failed to open "+host+".key for writing:", err)
		os.Exit(1)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()
}

func intermediateCA() {
	var err error
	ipriv, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		fmt.Println("Failed to generate private key:", err)
		os.Exit(1)
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	icat.SerialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		fmt.Println("Failed to generate serial number:", err)
		os.Exit(1)
	}
	h := sha512.New()
	pb, e := x509.MarshalPKIXPublicKey(&ipriv.PublicKey)
	if e != nil {
		fmt.Println(e.Error())
		return
	}
	h.Write(pb)
	icat.SubjectKeyId = h.Sum(nil)

	derBytes, err := x509.CreateCertificate(rand.Reader, &icat, &rcat, &ipriv.PublicKey, rpriv)
	if err != nil {
		fmt.Println("Failed to create certificate:", err)
		os.Exit(1)
	}
	certOut, err := os.Create("intermediate.crt")
	if err != nil {
		fmt.Println("Failed to open ca.pem for writing:", err)
		os.Exit(1)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	keyOut, err := os.OpenFile("intermediate.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Println("failed to open ca.key for writing:", err)
		os.Exit(1)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(ipriv)})
	keyOut.Close()
}

func rootCA() {
	var err error
	rpriv, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		fmt.Println("Failed to generate private key:", err)
		os.Exit(1)
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	rcat.SerialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		fmt.Println("Failed to generate serial number:", err)
		os.Exit(1)
	}
	h := sha512.New()
	pb, e := x509.MarshalPKIXPublicKey(&rpriv.PublicKey)
	if e != nil {
		fmt.Println(e.Error())
		return
	}
	h.Write(pb)
	rcat.SubjectKeyId = h.Sum(nil)

	derBytes, err := x509.CreateCertificate(rand.Reader, &rcat, &rcat, &rpriv.PublicKey, rpriv)
	if err != nil {
		fmt.Println("Failed to create certificate:", err)
		os.Exit(1)
	}
	certOut, err := os.Create("rootca.crt")
	if err != nil {
		fmt.Println("Failed to open ca.pem for writing:", err)
		os.Exit(1)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	keyOut, err := os.OpenFile("rootca.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Println("failed to open ca.key for writing:", err)
		os.Exit(1)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rpriv)})
	keyOut.Close()
}
