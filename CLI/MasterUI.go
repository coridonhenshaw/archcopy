package main

//func SetupMaster() common.FSAccess {
// var OutputFSAccess common.FSAccess

// t := new(rpcmaster.FSAccessRemoteStruct)

// ConnStr := rpc.ParseRemoteURL(UI.RPC.URL)

// if !ConnStr.SSH {
// 	if len(UI.RPC.PSK) == 0 {
// 		fmt.Fprintf(os.Stderr, "No PSK provided.\n")
// 		os.Exit(1)
// 	}
// 	rpc.SetPSK(UI.RPC.PSK)
// } else {
// 	ConnStr = SSHConnect(ConnStr)
// }

// TLSPieces := 0
// if len(UI.RPC.ClientCertificateKey) > 0 {
// 	TLSPieces++
// }
// if len(UI.RPC.ClientCertificate) > 0 {
// 	TLSPieces++
// }
// if len(UI.RPC.CACertificate) > 0 {
// 	TLSPieces++
// }

// switch TLSPieces {
// case 3:
// 	rpc.LoadClientCertificate(UI.RPC.ClientCertificate, UI.RPC.ClientCertificateKey)
// 	rpc.LoadCACertificate(UI.RPC.CACertificate, "")
// case 0:
// default:
// 	fmt.Fprintf(os.Stderr, "TLS requires a CA certificate, a client certificate, and a client private key.\n")
// 	os.Exit(1)
// }

// err := t.Connect(ConnStr)

// if err != nil {
// 	log.Panic(err)
// }

// OutputFSAccess = t
// UI.Multithread = true
// return OutputFSAccess
//	return nil
//}
