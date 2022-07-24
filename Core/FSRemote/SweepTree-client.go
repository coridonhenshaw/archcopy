package rpcmaster

import (
	rpc "Archcopy/RPC"
	"context"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"io"
	"path"

	pb "proto.local/archcopyRPC"
)

func (o *FSAccessRemoteStruct) SweepTree(Path string, FollowSymlinks bool, GenerateHashes bool,
	Setter func(Filename string, Filesize int64, Hash string)) error {

	var rp pb.SweepPackage

	rp.SessionKey = o.SessionKey
	rp.StartDirectory = []byte(Path)
	rp.FollowSymlinks = FollowSymlinks
	rp.Signature = rpc.Sign(rp.StartDirectory, rp.FollowSymlinks, rp.SessionKey)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	Stream, err := o.slave.SweepTree(ctx, &rp)

	if err != nil {
		return err
	}

	var Done bool
	for err == nil && !Done {
		var in *pb.SweepPackageReply
		in, err = Stream.Recv()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			break
		}

		if subtle.ConstantTimeCompare(in.Signature, rpc.Sign(in.Directory, in.Files)) == 0 {
			err = errors.New("Signature check failed.")
			return err
		}

		for _, v := range in.Files {
			var Fn = path.Join(string(in.Directory), string(v.Filename))
			var Hash string
			if GenerateHashes {
				Hash = hex.EncodeToString(v.Hash)
			}
			Setter(Fn, v.Size, Hash)
		}
	}
	return err
}
