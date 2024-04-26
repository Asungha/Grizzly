package Grizzly

import (
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type IServerStream[I protoreflect.ProtoMessage, O protoreflect.ProtoMessage] interface {
	SendAndClose(O) error
	Recv() (I, error)
	grpc.ServerStream
}

type IServerStreamContext interface {
	IServerStreamContextimpl()
	IContextimpl()
	Test() string
}

type ServerStreamContext[I protoreflect.ProtoMessage, O protoreflect.ProtoMessage] struct {
	stream  IServerStream[I, O]
	Context *Context
}

func (s *ServerStreamContext[I, O]) IServerStreamContextimpl() {}

func (s *ServerStreamContext[I, O]) IContextimpl() {}

func (s *ServerStreamContext[I, O]) SendAndClose(resp O) error {
	return s.stream.SendAndClose(resp)
}

func (s *ServerStreamContext[I, O]) Recv() (I, error) {
	return s.stream.Recv()
}

func NewServerStreamContext[I protoreflect.ProtoMessage, O protoreflect.ProtoMessage](stream IServerStream[I, O]) *ServerStreamContext[I, O] {
	return &ServerStreamContext[I, O]{stream: stream}
}
