package rpcmaster

import (
	rpc "Archcopy/RPC"
	util "Archcopy/util"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"errors"

	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	pb "proto.local/archcopyRPC"
)

func (o *FSAccessRemoteStruct) Connect(ConnStr rpc.RemoteConnectStruct) error {
	var err error
	// func WithContextDialer(f func(context.Context, string) (net.Conn, error)) DialOption

	opts := []grpc.DialOption{}

	if ConnStr.Dialer != nil {
		opts = append(opts, grpc.WithContextDialer(ConnStr.Dialer))
	}

	if rpc.Credentials.CACert != nil && rpc.ClientCert != nil {
		cp := x509.NewCertPool()
		cp.AppendCertsFromPEM(rpc.Credentials.CACert)
		config := &tls.Config{
			Certificates:       []tls.Certificate{*rpc.ClientCert},
			InsecureSkipVerify: false,
			RootCAs:            cp,
			ServerName:         "Archcopy",
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config)))
		o.conn, err = grpc.Dial(ConnStr.Dial, opts...)
		o.TLS = true

	} else if rpc.Credentials.CACert == nil && rpc.ClientCert == nil {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		o.conn, err = grpc.Dial(ConnStr.Dial, opts...)
	} else {
		util.Fatal(errors.New("Inconsistent TLS configuration."), "Starting master")
	}

	if err != nil {
		util.Fatal(err, "Starting master")
	}

	o.slave = pb.NewSlaveClient(o.conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*20)
	defer cancel()

	var ClientID = make([]byte, 32)
	rand.Read(ClientID)
	var Response []byte

	{
		var Credentials pb.ConnectInquire
		Credentials.ClientID = ClientID
		Credentials.OfferNonce = make([]byte, 64)
		rand.Read(Credentials.OfferNonce)
		Credentials.OfferSignature = rpc.Sign(Credentials.OfferNonce)

		Response1, err := o.slave.Connect(ctx, &Credentials)
		util.Fatal(err, "Failed to connect to remote")
		if len(Response1.Challenge) != 64 {
			util.Fatal(errors.New("Unexpected challenge length. Remote may have declined PSK."), "Connecting to slave")
		}
		if Response1.Phase != 0 {
			util.Fatal(errors.New("Connection phase error"), "Connecting to slave")
		}

		if subtle.ConstantTimeCompare(Response1.NonceSignature, rpc.Sign(Credentials.OfferNonce)) == 0 ||
			subtle.ConstantTimeCompare(Response1.ChallengeSignature, rpc.Sign(Response1.Challenge)) == 0 {
			util.Fatal(errors.New("Remote accepted a PSK that it does not have. Remote may be compromised."), "Connecting to slave")
		}

		Response = rpc.Sign(Response1.Challenge)
	}

	{
		var Credentials pb.ConnectInquire
		Credentials.ChallengeResponse = Response
		Credentials.ClientID = ClientID
		Credentials.Phase = 1
		Response2, err := o.slave.Connect(ctx, &Credentials)
		util.Fatal(err, "")

		if Response2.Phase != 1 {
			util.Fatal(errors.New("Connection phase error"), "Connecting to slave")
		}

		if len(Response2.Key) > 0 {
			o.SessionKey = Response2.Key
		} else {
			util.Fatal(errors.New("Remote did not offer a session key"), "Connecting to slave")
		}
	}

	return nil
}
