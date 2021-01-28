package http

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/peanut-cc/sugar/middleware"
	"github.com/peanut-cc/sugar/transport"
	"net/http"
)

// SupportPackageIsVersion1 These constants should not be referenced from any other code.
const SupportPackageIsVersion1 = true

// ServiceRegistrar wraps a single method that supports service registration.
type ServiceRegistrar interface {
	RegisterService(desc *ServiceDesc, impl interface{})
}

// ServiceDesc represents a HTTP service's specification.
type ServiceDesc struct {
	ServiceName string
	HandlerType interface{}
	Methods     []MethodDesc
	Metadata    interface{}
}

type serverMethodHandler func(srv interface{}, ctx context.Context, req *http.Request) (interface{}, error)

// MethodDesc represents a HTTP service's method specification.
type MethodDesc struct {
	Path    string
	Method  string
	Handler serverMethodHandler
}


// ServerOption is HTTP server option.
type ServerOption func(*Server)

// ServerDecodeRequestFunc is decode request func.
type ServerDecodeRequestFunc func(in interface{}, req *http.Request) error

// ServerEncodeResponseFunc is encode response func.
type ServerEncodeResponseFunc func(out interface{}, res http.ResponseWriter, req *http.Request) error

// ServerEncodeErrorFunc is encode error func.
type ServerEncodeErrorFunc func(err error, res http.ResponseWriter, req *http.Request)

// RecoveryHandlerFunc is recovery handler func.
type RecoveryHandlerFunc func(ctx context.Context, req, err interface{}) error

// ServerRequestDecoder with decode request option.
func ServerRequestDecoder(fn ServerEncodeErrorFunc) ServerOption {
	return func(s *Server) {
		s.errorEncoder = fn
	}
}

// ServerResponseEncoder with response handler option.
func ServerResponseEncoder(fn ServerEncodeResponseFunc) ServerOption {
	return func(s *Server) {
		s.responseEncoder = fn
	}
}

// ServerErrorEncoder with error handler option.
func ServerErrorEncoder(fn ServerEncodeErrorFunc) ServerOption {
	return func(s *Server) {
		s.errorEncoder = fn
	}
}

// ServerRecoveryHandler with recovery handler.
func ServerRecoveryHandler(fn RecoveryHandlerFunc) ServerOption {
	return func(s *Server) {
		s.recoveryHandler = fn
	}
}

// ServerMiddleware with server middleware option.
func ServerMiddleware(m ...middleware.Middleware) ServerOption {
	return func(s *Server) {
		s.globalMiddleware = middleware.Chain(m[0], m[1:]...)
	}
}

// Server is a HTTP server wrapper.
type Server struct {
	router            *mux.Router
	requestDecoder    ServerDecodeRequestFunc
	responseEncoder   ServerEncodeResponseFunc
	errorEncoder      ServerEncodeErrorFunc
	recoveryHandler   RecoveryHandlerFunc
	globalMiddleware  middleware.Middleware
	serviceMiddleware map[interface{}]middleware.Middleware
}

// NewServer creates a HTTP server by options.
func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		router:            mux.NewRouter(),
		requestDecoder:    DefaultRequestDecoder,
		responseEncoder:   DefaultResponseEncoder,
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

// Handle registers a new route with a matcher for the URL path.
func (s *Server) Handle(path string, handler http.Handler) {
	s.router.Handle(path, handler)
}

// HandleFunc registers a new route with a matcher for the URL path.
func (s *Server) HandleFunc(path string, h func(http.ResponseWriter, *http.Request)) {
	s.router.HandleFunc(path, h)
}

// ServeHTTP should write reply headers and data to the ResponseWriter and then return.
func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ctx := transport.NewContext(req.Context(), transport.Transport{Kind: "HTTP"})
	ctx = NewContext(ctx, ServerInfo{Request: req, Response: res})
	s.router.ServeHTTP(res, req.WithContext(ctx))
}

// RegisterService registers a service and its implementation to the HTTP server.
func (s *Server) RegisterService(sd *ServiceDesc, ss interface{}) {
	for _, method := range sd.Methods {
		s.registerHandle(ss, method)
	}
}

func (s *Server) registerHandle(srv interface{}, md MethodDesc) {
	s.router.HandleFunc(md.Path, func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if rerr := recover(); rerr != nil {
				err := s.recoveryHandler(req.Context(), req.Form, rerr)
				s.errorEncoder(err, res, req)
			}
		}()

		handler := func(ctx context.Context, in interface{}) (interface{}, error) {
			return md.Handler(srv, ctx, req)
		}
		if m, ok := s.serviceMiddleware[srv]; ok {
			handler = m(handler)
		}
		if s.globalMiddleware != nil {
			handler = s.globalMiddleware(handler)
		}

		reply, err := handler(req.Context(), req)
		if err != nil {
			s.errorEncoder(err, res, req)
			return
		}

		if err := s.responseEncoder(reply, res, req); err != nil {
			s.errorEncoder(err, res, req)
			return
		}

	}).Methods(md.Method)
}
