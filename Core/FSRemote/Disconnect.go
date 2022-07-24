package rpcmaster

import (
	rpc "Archcopy/RPC"
	"context"
	"time"

	pb "proto.local/archcopyRPC"
)

func (o *FSAccessRemoteStruct) Disconnect() error {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	SK := pb.SessionKey{Key: o.SessionKey}
	SK.Signature = rpc.Sign(SK.Key)

	_, err = o.slave.Disconnect(ctx, &SK)

	return err
}
