// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.9.2
// source: archcopy.proto

package archcopyRPC

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// SlaveClient is the client API for Slave service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SlaveClient interface {
	Connect(ctx context.Context, in *ConnectInquire, opts ...grpc.CallOption) (*ConnectResponse, error)
	SweepTree(ctx context.Context, in *SweepPackage, opts ...grpc.CallOption) (Slave_SweepTreeClient, error)
	CheckFiles(ctx context.Context, in *FilePackage, opts ...grpc.CallOption) (*FilePackageReply, error)
	WriteFile(ctx context.Context, opts ...grpc.CallOption) (Slave_WriteFileClient, error)
	ReadHash(ctx context.Context, in *Filename, opts ...grpc.CallOption) (*Hash, error)
	ReadFile(ctx context.Context, in *Filename, opts ...grpc.CallOption) (Slave_ReadFileClient, error)
	RenameFile(ctx context.Context, in *RenamePackage, opts ...grpc.CallOption) (*Status, error)
	Disconnect(ctx context.Context, in *SessionKey, opts ...grpc.CallOption) (*Status, error)
}

type slaveClient struct {
	cc grpc.ClientConnInterface
}

func NewSlaveClient(cc grpc.ClientConnInterface) SlaveClient {
	return &slaveClient{cc}
}

func (c *slaveClient) Connect(ctx context.Context, in *ConnectInquire, opts ...grpc.CallOption) (*ConnectResponse, error) {
	out := new(ConnectResponse)
	err := c.cc.Invoke(ctx, "/archcopyRPC.Slave/Connect", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *slaveClient) SweepTree(ctx context.Context, in *SweepPackage, opts ...grpc.CallOption) (Slave_SweepTreeClient, error) {
	stream, err := c.cc.NewStream(ctx, &Slave_ServiceDesc.Streams[0], "/archcopyRPC.Slave/SweepTree", opts...)
	if err != nil {
		return nil, err
	}
	x := &slaveSweepTreeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Slave_SweepTreeClient interface {
	Recv() (*SweepPackageReply, error)
	grpc.ClientStream
}

type slaveSweepTreeClient struct {
	grpc.ClientStream
}

func (x *slaveSweepTreeClient) Recv() (*SweepPackageReply, error) {
	m := new(SweepPackageReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *slaveClient) CheckFiles(ctx context.Context, in *FilePackage, opts ...grpc.CallOption) (*FilePackageReply, error) {
	out := new(FilePackageReply)
	err := c.cc.Invoke(ctx, "/archcopyRPC.Slave/CheckFiles", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *slaveClient) WriteFile(ctx context.Context, opts ...grpc.CallOption) (Slave_WriteFileClient, error) {
	stream, err := c.cc.NewStream(ctx, &Slave_ServiceDesc.Streams[1], "/archcopyRPC.Slave/WriteFile", opts...)
	if err != nil {
		return nil, err
	}
	x := &slaveWriteFileClient{stream}
	return x, nil
}

type Slave_WriteFileClient interface {
	Send(*File) error
	CloseAndRecv() (*WriteFileStatus, error)
	grpc.ClientStream
}

type slaveWriteFileClient struct {
	grpc.ClientStream
}

func (x *slaveWriteFileClient) Send(m *File) error {
	return x.ClientStream.SendMsg(m)
}

func (x *slaveWriteFileClient) CloseAndRecv() (*WriteFileStatus, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(WriteFileStatus)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *slaveClient) ReadHash(ctx context.Context, in *Filename, opts ...grpc.CallOption) (*Hash, error) {
	out := new(Hash)
	err := c.cc.Invoke(ctx, "/archcopyRPC.Slave/ReadHash", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *slaveClient) ReadFile(ctx context.Context, in *Filename, opts ...grpc.CallOption) (Slave_ReadFileClient, error) {
	stream, err := c.cc.NewStream(ctx, &Slave_ServiceDesc.Streams[2], "/archcopyRPC.Slave/ReadFile", opts...)
	if err != nil {
		return nil, err
	}
	x := &slaveReadFileClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Slave_ReadFileClient interface {
	Recv() (*File, error)
	grpc.ClientStream
}

type slaveReadFileClient struct {
	grpc.ClientStream
}

func (x *slaveReadFileClient) Recv() (*File, error) {
	m := new(File)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *slaveClient) RenameFile(ctx context.Context, in *RenamePackage, opts ...grpc.CallOption) (*Status, error) {
	out := new(Status)
	err := c.cc.Invoke(ctx, "/archcopyRPC.Slave/RenameFile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *slaveClient) Disconnect(ctx context.Context, in *SessionKey, opts ...grpc.CallOption) (*Status, error) {
	out := new(Status)
	err := c.cc.Invoke(ctx, "/archcopyRPC.Slave/Disconnect", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SlaveServer is the server API for Slave service.
// All implementations must embed UnimplementedSlaveServer
// for forward compatibility
type SlaveServer interface {
	Connect(context.Context, *ConnectInquire) (*ConnectResponse, error)
	SweepTree(*SweepPackage, Slave_SweepTreeServer) error
	CheckFiles(context.Context, *FilePackage) (*FilePackageReply, error)
	WriteFile(Slave_WriteFileServer) error
	ReadHash(context.Context, *Filename) (*Hash, error)
	ReadFile(*Filename, Slave_ReadFileServer) error
	RenameFile(context.Context, *RenamePackage) (*Status, error)
	Disconnect(context.Context, *SessionKey) (*Status, error)
	mustEmbedUnimplementedSlaveServer()
}

// UnimplementedSlaveServer must be embedded to have forward compatible implementations.
type UnimplementedSlaveServer struct {
}

func (UnimplementedSlaveServer) Connect(context.Context, *ConnectInquire) (*ConnectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Connect not implemented")
}
func (UnimplementedSlaveServer) SweepTree(*SweepPackage, Slave_SweepTreeServer) error {
	return status.Errorf(codes.Unimplemented, "method SweepTree not implemented")
}
func (UnimplementedSlaveServer) CheckFiles(context.Context, *FilePackage) (*FilePackageReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckFiles not implemented")
}
func (UnimplementedSlaveServer) WriteFile(Slave_WriteFileServer) error {
	return status.Errorf(codes.Unimplemented, "method WriteFile not implemented")
}
func (UnimplementedSlaveServer) ReadHash(context.Context, *Filename) (*Hash, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReadHash not implemented")
}
func (UnimplementedSlaveServer) ReadFile(*Filename, Slave_ReadFileServer) error {
	return status.Errorf(codes.Unimplemented, "method ReadFile not implemented")
}
func (UnimplementedSlaveServer) RenameFile(context.Context, *RenamePackage) (*Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RenameFile not implemented")
}
func (UnimplementedSlaveServer) Disconnect(context.Context, *SessionKey) (*Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Disconnect not implemented")
}
func (UnimplementedSlaveServer) mustEmbedUnimplementedSlaveServer() {}

// UnsafeSlaveServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SlaveServer will
// result in compilation errors.
type UnsafeSlaveServer interface {
	mustEmbedUnimplementedSlaveServer()
}

func RegisterSlaveServer(s grpc.ServiceRegistrar, srv SlaveServer) {
	s.RegisterService(&Slave_ServiceDesc, srv)
}

func _Slave_Connect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConnectInquire)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SlaveServer).Connect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/archcopyRPC.Slave/Connect",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SlaveServer).Connect(ctx, req.(*ConnectInquire))
	}
	return interceptor(ctx, in, info, handler)
}

func _Slave_SweepTree_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SweepPackage)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SlaveServer).SweepTree(m, &slaveSweepTreeServer{stream})
}

type Slave_SweepTreeServer interface {
	Send(*SweepPackageReply) error
	grpc.ServerStream
}

type slaveSweepTreeServer struct {
	grpc.ServerStream
}

func (x *slaveSweepTreeServer) Send(m *SweepPackageReply) error {
	return x.ServerStream.SendMsg(m)
}

func _Slave_CheckFiles_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FilePackage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SlaveServer).CheckFiles(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/archcopyRPC.Slave/CheckFiles",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SlaveServer).CheckFiles(ctx, req.(*FilePackage))
	}
	return interceptor(ctx, in, info, handler)
}

