package rpcmaster

import (
	common "Archcopy/Common"
	rpc "Archcopy/RPC"
	streamfile "Archcopy/StreamFile"
	util "Archcopy/util"
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/gob"
	"errors"
	"io/fs"
	"log"
	"os"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/metadata"

	pb "proto.local/archcopyRPC"
)

type RemoteUpperWriteDriver struct {
	Filename string
	fh       *os.File
	Atime    time.Time
	Mtime    time.Time

	slave      pb.SlaveClient
	SessionKey []byte

	StreamValid bool
	Stream      pb.Slave_WriteFileClient
	ctx         context.Context
	cancel      context.CancelFunc

	StreamSend *streamfile.StreamfileSendStruct

	result   common.TransferProgressStruct
	progress chan common.TransferProgressStruct
}

func (o *RemoteUpperWriteDriver) UseReadHash() bool { return true }

func (o *RemoteUpperWriteDriver) CreateFile(DirectoryPerm os.FileMode, Fullname string, perm fs.FileMode,
	Owner *common.OwnershipStruct,
	Atime time.Time, Mtime time.Time,
	Size int64, Sparse bool,
	OnExist common.ExistAction,
	ReadbackHash bool,
	Progress chan common.TransferProgressStruct) error {

	if len(o.SessionKey) == 0 {
		log.Panic("Not configured.")
	}

	o.progress = Progress

	var err error

	o.ctx, o.cancel = context.WithCancel(context.Background())

	var UID int
	var GID int

	if Owner != nil {
		UID = Owner.UID
		GID = Owner.GID
	}

	fmd := rpc.FileMetadataStruct{Filename: Fullname, Permissions: uint64(perm), UID: UID,
		GID: GID, ATime: Atime.Unix(), MTime: Mtime.Unix(), Size: Size, Sparse: Sparse,
		OnExist: OnExist, ReadbackHash: ReadbackHash}

	x, err := fmd.Encode()
	fmd.Signature = rpc.Sign([]byte(x))

	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(fmd)

	md := metadata.Pairs("sessionkey-bin", string(o.SessionKey), "gob-bin", buf.String())

	for k, v := range md {
		o.ctx = metadata.AppendToOutgoingContext(o.ctx, k, v[0])
	}

	o.Stream, err = o.slave.WriteFile(o.ctx)

	if err != nil {
		util.Fatal(err, "Writing file to remote")
	}

	o.StreamValid = true

	o.StreamSend = &streamfile.StreamfileSendStruct{}
	o.StreamSend.WriteFunc = func(In *streamfile.CompressedBlockStruct) error {
		var File pb.File
		File.Data = In.Data
		File.Compressed = In.Compressed
		File.Zero = In.Zero
		atomic.AddUint64(&o.result.TransferredBytesCompressed, uint64(len(File.Data)))
		atomic.AddUint64(&o.result.TransferredBytes, In.DecompressedSize)
		o.progress <- o.result

		Ctrl := make(chan error, 0)

		Timeout := time.NewTimer(1 * time.Minute)
		TimeoutChan := Timeout.C

		go func() {
			err := o.Stream.Send(&File)
			Ctrl <- err
		}()

		select {
		case err = <-Ctrl:
		case <-TimeoutChan:
			err = errors.New("Send timeout")
			o.cancel()
		}
		Timeout.Stop()

		return err
	}

	o.StreamSend.Initialize()

	return nil
}

func (o *RemoteUpperWriteDriver) Write(Buffer *[]byte) error {

	if len(o.SessionKey) == 0 || !o.StreamValid {
		log.Panic("Not configured.")
	}

	err := o.StreamSend.In(Buffer)

	if err != nil {
		o.Stream.CloseSend()
		o.ctx.Done()
		o.StreamValid = false
	}

	return err
}

func (o *RemoteUpperWriteDriver) Finalize(Cancel bool) (common.TransferProgressStruct, error) {
	var err error

	if !o.StreamValid {
		return o.result, errors.New("Not valid")
	}

	if len(o.SessionKey) == 0 {
		log.Panic("Not configured.")
	}

	err1 := o.StreamSend.Await()

	Status, err := o.Stream.CloseAndRecv()

	if err1 != nil {
		err = err1
	}

	o.ctx.Done()
	o.StreamValid = false

	if Status == nil {
		return o.result, err
	}

	Expected := Status.Signature
	Status.Signature = nil

	if subtle.ConstantTimeCompare(Expected, rpc.Sign(Status.Status)) == 0 {
		err = errors.New("Signature check failed for status message.")
		return o.result, err
	}

	o.result.ReceivedHash = string(Status.ReceivedHash)
	o.result.ReadbackHash = string(Status.ReadbackHash)

	o.result.Final = true
	o.progress <- o.result

	if Status.Status != 0 {
		log.Printf("Status == %v, err = %v", Status.Status, string(Status.Error))
		err = errors.New(string(Status.Error))
	}
	return o.result, err
}
