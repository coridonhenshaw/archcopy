package rpcslave

import (
	"Archcopy/HashFile"
	rpc "Archcopy/RPC"
	"context"
	"crypto/subtle"
	"errors"

	pb "proto.local/archcopyRPC"
)

func (s *ServerStruct) ReadHash(ctx context.Context, in *pb.Filename) (*pb.Hash, error) {

	var Hash pb.Hash

	if s.Housekeeping(&in.SessionKey) == nil {
		return &Hash, errors.New("Session failure.")
	}

	if subtle.ConstantTimeCompare(in.Signature,
		rpc.Sign(in.SessionKey, in.Filename)) == 0 {
		return &Hash, errors.New("Signature check failed.")
	}

	var err error
	var Hsh string

	Hsh, err = HashFile.HashFile(string(in.Filename), nil)

	Hash.Hash = []byte(Hsh)

	Hash.Signature = rpc.Sign(Hash.Hash)

	return &Hash, err
}
