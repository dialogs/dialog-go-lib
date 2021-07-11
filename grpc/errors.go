package grpc

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// non-retryable errors
var (
	ErrInvalidRequest   = status.Error(codes.InvalidArgument, "invalid request")
	ErrInternal         = status.Error(codes.Internal, "an error has occurred")
	ErrUnauthenticated  = status.Error(codes.Unauthenticated, "no valid session found")
	ErrUnimplemented    = status.Error(codes.Unimplemented, "unimplemented")
	ErrPermissionDenied = status.Error(codes.PermissionDenied, "forbidden")
	ErrForbidden        = status.Error(codes.PermissionDenied, "forbidden")
	ErrNotFound.        = status.Error(codes.NotFound, "entity not found")
)
