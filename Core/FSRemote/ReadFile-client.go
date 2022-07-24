package rpcmaster

import (
	common "Archcopy/Common"
	rpc "Archcopy/RPC"
	streamfile "Archcopy/StreamFile"
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/djherbis/times"
	"google.golang.org/grpc/metadata"
	pb "proto.local/archcopyRPC"
)

type ReadFileClientStruct struct {
	Size  int64
	Times times.Timespec
	Owner *common.OwnershipStruct

	fh *os.File

	results   common.TransferProgressStruct
	starttime time.Time
	progress  chan common.TransferProgressStruct

	slave      pb.SlaveClient
	SessionKey []byte

	ctx    context.Context
	cancel context.CancelFunc
	stream pb.Slave_ReadFileClient

	transferChan chan *[]byte
	errorChan    chan error
}

func (o *ReadFileClientStruct) Open(Filename string,
	ReadHash bool,
	Position int64,
	Progress chan common.TransferProgressStruct) (common.FilemetadataStruct, error) {

	var err error
	var rp pb.Filename
	var Retval common.FilemetadataStruct

	rp.SessionKey = o.SessionKey
	rp.Filename = []byte(Filename)
	rp.Offset = Position
	rp.Signature = rpc.Sign(rp.Filename, rp.Offset, rp.SessionKey)

	o.ctx, o.cancel = context.WithCancel(context.Background())

	o.stream, err = o.slave.ReadFile(o.ctx, &rp)
	if err != nil {
		o.cancel()
		return Retval, err
	}

	var Header metadata.MD
	Header, err = o.stream.Header()
	if err != nil {
		o.cancel()
		return Retval, err
	}

	Metadata := rpc.FileMetadataStruct{}

	{
		hmd := Header.Get("metadata")
		if len(hmd) != 1 {
			return Retval, errors.New("Missing header.")
		}
		err = Metadata.Decode(hmd[0])
		if err != nil {
			return Retval, err
		}
		Expected := Metadata.Signature
		Metadata.Signature = nil
		x, _ := Metadata.Encode()
		if subtle.ConstantTimeCompare(Expected, rpc.Sign([]byte(x))) == 0 {
			return Retval, errors.New("Signature check failed.")
		}
	}

	Retval.Atime = time.Unix(Metadata.ATime, 0)
	Retval.Mtime = time.Unix(Metadata.MTime, 0)
	Retval.Size = Metadata.Size
	Retval.Owner = &common.OwnershipStruct{GID: Metadata.GID, UID: Metadata.UID}
	Retval.Permissions = fs.FileMode(Metadata.Permissions)

	o.transferChan = make(chan *[]byte)
	o.errorChan = make(chan error, 1)

	go o.Listener()

	return Retval, err
}

func (o *ReadFileClientStruct) Listener() {
	var err error

	StreamReceive := &streamfile.StreamfileReceiveStruct{}

	HashThread := common.GetHashThread(32)
	HashThread.Start()

	StreamReceive.Initialize()
	StreamReceive.PostDecompFunc = func(dst *[]byte) error {
		HashThread.InputChan <- dst
		return nil
	}
	StreamReceive.WriteFunc = func(Block *[]byte) error {
		o.transferChan <- Block
		return nil
	}

	o.results.StartTime = time.Now()
	var Done bool
	for err == nil && !Done {
		var in *pb.File
		in, err = o.stream.Recv()
		if err == io.EOF {
			err = nil
			Done = true
		}
		if err != nil {
			StreamReceive.Cancel(err)
			break
		}

		err = StreamReceive.In(in)
	}

	if err == nil {
		a, b, c := StreamReceive.Await()
		o.results.TransferredBytesCompressed = uint64(a)
		o.results.TransferredBytes = uint64(b)
		err = c

		o.results.ReceivedHash = HashThread.Await()
	} else {
		StreamReceive.Cancel(err)
		select {
		case o.errorChan <- err:
		default:
		}
	}
	o.stream.Recv()
	Terminate := make([]byte, 0)
	o.transferChan <- &Terminate
	o.results.EndTime = time.Now()
}

func (o *ReadFileClientStruct) Read() ([]byte, error) {
	select {
	case Buf := <-o.transferChan:
		return *Buf, nil
	case err := <-o.errorChan:
		return nil, err
	}
	// unreachable.
}

func (o *ReadFileClientStruct) Close() (common.TransferProgressStruct, error) {

	var err error

	for {
		_, err = o.stream.Recv()
		if err != nil {
			break
		}
		fmt.Println("Waiting for EOF.")
		time.Sleep(1 * time.Second)
	}

	var Trailer metadata.MD
	Trailer = o.stream.Trailer()

	{
		TrailerHashSlice := Trailer.Get("hash")
		if len(TrailerHashSlice) != 1 {
			return o.results, errors.New("Malformed or missing trailer (hash).")
		}
		TrailerHash := TrailerHashSlice[0]

		TrailerSignatureSlice := Trailer.Get("signature-bin")
		if len(TrailerHashSlice) != 1 {
			return o.results, errors.New("Malformed or missing trailer (signature).")
		}
		TrailerSignature := TrailerSignatureSlice[0]

		if subtle.ConstantTimeCompare([]byte(TrailerSignature), rpc.Sign([]byte(TrailerHash))) == 0 {
			return o.results, errors.New("Signature check failed.")
		}

		if subtle.ConstantTimeCompare([]byte(TrailerHash), []byte(o.results.ReceivedHash)) == 0 {
			return o.results, errors.New("Hash check failed.")
		}
	}

	return o.results, nil
}
