package rpcslave

import (
	rpc "Archcopy/RPC"
	"context"
	"crypto/subtle"
	"errors"
	"os"

	pb "proto.local/archcopyRPC"
)

func (s *ServerStruct) CheckFiles(ctx context.Context, In *pb.FilePackage) (*pb.FilePackageReply, error) {
	var Reply = new(pb.FilePackageReply)

	if s.Housekeeping(&In.SessionKey.Key) == nil {
		return Reply, errors.New("Session failure.")
	}

	Expected := In.Signature
	In.Signature = nil

	Candidate := rpc.Sign(In.Filenames)

	if subtle.ConstantTimeCompare(Expected, Candidate) == 0 {
		return Reply, errors.New("Signature check failed.")
	}

	Reply.Existing = make(map[uint32]uint64)

	for i, v := range In.Filenames {
		stat, err := os.Stat(string(v))
		if err == nil {
			Reply.Existing[uint32(i)] = uint64(stat.Size())
		}
	}

	return Reply, nil
}
