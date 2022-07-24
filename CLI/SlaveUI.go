package main

import (
	"ArchcopyCLI/util"
	"errors"
	"fmt"
	"os"

	rpc "Archcopy/RPC"
	rpcslave "Archcopy/rpcslave"
)

func Slave() int {
	var Server rpcslave.ServerStruct

	if UI.RPC.Auto {
		rpc.GenerateCredentials()
		rpc.ShowCredentials()
	} else {
		TLSPieces := 0
		if len(UI.RPC.ServerCertificateKey) != 0 {
			TLSPieces++
		}
		if len(UI.RPC.ServerCertificate) > 0 {
			TLSPieces++
		}
		if len(UI.RPC.CACertificate) > 0 {
			TLSPieces++
		}

		switch TLSPieces {
		case 3:
			rpc.LoadServerCertificate(UI.RPC.ServerCertificate, UI.RPC.ServerCertificateKey)
			rpc.LoadCACertificate(UI.RPC.CACertificate, "")
		case 0:
		default:
			util.Fatal(errors.New("TLS requires a CA certificate, a server certificate, and a server private key."), "")
		}

		if len(UI.RPC.PSK) == 0 {
			util.Fatal(errors.New("No PSK provided."), "")
		}
		rpc.SetPSK(UI.RPC.PSK)
	}

	ConStr := rpc.ParseRemoteURL(UI.RPC.SlaveURL)
	if ConStr.SSH {
		util.Fatal(errors.New("SSH transport not valid in slave mode."), "")
	}

	Server.Network = ConStr.Network
	Server.Address = ConStr.Address

	Server.Certificate = rpc.ServerCert
	Server.CACertificate = rpc.Credentials.CACert

	Server.SingleSession = UI.RPC.SingleSession

	TLS := "disabled"
	if rpc.ServerCert != nil {
		TLS = "enabled"
	}
	fmt.Printf("Slave mode. Waiting for connection on %v://%v. TLS is %v.\n", Server.Network, Server.Address, TLS)

	defer func() {
		if Server.Network == "unix" {
			os.Remove(Server.Address)
		}
	}()

	Server.Slave()

	return 0
}
