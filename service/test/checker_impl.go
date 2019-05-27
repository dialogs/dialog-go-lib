package test

import (
	context "context"

	types "github.com/gogo/protobuf/types"
)

// A CheckerImpl is a checker server
type CheckerImpl struct {
}

// NewCheckerImpl creates a new CheckerImpl
func NewCheckerImpl() *CheckerImpl {
	return &CheckerImpl{}
}

// Ping responds to the request
func (c *CheckerImpl) Ping(_ context.Context, _ *types.Empty) (*types.Empty, error) {
	return &types.Empty{}, nil
}
