package Grizzly

import (
	"context"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type IUnaryContext interface {
	IUnaryContextimpl()
}

type UnaryContext[I protoreflect.ProtoMessage, O protoreflect.ProtoMessage] struct {
	Context *context.Context
}

func (s *UnaryContext[I, O]) IUnaryContextimpl() {}

func (s *UnaryContext[I, O]) IContextimpl() {}

func NewUnaryContext[I protoreflect.ProtoMessage, O protoreflect.ProtoMessage](ctx *context.Context) *UnaryContext[I, O] {
	return &UnaryContext[I, O]{Context: ctx}
}
