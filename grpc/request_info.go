package grpc

import (
	"context"

	"github.com/markphelps/optional"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type CommonRequestInfo struct {
	RequestID   string
	Metadata    metadata.MD
	RequestName string
	Span        opentracing.Span
	Context     context.Context
	Logger      *zap.Logger
}

type PublicRequestInfo struct {
	CommonRequestInfo
	TokenHash string
	UserID    optional.String
}

func (ri *CommonRequestInfo) GetChildSpan(name string) opentracing.Span {
	return opentracing.StartSpan(name, opentracing.ChildOf(ri.Span.Context()))
}

func (ri *CommonRequestInfo) WithSpan(span opentracing.Span) *CommonRequestInfo {
	return &CommonRequestInfo{
		RequestID:   ri.RequestID,
		Metadata:    ri.Metadata,
		RequestName: ri.RequestName,
		Span:        span,
		Context:     ri.Context,
		Logger:      ri.Logger,
	}
}

func (ri *PublicRequestInfo) WithSpan(span opentracing.Span) *PublicRequestInfo {
	return &PublicRequestInfo{
		CommonRequestInfo: *ri.CommonRequestInfo.WithSpan(span),
		TokenHash:         ri.TokenHash,
		UserID:            ri.UserID,
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

func GetCommonRequestInfo(ctx context.Context, requestName string, log *zap.Logger) (*CommonRequestInfo, error) {
	var err error

	md, found := metadata.FromIncomingContext(ctx)
	if !found {
		return nil, ErrInvalidRequest
	}
	zapReq := zap.String("method", requestName)
	ri := &CommonRequestInfo{}
	ri.Metadata = md
	if ri.RequestID, err = getFromMetadata(md, "x-request-id"); err != nil {
		log.Error("no x-request-id found", zapReq)
		return nil, ErrInvalidRequest
	}
	ri.Logger = log.With(zapReq, zap.String("request_id", ri.RequestID))
	ri.RequestName = requestName
	ri.Span, ri.Context = opentracing.StartSpanFromContext(ctx, requestName)
	return ri, nil
}

func GetPublicRequestInfo(ctx context.Context, requestName string, log *zap.Logger) (*PublicRequestInfo, error) {
	ci, err := GetCommonRequestInfo(ctx, requestName, log)
	if err != nil {
		return nil, err
	}
	ri := &PublicRequestInfo{CommonRequestInfo: *ci}

	if ri.TokenHash, err = getFromMetadata(ci.Metadata, "x-auth-ticket"); err != nil {
		ri.Logger.Error("no x-auth-ticket found")
		return nil, ErrUnauthenticated
	}
	if v, no := getFromMetadata(ci.Metadata, "x-user-id"); no == nil {
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
