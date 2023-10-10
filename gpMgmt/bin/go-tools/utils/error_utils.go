package utils

import (
	"fmt"

	"google.golang.org/grpc/status"
)

func FormatGrpcError(err error) error {
	if err == nil {
		return nil
	}

	grpcErr, ok := status.FromError(err)
	if ok {
		errorDescription := grpcErr.Message()
		return fmt.Errorf(errorDescription)
	}

	return err
}
