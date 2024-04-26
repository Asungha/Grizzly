package Grizzly

var USE_INPUT_VALIDATION = GetEnv("USE_INPUT_VALIDATION", Optional("true")).Bool.Get()
var USE_OUTPUT_VALIDATION = GetEnv("USE_OUTPUT_VALIDATION", Optional("true")).Bool.Get()

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
