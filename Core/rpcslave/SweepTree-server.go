package rpcslave

import (
	walktree "Archcopy/FSLocal/WalkTree"
	"Archcopy/HashFile"
	rpc "Archcopy/RPC"
	util "Archcopy/util"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"

	pb "proto.local/archcopyRPC"
)

func (o *ServerStruct) SweepTree(in *pb.SweepPackage, out pb.Slave_SweepTreeServer) error {
	var err error

	Session := o.Housekeeping(&in.SessionKey)

	if Session == nil {
		err = errors.New("Session failure.")
		fmt.Println("err1", err)
		return err
	}

	if subtle.ConstantTimeCompare(in.Signature, rpc.Sign(in.StartDirectory, in.FollowSymlinks, in.SessionKey)) == 0 {
		err = errors.New("Signature check failed.")
		fmt.Println("err2", err)
		return err
	}

	fmt.Printf("SweepTree (%v): %v\n", util.Base64(Session.ClientID), string(in.StartDirectory))

	WalkTreeCallback := WalkTreeCallbackStruct{Out: &out}
	WalkTree := walktree.WalkTreeStruct{Callback: &WalkTreeCallback, FollowSymlinks: in.FollowSymlinks}

	WalkTree.Sweep(string(in.StartDirectory))

	WalkTreeCallback.Pkg.Signature = rpc.Sign(WalkTreeCallback.Pkg.Directory, WalkTreeCallback.Pkg.Files)
	out.Send(WalkTreeCallback.Pkg)

	err = WalkTreeCallback.err

	return err
}

type WalkTreeCallbackStruct struct {
	Pkg            *pb.SweepPackageReply
	Out            *pb.Slave_SweepTreeServer
	err            error
	GenerateHashes bool
}

func (o *WalkTreeCallbackStruct) NonFileObject(p string, f os.FileInfo) {
	return
}

func (o *WalkTreeCallbackStruct) File(p string, f os.FileInfo) {

	var HashBytes []byte
	if o.GenerateHashes {
		Hash, _ := HashFile.HashFile(p, nil)
		HashBytes, _ = hex.DecodeString(Hash)
	}

	F := pb.SweepPackageFile{Filename: []byte(f.Name()), Size: f.Size(), Hash: HashBytes}

	(*o.Pkg).Files = append((*o.Pkg).Files, &F)
}

func (o *WalkTreeCallbackStruct) PreDirectory(p string) bool {
	// return false to halt sweep
	return true
}

func (o *WalkTreeCallbackStruct) Directory(p string) {
	if o.Pkg != nil {
		o.Pkg.Signature = rpc.Sign(o.Pkg.Directory, o.Pkg.Files)
		(*o.Out).Send(o.Pkg)
	}
	o.Pkg = &pb.SweepPackageReply{Directory: []byte(p)}
	return
}

func (o *WalkTreeCallbackStruct) DirectoryError(Fullname string, err error) bool {
	// return false to halt sweep
	if errors.Is(err, fs.ErrPermission) {
		return true
	}

	if errors.Is(err, fs.ErrNotExist) {
		return true
	}

	return false
}
