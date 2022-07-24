package rpc

import (
	util "Archcopy/util"
	"context"
	"errors"
	"net"
	"strings"
)

type RemoteConnectStruct struct {
	//TCP/Unix
	Network string
	Address string
	Dial    string
	//SSH
	SSH    bool
	User   string
	Port   string
	Host   string
	Dialer func(context.Context, string) (net.Conn, error)
}

// SSH://user@host
// TCP://127.0.0.1:5555
// UNIX:///run/archcopy

func ParseRemoteURL(In string) RemoteConnectStruct {
	var Out RemoteConnectStruct
	InUPPER := strings.ToUpper(In)

	if strings.HasPrefix(InUPPER, "SSH://") {
		Out.SSH = true
		Out.Network = "tcp"
		Out.Port = "22"

		m := strings.Split(In, "//")[1]
		n := strings.Split(m, "@")
		Out.Host = n[1]
		Out.User = n[0]

	} else if strings.HasPrefix(InUPPER, "TCP://") {
		Out.Network = "tcp"
		Out.Address = strings.Split(In, "//")[1]
		Out.Dial = Out.Address
	} else if strings.HasPrefix(InUPPER, "UNIX://") {
		Out.Network = "unix"
		Out.Address = strings.Split(In, "//")[1]
		Out.Dial = "unix:" + Out.Address
	} else {
		util.Fatal(errors.New("Invalid remote URL prefix."), "")
	}

	return Out
}
