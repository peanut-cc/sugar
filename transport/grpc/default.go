package grpc

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/peanut-cc/sugar/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime"
)

// DefaultErrorEncoder is default error encoder.
func DefaultErrorEncoder(err error) error {
	se, ok := err.(*errors.StatusError)
	if !ok {
		se = &errors.StatusError{
			Code:    2,
			Reason:  "Unknown",
			Message: "Unknown: " + err.Error(),
		}
	}
	gs := status.Newf(codes.Code(se.Code), "%s: %s", se.Reason, se.Message)
	details := []proto.Message{
		&errdetails.ErrorInfo{
			Reason:   se.Reason,
			Metadata: map[string]string{"message": se.Message},
		},
	}
	gs, err = gs.WithDetails(details...)
	if err != nil {
		return err
	}
	return gs.Err()
}

// DefaultErrorDecoder is default error decoder.
func DefaultErrorDecoder(err error) error {
	gs := status.Convert(err)
	for _, detail := range gs.Details() {
		switch d := detail.(type) {
		case *errdetails.ErrorInfo:
			return &errors.StatusError{
				Code:    int32(gs.Code()),
				Reason:  d.Reason,
				Message: d.Metadata["message"],
			}
		}
	}
	return &errors.StatusError{Code: int32(gs.Code())}
}

// DefaultRecoveryHandler is default recovery handler.
func DefaultRecoveryHandler(ctx context.Context, req, err interface{}) error {
	buf := make([]byte, 65536)
	n := runtime.Stack(buf, false)
	buf = buf[:n]
	fmt.Printf("panic: %v %v\nstack: %s\n", req, err, buf)
	return errors.Unknown("Unknown", "panic triggered: %v", err)
}
