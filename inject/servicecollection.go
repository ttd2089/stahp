package inject

import (
	"fmt"
	"maps"
	"reflect"
)

// A ServiceCollection is a collection into which services can be registered and from which a
// [ServiceProvider] may be built.
type ServiceCollection struct {
	registrations map[reflect.Type]serviceRegistration
}

type serviceRegistration struct {
	lifetime ServiceLifetime
	factory  func() (any, error)
}

// Build creates a [ServiceProvider] from the target [ServiceCollection]. A non-nil error is
// returned when the [ServiceCollection] is determined to be in a bad state at the time of the
// call, e.g. if a registered service has a dependency on a service type for which no
// implementation is registered, or if there are circular dependencies.
func (collection *ServiceCollection) Build() (ServiceProvider, error) {
	provider := ServiceProvider{
		registrations: make(map[reflect.Type]serviceRegistration),
	}
	maps.Copy(provider.registrations, collection.registrations)
	return provider, nil
}

// RegisterType registers the type of the given T as the concrete type to satisfy the service type
// T when instances are resolved from a [ServiceProvider] built from the given [ServiceCollection].
// After the instance is resolved, every exported field will be initialized by the same
// [ServiceProvider]. Note that the given instance of T is not used directly even for types
// registered with Singleton lifetime.
func RegisterType[T any](collection *ServiceCollection, lifetime ServiceLifetime, type_ T) {
	if lifetime != Transient {
		panic("unimplemented")
	}

	if collection.registrations == nil {
		collection.registrations = make(map[reflect.Type]serviceRegistration)
	}
	collection.registrations[TypeOf[T]()] = serviceRegistration{
		lifetime: lifetime,
		factory: func() (any, error) {
			var t T
			return t, nil
		},
	}
}

// RegisterFunc registers the given function as the factory to satisfy the service type T when
// instances are resolved from a [ServiceProvider] built from the given [ServiceCollection].
// Instances resolved by functions will not have their exported fields populated automatically so
// the function must handle all required service initialization. The function will receive the same
// [ServiceResolver] from which the instance is being resolved and may use it to initialize the
// service.
func RegisterFunc[T any](
	collection *ServiceCollection,
	lifetime ServiceLifetime,
	fn func(*ServiceResolver) T,
) {
	// NOTE: We can validate cyclic dependencies in RegisterType at build time but we can't find
	// them here without running the funcs. Consider executing them with a ServiceResolver impl
	// that tracks the resolutions to find cycles.
	panic("unimplemented")
}

// A ServiceProvider is a factory from which services can be resolved by type.
type ServiceProvider struct {
	registrations map[reflect.Type]serviceRegistration
}

var _ ServiceResolver = &ServiceProvider{}

// NewScope creates a new ServiceProvider which will create distinct instances when resolving any
// [Scoped] services.
func (s *ServiceProvider) NewScope() ServiceProvider {
	panic("unimplemented")
}

// Resolve provides an instance of the requested type if one is registered.
func (provider *ServiceProvider) Resolve(type_ reflect.Type) (any, error) {
	registration, ok := provider.registrations[type_]
	if !ok {
		return nil, fmt.Errorf("no implementation registered for service type %v", type_)
	}
	switch registration.lifetime {
	case Transient:
		return registration.factory()
	default:
		panic("unimplemented")
	}
}
