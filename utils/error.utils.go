package grez

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Error(err error, code codes.Code) error {
	return status.Error(code, err.Error())
}

func ServiceError(err error) error {
	return status.Error(codes.Internal, "Service error: "+err.Error())
}

func InvalidArgumentError(err error) error {
	return status.Error(codes.InvalidArgument, "Input data invalid: "+err.Error())
}

func DataLossError(err error) error {
	return status.Error(codes.DataLoss, "Service response invalid: "+err.Error())
}

func NotFoundError() error {
	return status.Error(codes.NotFound, "Data not found")
}

func InternalError(err error) error {
	return status.Error(codes.Internal, "Internal error: "+err.Error())
}
