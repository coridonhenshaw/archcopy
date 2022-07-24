package rpcmaster

import (
	common "Archcopy/Common"

	"google.golang.org/grpc"
	pb "proto.local/archcopyRPC"
)

type FSAccessRemoteStruct struct {
	conn  *grpc.ClientConn
	slave pb.SlaveClient
	TLS   bool

	SessionKey []byte
}

func (o *FSAccessRemoteStruct) GetFileWriter() common.FileWriterInterface {
	var rc RemoteUpperWriteDriver

	rc.slave = o.slave
	rc.SessionKey = o.SessionKey

	return &rc
}

func (o *FSAccessRemoteStruct) GetFileReader() common.FileReaderInterface {
	var rc ReadFileClientStruct

	rc.slave = o.slave
	rc.SessionKey = o.SessionKey

	return &rc
}

func (o *FSAccessRemoteStruct) Open() { return }

func (o *FSAccessRemoteStruct) Close() {
	o.Disconnect()
	o.conn.Close()
	return
}
