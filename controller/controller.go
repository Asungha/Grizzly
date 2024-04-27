package controller

import (
	utils "github.com/Asungha/Grizzly/utils"
	"google.golang.org/protobuf/reflect/protoreflect"

	eventbus "github.com/Asungha/Grizzly/eventbus"
)

var USE_INPUT_VALIDATION = utils.GetEnv("USE_INPUT_VALIDATION", utils.Optional("true")).Bool.Get()
var USE_OUTPUT_VALIDATION = utils.GetEnv("USE_OUTPUT_VALIDATION", utils.Optional("true")).Bool.Get()

type IContoller[
	Req protoreflect.ProtoMessage,
	Res protoreflect.ProtoMessage,
] interface {
	ImpCotroller()

	SubscribeRequestPipe(string) eventbus.IPipe[Req]
	SubscribeResponsePipe(string) eventbus.IPipe[Res]

	PublishRequestPipe(string, Req)
	PublishResponsePipe(string, Res)
}

type Controller[
	Req protoreflect.ProtoMessage,
	Res protoreflect.ProtoMessage,
] struct {
	RequestPipe  map[string]eventbus.IPipe[Req]
	ResponsePipe map[string]eventbus.IPipe[Res]
}

func (c *Controller[
	Req,
	Res,
]) ImpCotroller() {
}

func (c *Controller[Req, Res]) SubscribeRequestPipe(key string) eventbus.IPipe[Req] {
	if c.RequestPipe[key] != nil {
		return c.RequestPipe[key]
	}
	c.RequestPipe[key] = eventbus.
	return c.RequestPipe[key]
}

type FunctionOptions struct {
	inputValidation  bool
	outputValidation bool
}

func FunctionOptionsBuilder() *FunctionOptions {
	return &FunctionOptions{}
}

func (o *FunctionOptions) InputValidation(v bool) *FunctionOptions {
	o.inputValidation = v
	return o
}

func (o *FunctionOptions) OutputValidation(v bool) *FunctionOptions {
	o.outputValidation = v
	return o
}
