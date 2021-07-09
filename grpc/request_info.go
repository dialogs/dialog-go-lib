package grpc

import (
	"context"

	"github.com/markphelps/optional"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type RequestInfo struct {
	RequestID   string
	TokenHash   string
	Metadata    metadata.MD
	RequestName string
	UserID      optional.String
	Span        opentracing.Span
	Context     context.Context
	Logger      *zap.Logger
}

func (ri *RequestInfo) GetChildSpan(name string) opentracing.Span {
	return opentracing.StartSpan(name, opentracing.ChildOf(ri.Span.Context()))
}

func (ri *RequestInfo) WithSpan(span opentracing.Span) *RequestInfo {
	return &RequestInfo{
		RequestID:   ri.RequestID,
		TokenHash:   ri.TokenHash,
		Metadata:    ri.Metadata,
		RequestName: ri.RequestName,
		UserID:      ri.UserID,
		Span:        span,
		Context:     ri.Context,
		Logger:      ri.Logger,
	}
}

type RequestInfoError struct {
	underlying error
}

func (rie RequestInfoError) Error() string {
	return rie.underlying.Error()
}

func getFromMetadata(md metadata.MD, key string) (string, error) {
	vals := md.Get(key)
	if len(vals) != 1 {
		return "", ErrUnauthenticated
	}
	return vals[0], nil
}

func GetRequestInfo(ctx context.Context, requestName string, log *zap.Logger) (*RequestInfo, error) {
	var err error

	md, found := metadata.FromIncomingContext(ctx)
	if !found {
		return nil, ErrInvalidRequest
	}
	zapReq := zap.String("method", requestName)
	ri := &RequestInfo{}
	ri.Metadata = md
	if ri.RequestID, err = getFromMetadata(md, "x-request-id"); err != nil {
		log.Error("no x-request-id found", zapReq)
		return nil, ErrInvalidRequest
	}
	ri.Logger = log.With(zapReq, zap.String("request_id", ri.RequestID))

	if ri.TokenHash, err = getFromMetadata(md, "x-auth-ticket"); err != nil {
		ri.Logger.Error("no x-auth-ticket found")
		return nil, ErrUnauthenticated
	}
	if v, no := getFromMetadata(md, "x-user-id"); no == nil {
		ri.UserID = optional.NewString(v)
	}
	ri.RequestName = requestName
	ri.Span, ri.Context = opentracing.StartSpanFromContext(ctx, requestName)

	if uid, no := ri.UserID.Get(); no == nil {
		ri.Logger = ri.Logger.With(zap.String("user_id", uid))
	} else {
		ri.Logger = ri.Logger.With(zap.String("user_id", "<ANONYMOUS>"))
	}
	return ri, nil
}
