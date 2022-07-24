package rpc

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	util "Archcopy/util"

	"golang.org/x/crypto/pbkdf2"
)

type Cred struct {
	PSK    string
	CACert []byte
}

var (
	Credentials Cred
	ExpandedPSK []byte
	ServerCert  *tls.Certificate
	ClientCert  *tls.Certificate
)

func SetPSK(PSK string) {
	Credentials.PSK = PSK
	ExpandedPSK = pbkdf2.Key([]byte(PSK), []byte("peanuts"), 4096, 32, sha256.New)
}

func Sign(Input ...interface{}) []byte {
	if len(ExpandedPSK) == 0 {
		log.Panic("Key not set.")
	}
	h := hmac.New(sha256.New, ExpandedPSK)
	St, _ := json.Marshal(Input)
	h.Write(St)
	q := h.Sum(nil)
	return q
}

func GenerateCredentials() {
	var err error
	const AutoPSKLength = 6 * 3
	PSKBytes := make([]byte, AutoPSKLength)
	n, err := rand.Read(PSKBytes)
	if err != nil || n != AutoPSKLength {
		log.Panic(err)
	}
	PSKStr := base64.StdEncoding.EncodeToString(PSKBytes)

	SetPSK(PSKStr)
}

func ShowCredentials() {
	fmt.Print("JSON: ")
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(Credentials)
}

func LoadCredentials(JSON string) error {
	err := json.Unmarshal([]byte(JSON), &Credentials)
	SetPSK(Credentials.PSK)
	return err
}

func LoadServerCertificate(CertFilename string, KeyFilename string) {
	cert, err := tls.LoadX509KeyPair(CertFilename, KeyFilename)
	if err != nil {
		util.Fatal(err, "Load server certificate/key pair")
	}
	ServerCert = &cert
}

func LoadClientCertificate(CertFilename string, KeyFilename string) {
	cert, err := tls.LoadX509KeyPair(CertFilename, KeyFilename)
	if err != nil {
		util.Fatal(err, "Load client certificate/key pair")
	}
	ClientCert = &cert
}

func LoadCACertificate(CertFilename string, KeyFilename string) {
	var err error
	Credentials.CACert, err = ioutil.ReadFile(CertFilename)
	util.Fatal(err, "Load certificate authority")
}