func _Slave_WriteFile_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SlaveServer).WriteFile(&slaveWriteFileServer{stream})
}

type Slave_WriteFileServer interface {
	SendAndClose(*WriteFileStatus) error
	Recv() (*File, error)
	grpc.ServerStream
}

type slaveWriteFileServer struct {
	grpc.ServerStream
}

func (x *slaveWriteFileServer) SendAndClose(m *WriteFileStatus) error {
	return x.ServerStream.SendMsg(m)
}

func (x *slaveWriteFileServer) Recv() (*File, error) {
	m := new(File)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Slave_ReadHash_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Filename)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SlaveServer).ReadHash(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/archcopyRPC.Slave/ReadHash",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SlaveServer).ReadHash(ctx, req.(*Filename))
	}
	return interceptor(ctx, in, info, handler)
}

func _Slave_ReadFile_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Filename)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SlaveServer).ReadFile(m, &slaveReadFileServer{stream})
}

type Slave_ReadFileServer interface {
	Send(*File) error
	grpc.ServerStream
}

type slaveReadFileServer struct {
	grpc.ServerStream
}

func (x *slaveReadFileServer) Send(m *File) error {
	return x.ServerStream.SendMsg(m)
}

func _Slave_RenameFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RenamePackage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SlaveServer).RenameFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/archcopyRPC.Slave/RenameFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SlaveServer).RenameFile(ctx, req.(*RenamePackage))
	}
	return interceptor(ctx, in, info, handler)
}

func _Slave_Disconnect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SessionKey)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SlaveServer).Disconnect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/archcopyRPC.Slave/Disconnect",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SlaveServer).Disconnect(ctx, req.(*SessionKey))
	}
	return interceptor(ctx, in, info, handler)
}

// Slave_ServiceDesc is the grpc.ServiceDesc for Slave service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Slave_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "archcopyRPC.Slave",
	HandlerType: (*SlaveServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Connect",
			Handler:    _Slave_Connect_Handler,
		},
		{
			MethodName: "CheckFiles",
			Handler:    _Slave_CheckFiles_Handler,
		},
		{
			MethodName: "ReadHash",
			Handler:    _Slave_ReadHash_Handler,
		},
		{
			MethodName: "RenameFile",
			Handler:    _Slave_RenameFile_Handler,
		},
		{
			MethodName: "Disconnect",
			Handler:    _Slave_Disconnect_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SweepTree",
			Handler:       _Slave_SweepTree_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "WriteFile",
			Handler:       _Slave_WriteFile_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "ReadFile",
			Handler:       _Slave_ReadFile_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "archcopy.proto",
}
