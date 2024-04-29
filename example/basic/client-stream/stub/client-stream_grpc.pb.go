// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.20.3
// source: proto/client-stream.proto

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
	ClientStreamService_ClientStreamCall_FullMethodName = "/unary.ClientStreamService/ClientStreamCall"
)

// ClientStreamServiceClient is the client API for ClientStreamService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ClientStreamServiceClient interface {
	ClientStreamCall(ctx context.Context, opts ...grpc.CallOption) (ClientStreamService_ClientStreamCallClient, error)
}

type clientStreamServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewClientStreamServiceClient(cc grpc.ClientConnInterface) ClientStreamServiceClient {
	return &clientStreamServiceClient{cc}
}

func (c *clientStreamServiceClient) ClientStreamCall(ctx context.Context, opts ...grpc.CallOption) (ClientStreamService_ClientStreamCallClient, error) {
	stream, err := c.cc.NewStream(ctx, &ClientStreamService_ServiceDesc.Streams[0], ClientStreamService_ClientStreamCall_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &clientStreamServiceClientStreamCallClient{stream}
	return x, nil
}

type ClientStreamService_ClientStreamCallClient interface {
	Send(*ClientStreamRequest) error
	CloseAndRecv() (*ClientStreamResponse, error)
	grpc.ClientStream
}

type clientStreamServiceClientStreamCallClient struct {
	grpc.ClientStream
}

func (x *clientStreamServiceClientStreamCallClient) Send(m *ClientStreamRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *clientStreamServiceClientStreamCallClient) CloseAndRecv() (*ClientStreamResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(ClientStreamResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ClientStreamServiceServer is the server API for ClientStreamService service.
// All implementations must embed UnimplementedClientStreamServiceServer
// for forward compatibility
type ClientStreamServiceServer interface {
	ClientStreamCall(ClientStreamService_ClientStreamCallServer) error
	mustEmbedUnimplementedClientStreamServiceServer()
}

// UnimplementedClientStreamServiceServer must be embedded to have forward compatible implementations.
type UnimplementedClientStreamServiceServer struct {
}

func (UnimplementedClientStreamServiceServer) ClientStreamCall(ClientStreamService_ClientStreamCallServer) error {
	return status.Errorf(codes.Unimplemented, "method ClientStreamCall not implemented")
}
func (UnimplementedClientStreamServiceServer) mustEmbedUnimplementedClientStreamServiceServer() {}

// UnsafeClientStreamServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ClientStreamServiceServer will
// result in compilation errors.
type UnsafeClientStreamServiceServer interface {
	mustEmbedUnimplementedClientStreamServiceServer()
}

func RegisterClientStreamServiceServer(s grpc.ServiceRegistrar, srv ClientStreamServiceServer) {
	s.RegisterService(&ClientStreamService_ServiceDesc, srv)
}

func _ClientStreamService_ClientStreamCall_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ClientStreamServiceServer).ClientStreamCall(&clientStreamServiceClientStreamCallServer{stream})
}

type ClientStreamService_ClientStreamCallServer interface {
	SendAndClose(*ClientStreamResponse) error
	Recv() (*ClientStreamRequest, error)
	grpc.ServerStream
}

type clientStreamServiceClientStreamCallServer struct {
	grpc.ServerStream
}

func (x *clientStreamServiceClientStreamCallServer) SendAndClose(m *ClientStreamResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *clientStreamServiceClientStreamCallServer) Recv() (*ClientStreamRequest, error) {
	m := new(ClientStreamRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ClientStreamService_ServiceDesc is the grpc.ServiceDesc for ClientStreamService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ClientStreamService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "unary.ClientStreamService",
	HandlerType: (*ClientStreamServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ClientStreamCall",
			Handler:       _ClientStreamService_ClientStreamCall_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "proto/client-stream.proto",
}