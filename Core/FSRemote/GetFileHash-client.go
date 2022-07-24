package rpcmaster

import (
	rpc "Archcopy/RPC"
	"context"
	"crypto/subtle"
	"errors"
	"time"

	pb "proto.local/archcopyRPC"
)

func (o *FSAccessRemoteStruct) GetFileHash(Source string) (string, error) {
	var rp pb.Filename

	rp.SessionKey = o.SessionKey
	rp.Filename = []byte(Source)
	rp.Signature = rpc.Sign(rp.SessionKey, rp.Filename)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*240)
	defer cancel()

	Hash, err := o.slave.ReadHash(ctx, &rp)

	if err != nil {
		return "", err
	}

	if subtle.ConstantTimeCompare(Hash.Signature, rpc.Sign(Hash.Hash)) == 0 {
		return "", errors.New("Signature check failed.")
	}

	return string(Hash.Hash), nil
}
