// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.20.3
// source: proto/server-stream.proto

package __

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

const (
	ServerStreamService_ServerStreamCall_FullMethodName = "/unary.ServerStreamService/ServerStreamCall"
)

// ServerStreamServiceClient is the client API for ServerStreamService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ServerStreamServiceClient interface {
	ServerStreamCall(ctx context.Context, in *ServerStreamRequest, opts ...grpc.CallOption) (ServerStreamService_ServerStreamCallClient, error)
}

type serverStreamServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewServerStreamServiceClient(cc grpc.ClientConnInterface) ServerStreamServiceClient {
	return &serverStreamServiceClient{cc}
}

func (c *serverStreamServiceClient) ServerStreamCall(ctx context.Context, in *ServerStreamRequest, opts ...grpc.CallOption) (ServerStreamService_ServerStreamCallClient, error) {
	stream, err := c.cc.NewStream(ctx, &ServerStreamService_ServiceDesc.Streams[0], ServerStreamService_ServerStreamCall_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &serverStreamServiceServerStreamCallClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ServerStreamService_ServerStreamCallClient interface {
	Recv() (*ServerStreamResponse, error)
	grpc.ClientStream
}

type serverStreamServiceServerStreamCallClient struct {
	grpc.ClientStream
}

func (x *serverStreamServiceServerStreamCallClient) Recv() (*ServerStreamResponse, error) {
	m := new(ServerStreamResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ServerStreamServiceServer is the server API for ServerStreamService service.
// All implementations must embed UnimplementedServerStreamServiceServer
// for forward compatibility
type ServerStreamServiceServer interface {
	ServerStreamCall(*ServerStreamRequest, ServerStreamService_ServerStreamCallServer) error
	mustEmbedUnimplementedServerStreamServiceServer()
}

// UnimplementedServerStreamServiceServer must be embedded to have forward compatible implementations.
type UnimplementedServerStreamServiceServer struct {
}

func (UnimplementedServerStreamServiceServer) ServerStreamCall(*ServerStreamRequest, ServerStreamService_ServerStreamCallServer) error {
	return status.Errorf(codes.Unimplemented, "method ServerStreamCall not implemented")
}
func (UnimplementedServerStreamServiceServer) mustEmbedUnimplementedServerStreamServiceServer() {}

// UnsafeServerStreamServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ServerStreamServiceServer will
// result in compilation errors.
type UnsafeServerStreamServiceServer interface {
	mustEmbedUnimplementedServerStreamServiceServer()
}

func RegisterServerStreamServiceServer(s grpc.ServiceRegistrar, srv ServerStreamServiceServer) {
	s.RegisterService(&ServerStreamService_ServiceDesc, srv)
}

func _ServerStreamService_ServerStreamCall_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ServerStreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ServerStreamServiceServer).ServerStreamCall(m, &serverStreamServiceServerStreamCallServer{stream})
}

type ServerStreamService_ServerStreamCallServer interface {
	Send(*ServerStreamResponse) error
	grpc.ServerStream
}

type serverStreamServiceServerStreamCallServer struct {
	grpc.ServerStream
}

func (x *serverStreamServiceServerStreamCallServer) Send(m *ServerStreamResponse) error {
	return x.ServerStream.SendMsg(m)
}

// ServerStreamService_ServiceDesc is the grpc.ServiceDesc for ServerStreamService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ServerStreamService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "unary.ServerStreamService",
	HandlerType: (*ServerStreamServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ServerStreamCall",
			Handler:       _ServerStreamService_ServerStreamCall_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/server-stream.proto",
}