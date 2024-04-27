package controller

import utils "github.com/Asungha/Grizzly/utils"

var USE_INPUT_VALIDATION = utils.GetEnv("USE_INPUT_VALIDATION", utils.Optional("true")).Bool.Get()
var USE_OUTPUT_VALIDATION = utils.GetEnv("USE_OUTPUT_VALIDATION", utils.Optional("true")).Bool.Get()

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
