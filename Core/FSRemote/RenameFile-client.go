package rpcmaster

import (
	rpc "Archcopy/RPC"
	"context"
	"crypto/subtle"
	"errors"
	"time"

	pb "proto.local/archcopyRPC"
)

func (o *FSAccessRemoteStruct) RenameFile(Source string, Destination string) error {
	var rp pb.RenamePackage

	rp.SessionKey = o.SessionKey
	rp.Source = []byte(Source)
	rp.Destination = []byte(Destination)
	rp.Signature = rpc.Sign(rp.SessionKey, rp.Source, rp.Destination)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	Status, err := o.slave.RenameFile(ctx, &rp)

	if err != nil {
		return err
	}

	if subtle.ConstantTimeCompare(Status.Signature, rpc.Sign(Status.Status)) == 0 {
		return errors.New("Signature check failed for status message.")
	}

	if Status.Status == 0 {
		return nil
	}

	return err
}
