package main

import (
	common "Archcopy/Common"
	"Archcopy/FSLocal"
	rpcmaster "Archcopy/FSRemote"
	rpc "Archcopy/RPC"
	"ArchcopyCLI/util"
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"

	"golang.org/x/crypto/ssh"
)

type TransportStruct struct {
	SSHSession *ssh.Session
	SSHConn    net.Conn

	IsRemote bool

	FSAccess common.FSAccess
}

func SetupTransport(URL string) (*TransportStruct, error) {
	var err error
	var TS TransportStruct

	if len(URL) == 0 {
		TS.FSAccess = new(FSLocal.FSAccessLocalStruct)
		return &TS, err
	}

	TS.IsRemote = true

	t := new(rpcmaster.FSAccessRemoteStruct)

	ConnStr := rpc.ParseRemoteURL(URL)

	if !ConnStr.SSH {
		if len(UI.RPC.PSK) == 0 {
			fmt.Fprintf(os.Stderr, "No PSK provided.\n")
			os.Exit(1)
		}
		rpc.SetPSK(UI.RPC.PSK)
	} else {
		ConnStr = SSHConnect(ConnStr, &TS)
	}

	TLSPieces := 0
	if len(UI.RPC.ClientCertificateKey) > 0 {
		TLSPieces++
	}
	if len(UI.RPC.ClientCertificate) > 0 {
		TLSPieces++
	}
	if len(UI.RPC.CACertificate) > 0 {
		TLSPieces++
	}

	switch TLSPieces {
	case 3:
		rpc.LoadClientCertificate(UI.RPC.ClientCertificate, UI.RPC.ClientCertificateKey)
		rpc.LoadCACertificate(UI.RPC.CACertificate, "")
	case 0:
	default:
		fmt.Fprintf(os.Stderr, "TLS requires a CA certificate, a client certificate, and a client private key.\n")
		os.Exit(1)
	}

	err = t.Connect(ConnStr)

	if err != nil {
		util.Fatal(err, "Connecting transport")
	}

	TS.FSAccess = t

	return &TS, err
}

func (o *TransportStruct) Close() {
	o.FSAccess.Close()
	if o.SSHSession != nil {
		o.SSHConn.Close()
		o.SSHSession.Close()
	}
}

//
// SSH logic
//

func SSHConnect(ConnStr rpc.RemoteConnectStruct, Transport *TransportStruct) rpc.RemoteConnectStruct {
	var RetConnStr rpc.RemoteConnectStruct

	raddr := "/tmp/archcopy-" + RandString()
	rsocket := "unix://" + raddr
	command := "archcopy --listenurl " + rsocket + " slave --auto --singlesession"

	conf := &ssh.ClientConfig{
		User:            ConnStr.User,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conf.Auth = ReadKeys()

	client, err := ssh.Dial(ConnStr.Network, strings.Join([]string{ConnStr.Host, ":", ConnStr.Port}, ""), conf)
	util.Fatal(err, "failed to dial SSH server")
	Transport.SSHSession, err = client.NewSession()
	util.Fatal(err, "failed to create SSH session")

	stdoutPipe, err := Transport.SSHSession.StdoutPipe()
	util.Fatal(err, "")

	JTC := make(chan string)
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			Line := scanner.Text()
			fmt.Println("R:", Line)
			if strings.HasPrefix(Line, "JSON: {") {
				JTC <- strings.TrimPrefix(Line, "JSON: ")
				break
			}
		}
		for scanner.Scan() {
			Line := scanner.Text()
			fmt.Println("R:", Line)
		}
	}()

	err = Transport.SSHSession.Start(command)
	util.Fatal(err, "failed to run command over SSH")

	var JSONText string

	JSONText = <-JTC

	err = rpc.LoadCredentials(JSONText)
	util.Fatal(err, "")

	RetConnStr.Network = "unix"
	RetConnStr.Address = raddr
	RetConnStr.SSH = true
	RetConnStr.Dial = "unix:///" + raddr

	//	RetConnStr.Dialer = func() (net.Conn, error) { return client.Dial(RetConnStr.Network, RetConnStr.Address) }
	RetConnStr.Dialer = func(_ context.Context, addr string) (net.Conn, error) {
		var err error
		Transport.SSHConn, err = client.Dial("unix", RetConnStr.Address)
		return Transport.SSHConn, err
	}

	return RetConnStr
}

func RandString() string {
	b := make([]byte, 24)
	rand.Read(b)
	return util.Base64(b)
}

func ReadKeys() []ssh.AuthMethod {
	var r []ssh.AuthMethod

	if len(UI.RPC.SSHPubKey) == 0 {
		pth, err := os.UserHomeDir()
		util.Fatal(err, "Get home directory.")

		UI.RPC.SSHPubKey = path.Join(pth, ".ssh/id_rsa")
	}

	if len(UI.RPC.SSHPubKeyPass) == 0 {
		UI.RPC.SSHPubKeyPass = os.Getenv("archcopysshpassphrase")
	}

	r = append(r, readPubKey(UI.RPC.SSHPubKey, UI.RPC.SSHPubKeyPass))

	return r
}

func readPubKey(file string, keyPass string) ssh.AuthMethod {
	var key ssh.Signer
	var err error
	var b []byte
	b, err = ioutil.ReadFile(file)
	util.Fatal(err, "failed to read public key")
	if !strings.Contains(string(b), "ENCRYPTED") {
		key, err = ssh.ParsePrivateKey(b)
		util.Fatal(err, "failed to parse private key")
	} else {
		key, err = ssh.ParsePrivateKeyWithPassphrase(b, []byte(keyPass))
		util.Fatal(err, "failed to parse password-protected private key")
	}
	return ssh.PublicKeys(key)
}
