package rpcslave

import (
	common "Archcopy/Common"
	"Archcopy/FSLocal"
	rpc "Archcopy/RPC"
	streamfile "Archcopy/StreamFile"
	"Archcopy/util"
	"bytes"
	"crypto/subtle"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"time"

	"github.com/dustin/go-humanize"
	"google.golang.org/grpc/metadata"
	pb "proto.local/archcopyRPC"
)

type FileSessionStruct struct {
	rpc.FileMetadataStruct

	Expires time.Time

	UWD FSLocal.LocalUpperWriteDriver

	hashThread *common.HashThreadStruct
	doReadback bool

	StreamReceive *streamfile.StreamfileReceiveStruct

	StartTime time.Time
	EndTime   time.Time
}

func (s *ServerStruct) HandleMetadata(a rpc.FileMetadataStruct) (*FileSessionStruct, error) {

	var err error

	Fss := FileSessionStruct{
		UWD: FSLocal.LocalUpperWriteDriver{},
	}

	Fss.Filename = a.Filename
	Fss.Permissions = a.Permissions
	Fss.ATime = a.ATime
	Fss.MTime = a.MTime
	Fss.Size = a.Size

	var Owner *common.OwnershipStruct
	if s.HaveRoot {
		Owner = new(common.OwnershipStruct)
		Owner.UID = a.UID
		Owner.GID = a.GID
	}

	ProgressChannel := make(chan common.TransferProgressStruct)
	go func() {
		for {
			q := <-ProgressChannel
			if q.Final {
				return
			}
		}
	}()

	Fss.doReadback = a.ReadbackHash

	err = Fss.UWD.CreateFile(0700,
		Fss.Filename,
		fs.FileMode(Fss.Permissions),
		Owner,
		time.Unix(int64(a.ATime), 0),
		time.Unix(int64(a.MTime), 0),
		int64(Fss.Size),
		a.Sparse,
		a.OnExist,
		a.ReadbackHash,
		ProgressChannel)

	if err != nil {
		return nil, err
	}
	return &Fss, err
}

func (s *ServerStruct) ExtractMetadata(md metadata.MD) (*SessionStruct, *rpc.FileMetadataStruct, error) {
	var Session *SessionStruct
	{
		if t, ok := md["sessionkey-bin"]; ok || len(t) > 0 {
			b := []byte(t[0])
			Session = s.Housekeeping(&b)
			if Session == nil {
				return nil, nil, errors.New("Invalid session key.")
			}
		} else {
			return nil, nil, errors.New("Missing session key.")
		}
	}

	var Metadata rpc.FileMetadataStruct
	{
		if t, ok := md["gob-bin"]; ok || len(t) > 0 {
			buf := bytes.NewBuffer([]byte(t[0]))
			dec := gob.NewDecoder(buf)

			if err := dec.Decode(&Metadata); err != nil {
				return nil, nil, errors.New("Failed to decode GOB.")
			}

			var x string
			Expected := Metadata.Signature
			Metadata.Signature = nil
			x, _ = Metadata.Encode()
			if subtle.ConstantTimeCompare(Expected, rpc.Sign([]byte(x))) == 0 {
				return nil, nil, errors.New("Signature check failed.")
			}

		} else {
			return nil, nil, errors.New("Missing metadata.")
		}
	}

	return Session, &Metadata, nil
}

func (s *ServerStruct) WriteFile(instream pb.Slave_WriteFileServer) error {
	var FileSession *FileSessionStruct
	var Session *SessionStruct
	var err error

	md, _ := metadata.FromIncomingContext(instream.Context())

	var Metadata *rpc.FileMetadataStruct
	Session, Metadata, err = s.ExtractMetadata(md)

	FileSession, err = s.HandleMetadata(*Metadata)

	if err != nil {
		fmt.Printf("Failed (%v): %v\n", err, Metadata.Filename)
		return err
	}

	fmt.Printf("WriteFile (%v): %v\n", util.Base64(Session.ClientID), FileSession.Filename)

	FileSession.hashThread = common.GetHashThread(128)
	FileSession.hashThread.Start()

	FileSession.StreamReceive = &streamfile.StreamfileReceiveStruct{}

	FileSession.StreamReceive.Initialize()
	FileSession.StreamReceive.PostDecompFunc = func(dst *[]byte) error {
		FileSession.hashThread.InputChan <- dst
		return nil
	}
	FileSession.StreamReceive.WriteFunc = func(Block *[]byte) error {
		return FileSession.UWD.Write(Block)
	}

	FileSession.StartTime = time.Now()
	var Done bool
	for err == nil && !Done {
		var in *pb.File
		in, err = instream.Recv()
		if err == io.EOF {
			err = nil
			Done = true
		}
		if err != nil {
			FileSession.StreamReceive.Cancel(err)
			break
		}

		err = FileSession.StreamReceive.In(in)

		Session.Touch()
	}

	BytesRcvdCompressed, BytesRcvdDecompressed, olderr := FileSession.StreamReceive.Await()
	FileSession.EndTime = time.Now()

	var Retval pb.WriteFileStatus

	var Result common.TransferProgressStruct

	Result, err = FileSession.UWD.Finalize(err != nil)
	Retval.ReadbackHash = []byte(Result.ReadbackHash)

	if err == nil && olderr != nil {
		err = olderr
	}

	FileSession.hashThread.InputChan <- nil
	Retval.ReceivedHash = []byte(FileSession.hashThread.Await())

	if err == nil && FileSession.doReadback {
		if subtle.ConstantTimeCompare(Retval.ReadbackHash, Retval.ReceivedHash) == 0 {
			err = errors.New("Received hash does not equal readback hash. Probable hardware failure.")
		}
	}

	if err == nil {
		Duration := FileSession.EndTime.Sub(FileSession.StartTime)
		Rate := int(float64(BytesRcvdCompressed) / Duration.Seconds())
		Rate2 := int(float64(BytesRcvdDecompressed) / Duration.Seconds())
		Ratio := float64(BytesRcvdCompressed) / float64(BytesRcvdDecompressed)
		fmt.Printf("Finalize: %v %v/s %v/s %.3v %v %v\n",
			Duration.Truncate(1*time.Millisecond),
			humanize.Bytes(uint64(Rate)),
			humanize.Bytes(uint64(Rate2)),
			Ratio,
			Result.ReadbackHash,
			FileSession.Filename)
	} else {
		fmt.Printf("Failed (%v): %v\n", err, FileSession.Filename)
	}

	if err != nil {
		Retval.Status = 1
		Retval.Error = []byte(err.Error())
	}

	Retval.Signature = rpc.Sign(Retval.Status)
	instream.SendAndClose(&Retval)

	return err
}
