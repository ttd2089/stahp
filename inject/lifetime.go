package inject

// A ServiceLifetime describes the instantiation semantics for a service provided by a
// [ServiceProvider].
type ServiceLifetime int

const (
	// A Transient service has a new instance created every time the service is resolved.
	Transient ServiceLifetime = iota + 1

	// A Scoped service has a single instance created per [ServiceProvider]. If the same service is
	// resolved multiple times from the same [ServiceProvider] then the same instance will be
	// returned. If a new [ServiceProvider] scope is created and the same service is resolved from
	// the new [ServiceProvider] then a single new instance will be created and returned every time
	// the service is resolved from the new [ServiceProvider].
	//
	// NOTE: That because of the need to return the same instance multiple times, Scoped services
	// can only be registered for interfaces and pointer types.
	Scoped

	// A Singleton service has a single instance created and returned every time the service is
	// resolved from a top level [ServiceProvider] or any its descendant scopes.
	//
	// NOTE: That because of the need to return the same instance multiple times, Singleton
	// services can only be registered for interfaces and pointer types.
	Singleton
)

func (lifetime ServiceLifetime) String() string {
	switch lifetime {
	case Transient:
		return "Transient"
	case Scoped:
		return "Scoped"
	case Singleton:
		return "Singleton"
	}
	return "<unknown ServiceLifetime>"
}
