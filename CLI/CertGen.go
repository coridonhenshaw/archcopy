package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"time"
)

func MakeCA() (Certificate []byte, PrivateKey []byte, OutCA *x509.Certificate, OutCAPrivateKey *rsa.PrivateKey) {
	Sn, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		log.Panic(err)
	}
	ca := &x509.Certificate{
		SerialNumber: Sn,
		Subject: pkix.Name{
			Organization:  []string{"Archcopy"},
			Country:       []string{""},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 4, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Panic(err)
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		log.Panic(err)
	}

	cert, err := x509.ParseCertificate(caBytes)
	if err != nil {
		panic("Failed to parse certificate:" + err.Error())
	}

	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	return caPEM.Bytes(), caPrivKeyPEM.Bytes(), cert, caPrivKey
}

func MakeCertificate(ca *x509.Certificate, caPrivKey *rsa.PrivateKey) (Certificate []byte, PrivateKey []byte, outc *x509.Certificate, outd *rsa.PrivateKey) {
	Sn, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		log.Panic(err)
	}

	cert := &x509.Certificate{
		SerialNumber: Sn,
		Subject: pkix.Name{
			Organization: []string{"Archcopy"},
			CommonName:   "Archcopy",
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(0, 4, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		DNSNames:     []string{"Archcopy"},
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Panic(err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		log.Panic(err)
	}

	certo, err := x509.ParseCertificate(certBytes)
	if err != nil {
		panic("Failed to parse certificate:" + err.Error())
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	return certPEM.Bytes(), certPrivKeyPEM.Bytes(), certo, certPrivKey

}

func MakeCerts() int {
	fmt.Print("Generating TLS certificates.\n\n")
	CACert, _, c, d := MakeCA()
	os.WriteFile("ca,cert", CACert, 0600)
	fmt.Println("Certificate Authority: ca.cert")

	{
		Cert, Key, _, _ := MakeCertificate(c, d)
		os.WriteFile("server.cert", Cert, 0600)
		os.WriteFile("server.key", Key, 0600)
	}
	fmt.Println("Server: server.cert and server.key")

	{
		Cert, Key, _, _ := MakeCertificate(c, d)
		os.WriteFile("client.cert", Cert, 0600)
		os.WriteFile("client.key", Key, 0600)
	}

	fmt.Print("Server: client.cert and client.key\n\n")
	return 0
}

func verifyLow(root, child *x509.Certificate) {
	roots := x509.NewCertPool()
	roots.AddCert(root)
	opts := x509.VerifyOptions{
		Roots: roots,
	}

	if _, err := child.Verify(opts); err != nil {
		panic("failed to verify certificate: " + err.Error())
	}
	fmt.Println("Low Verified")
}
