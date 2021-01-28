package grpc

import (
	"context"
	"github.com/peanut-cc/sugar/middleware"
	"github.com/peanut-cc/sugar/transport"
	"google.golang.org/grpc"
)

// ServerOption is gRPC server option.
type ServerOption func(o *Server)

// ServerEncodeErrorFunc is encode error func.
type ServerEncodeErrorFunc func(err error) error

// RecoveryHandlerFunc is recovery handler func.
type RecoveryHandlerFunc func(ctx context.Context, req, err interface{}) error

// ServerMiddleware with server middleware.
func ServerMiddleware(m ...middleware.Middleware) ServerOption {
	return func(o *Server) {
		o.globalMiddleware = middleware.Chain(m[0], m[1:]...)
	}
}

// ServerErrorEncoder with server error encoder.
func ServerErrorEncoder(d ServerEncodeErrorFunc) ServerOption {
	return func(o *Server) {
		o.errorEncoder = d
	}
}

// ServerRecoveryHandler with server recovery handler.
func ServerRecoveryHandler(h RecoveryHandlerFunc) ServerOption {
	return func(s *Server) {
		s.recoveryHandler = h
	}
}

// Server is a gRPC server wrapper.
type Server struct {
	globalMiddleware  middleware.Middleware
	serviceMiddleware map[interface{}]middleware.Middleware
	errorEncoder      ServerEncodeErrorFunc
	recoveryHandler   RecoveryHandlerFunc
}

// NewServer creates a gRPC server by options.
func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		errorEncoder:      DefaultErrorEncoder,
		recoveryHandler:   DefaultRecoveryHandler,
		serviceMiddleware: make(map[interface{}]middleware.Middleware),
	}
	for _, o := range opts {
		o(srv)
	}
	return srv
}

// Use use a middleware to the transport.
func (s *Server) Use(srv interface{}, m ...middleware.Middleware) {
	s.serviceMiddleware[srv] = middleware.Chain(m[0], m[1:]...)
}

// UnaryInterceptor returns a unary server interceptor.
func (s *Server) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (reply interface{}, err error) {
		defer func() {
			if rerr := recover(); rerr != nil {
				err = s.errorEncoder(s.recoveryHandler(ctx, req, rerr))
			}
		}()
		ctx = transport.NewContext(ctx, transport.Transport{Kind: "GRPC"})
		ctx = NewContext(ctx, ServerInfo{Server: info.Server, FullMethod: info.FullMethod})
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			return handler(ctx, req)
		}
		if m, ok := s.serviceMiddleware[info.Server]; ok {
			h = m(h)
		}
		if s.globalMiddleware != nil {
			h = s.globalMiddleware(h)
		}
		if reply, err = h(ctx, req); err != nil {
			return nil, s.errorEncoder(err)
		}
		return
	}
}
