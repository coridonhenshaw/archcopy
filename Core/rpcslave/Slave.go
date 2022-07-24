package rpcslave

import (
	util "Archcopy/util"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	pb "proto.local/archcopyRPC"
)

const SessionLife = 1 * time.Hour

var Server *ServerStruct

type SessionKey string

type SessionStruct struct {
	ClientID        []byte
	Expires         time.Time
	CheckSignatures bool
}

type ClientTableStruct struct {
	Phase     int
	Challenge []byte
	Expires   time.Time
}

func (o *SessionStruct) Touch() {
	Server.SessionTableMutex.Lock()
	o.Expires = time.Now().Add(SessionLife)
	Server.SessionTableMutex.Unlock()
}

type ServerStruct struct {
	pb.UnimplementedSlaveServer

	Network       string
	Address       string
	Certificate   *tls.Certificate
	CACertificate []byte
	TLS           bool
	SingleSession bool

	ClientTable       map[string]*ClientTableStruct
	SessionTable      map[SessionKey]*SessionStruct
	SessionTableMutex sync.RWMutex

	HaveRoot bool

	GRPC *grpc.Server
}

func (s *ServerStruct) Housekeeping(SK *[]byte) *SessionStruct {
	var rc *SessionStruct
	Now := time.Now()

	s.SessionTableMutex.Lock()
	defer s.SessionTableMutex.Unlock()
	if SK != nil {
		k, v := s.SessionTable[SessionKey(*SK)]
		if v {
			k.Expires = Now.Add(SessionLife)
			rc = k
		}
	}

	for k, v := range s.ClientTable {
		if v.Expires.Before(Now) {
			delete(s.ClientTable, k)
		}
	}
	for k, v := range s.SessionTable {

		if v.Expires.Before(Now) {
			delete(s.SessionTable, k)
		}
	}

	return rc
}

func (o *ServerStruct) Slave() {
	var lis net.Listener
	var err error
	if o.Certificate != nil && o.CACertificate != nil {
		cp := x509.NewCertPool()
		cp.AppendCertsFromPEM(o.CACertificate)
		config := tls.Config{Certificates: []tls.Certificate{*o.Certificate},
			InsecureSkipVerify: false,
			ClientCAs:          cp,
			ClientAuth:         tls.RequireAndVerifyClientCert,
			ServerName:         "Archcopy",
		}
		config.Rand = rand.Reader
		lis, err = tls.Listen(o.Network, o.Address, &config)

		o.TLS = true
	} else if o.Certificate == nil && o.CACertificate == nil {
		lis, err = net.Listen(o.Network, o.Address)
		o.TLS = false
	} else {
		util.Fatal(errors.New("Inconsistent TLS parameters."), "Starting slave")
	}
	if err != nil {
		util.Fatal(err, "Loading slave certificates")
	}

	o.GRPC = grpc.NewServer()

	o.SessionTable = make(map[SessionKey]*SessionStruct)
	o.ClientTable = make(map[string]*ClientTableStruct)

	go func(o *ServerStruct) {
		for {
			time.Sleep(2 * time.Minute)
			o.Housekeeping(nil)
			if len(o.SessionTable) == 0 {
				if o.SingleSession {
					o.GRPC.GracefulStop()
					fmt.Println("No active sessions -- terminating.")
				}
			}
		}
	}(o)

	o.HaveRoot = util.CheckForRoot()
	Server = o

	pb.RegisterSlaveServer(o.GRPC, o)
	if err := o.GRPC.Serve(lis); err != nil {
		util.Fatal(err, "Starting slave")
	}
}
