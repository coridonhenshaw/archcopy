package rpcslave

import (
	rpc "Archcopy/RPC"
	"context"
	"crypto/subtle"
	"errors"
	"os"

	pb "proto.local/archcopyRPC"
)

func (s *ServerStruct) RenameFile(ctx context.Context, in *pb.RenamePackage) (*pb.Status, error) {
	var rc pb.Status
	if s.Housekeeping(&in.SessionKey) == nil {
		return &rc, errors.New("Session failure.")
	}

	if subtle.ConstantTimeCompare(in.Signature,
		rpc.Sign(in.SessionKey, in.Source, in.Destination)) == 0 {

		return &rc, errors.New("Signature check failed.")
	}

	err := os.Rename(string(in.Source), string(in.Destination))

	if err != nil {
		rc.Status = 1
		rc.Variant = []byte(err.Error())
	}

	rc.Signature = rpc.Sign(rc.Status)

	return &rc, err
}
