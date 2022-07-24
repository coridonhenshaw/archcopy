package rpcslave

import (
	common "Archcopy/Common"
	"Archcopy/FSLocal"
	rpc "Archcopy/RPC"
	streamfile "Archcopy/StreamFile"
	util "Archcopy/util"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/metadata"
	pb "proto.local/archcopyRPC"
)

func (o *ServerStruct) ReadFile(in *pb.Filename, Stream pb.Slave_ReadFileServer) error {
	var err error
	var f FSLocal.FSAccessLocalStruct

	Session := o.Housekeeping(&in.SessionKey)

	if Session == nil {
		err := errors.New("Session failure.")
		fmt.Println("err1", err)
		return err
	}

	if subtle.ConstantTimeCompare(in.Signature, rpc.Sign(in.Filename, in.Offset, in.SessionKey)) == 0 {
		err := errors.New("Signature check failed.")
		fmt.Println("err2", err)
		return err
	}

	fmt.Printf("ReadFile (%v): %v\n", util.Base64(Session.ClientID), string(in.Filename))

	Lro := f.GetFileReader()

	ProgressChannel := make(chan common.TransferProgressStruct)
	go func() {
		for {
			q := <-ProgressChannel
			if q.Final {
				return
			}
		}
	}()

	var FMD1 common.FilemetadataStruct
	FMD1, err = Lro.Open(string(in.Filename), true, in.Offset, ProgressChannel)
	if err != nil {
		return err
	}

	StreamSend := &streamfile.StreamfileSendStruct{}
	StreamSend.WriteFunc = func(In *streamfile.CompressedBlockStruct) error {
		var File pb.File
		File.Data = In.Data
		File.Compressed = In.Compressed
		File.Zero = In.Zero
		// atomic.AddUint64(&o.result.TransferredBytesCompressed, uint64(len(File.Data)))
		// atomic.AddUint64(&o.result.TransferredBytes, In.DecompressedSize)
		// o.progress <- o.result

		Ctrl := make(chan error, 0)

		Timeout := time.NewTimer(1 * time.Minute)
		TimeoutChan := Timeout.C

		go func() {
			err := Stream.Send(&File)
			Ctrl <- err
		}()

		select {
		case err = <-Ctrl:
		case <-TimeoutChan:
			err = errors.New("Send timeout")
		}
		Timeout.Stop()

		return err
	}

	fmd := rpc.FileMetadataStruct{Permissions: uint64(FMD1.Permissions), UID: FMD1.Owner.UID,
		GID: FMD1.Owner.GID, ATime: FMD1.Atime.Unix(), MTime: FMD1.Mtime.Unix(), Size: FMD1.Size}

	x, err := fmd.Encode()
	fmd.Signature = rpc.Sign([]byte(x))
	x, err = fmd.Encode()

	header := metadata.Pairs("metadata", x)
	Stream.SendHeader(header)

	StreamSend.Initialize()

	for {
		var Block []byte
		Block, err = Lro.Read()
		if len(Block) == 0 || err != nil {
			break
		}
		err = StreamSend.In(&Block)
		if err != nil {
			break
		}
	}

	err = StreamSend.Await()

	tps, err1 := Lro.Close()

	if err != nil {
		return err
	} else if err1 != nil {
		return err1
	}

	trailer := metadata.Pairs("hash", tps.ReceivedHash)
	trailer.Append("signature-bin", string(rpc.Sign([]byte(tps.ReceivedHash))))
	Stream.SetTrailer(trailer)

	return nil
}
