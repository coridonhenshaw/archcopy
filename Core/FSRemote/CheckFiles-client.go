package rpcmaster

import (
	rpc "Archcopy/RPC"
	util "Archcopy/util"
	"context"
	"time"

	pb "proto.local/archcopyRPC"
)

func (o *FSAccessRemoteStruct) CheckExists(Accessor func(i int) string, Setter func(i int, Exists bool, Size uint64)) error {

	var fp pb.FilePackage
	fp.SessionKey = &pb.SessionKey{Key: o.SessionKey}

	var i int
	for i = 0; ; i++ {
		Fn := Accessor(i)
		if len(Fn) == 0 {
			break
		}

		fp.Filenames = append(fp.Filenames, []byte(Fn))
	}

	fp.Signature = rpc.Sign(fp.Filenames)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	fpr, err := o.slave.CheckFiles(ctx, &fp)

	if err != nil {
		util.Fatal(err, "Getting file list from remote")
	}

	for j := 0; j < i; j++ {
		Size, Present := fpr.Existing[uint32(j)]
		if Present {
			Setter(j, true, Size)
		} else {
			Setter(j, false, Size)
		}
	}

	return nil
}
