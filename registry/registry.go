package registry


import "context"

// Registrar is service registrar interface.
type Registrar interface {
	// Register the registration.
	Register(ctx context.Context, svc *Service) error
	// Deregister the registration.
	Deregister(ctx context.Context, svc *Service) error
}

// Discovery is service discovery interface.
type Discovery interface {
	// GetService return the service instances in memory according to the service name.
	GetService(name string) ([]*Service, error)
	// GetService return all service instances in memory which aggregated according to service name.
	ListService() (map[string][]*Service, error)
	// Resolve creates a watcher according to the service name.
	Resolve(name string) (Watcher, error)
}

// Watcher is service watcher.
type Watcher interface {
	// Watch returns services in the following two cases:
	// 1.the first time to watch and the service instance list is not empty.
	// 2.any service instance changes found.
	// if the above two conditions are not met, it will block until context deadline exceeded or canceled
	Watch(ctx context.Context) ([]*Service, error)
	// Close close the watcher.
	Close()
}

// Service is an instance of a service in a discovery system.
type Service struct {
	// ID is the unique instance ID as registered.
	ID string
	// Name is the service name as registered.
	Name string
	// Version is the version of the compiled.
	Version string
	// Metadata is the kv pair metadata associated with the service instance.
	Metadata map[string]string
	// Endpoints is endpoint addresses of the service instance.
	// schema:
	//   http://127.0.0.1:8000?isSecure=false
	//   grpc://127.0.0.1:9000?isSecure=false
	Endpoints []string
}
