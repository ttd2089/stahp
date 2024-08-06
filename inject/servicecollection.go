package inject

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"sync"
)

// ErrNonTransientStruct is returned when a struct type is registered with a [ServiceLifetime]
// other than [Transient]. Struct types may only be registered as Transient because an instance of
// a struct can not be shared. Register a pointer to a struct to use with a [ServiceLifetime] other
// than [Transient].
var ErrNonTransientStruct = errors.New("struct types may only be registered as Transient")

// ErrInvalidImplementation is returned when an implementation is registered for a service type
// it cannot be assigned to.
var ErrInvalidImplementation = errors.New("implementation type must be assignable to service type")

// A ServiceCollection is a collection into which services can be registered and from which a
// [ServiceProvider] may be built.
type ServiceCollection struct {
	registrations map[reflect.Type]serviceRegistration
}

// Build creates a [ServiceProvider] from the target [ServiceCollection]. A non-nil error is
// returned when the [ServiceCollection] is determined to be in a bad state at the time of the
// call, e.g. if a registered service has a dependency on a service type for which no
// implementation is registered, or if there are circular dependencies.
func (services *ServiceCollection) Build() (ServiceProvider, error) {
	if services == nil {
		return ServiceProvider{}, errors.New("cannot build ServiceProvider from nil ServiceCollection")
	}
	// TODO: analyze graph for validity
	registrations := make(map[reflect.Type]serviceRegistration, len(services.registrations))
	maps.Copy(registrations, services.registrations)
	return ServiceProvider{
		registrations:   registrations,
		scopedInstances: make(map[reflect.Type]any, len(services.registrations)),
	}, nil
}

func (services *ServiceCollection) addRegistration(serviceType reflect.Type, registration serviceRegistration) {
	if services.registrations == nil {
		services.registrations = make(map[reflect.Type]serviceRegistration)
	}
	services.registrations[serviceType] = registration
}

// A ServiceProvider is a factory from which services can be resolved by type.
type ServiceProvider struct {
	registrations   map[reflect.Type]serviceRegistration
	mu              sync.Mutex
	scopedInstances map[reflect.Type]any
}

// NewScope creates a new ServiceProvider which will create distinct instances when resolving any
// [Scoped] services.
func (provider *ServiceProvider) NewScope() ServiceProvider {
	panic("unimplemented")
}

// Resolve provides an instance of the requested type if one is registered.
func (provider *ServiceProvider) Resolve(type_ reflect.Type) (any, error) {
	if provider == nil {
		return nil, errors.New("cannot resolve instances from nil ServiceProvider")
	}
	registration, ok := provider.registrations[type_]
	if !ok {
		return nil, fmt.Errorf("no implementation registered for service type %v", type_)
	}
	switch registration.lifetime {
	case Transient:
		return registration.factory(provider)
	case Scoped:
		return provider.resolveScoped(type_, registration.factory)
	case Singleton:
		return registration.factory(provider)
	default:
		panic("this code should be unreachable: please open a an issue at https://github.com/ttd2089/stahp/issues/new")
	}
}

func (provider *ServiceProvider) resolveScoped(type_ reflect.Type, factory factoryFunc) (any, error) {
	// No need to lock if we've already saved the scoped instance.
	if service, ok := provider.scopedInstances[type_]; ok {
		return service, nil
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	// Someone may have saved a scoped instance while we were waiting for a lock so check again.
	if service, ok := provider.scopedInstances[type_]; ok {
		return service, nil
	}
	// Build, save, and return the scoped instance.
	service, err := factory(provider)
	if err != nil {
		return nil, err
	}
	// We would have initialized this but since we can't stop someone from creating a default
	// instance we need to avoid writes to nil maps.
	if provider.scopedInstances == nil {
		provider.scopedInstances = make(map[reflect.Type]any, len(provider.registrations))
	}
	provider.scopedInstances[type_] = service
	return service, nil
}

type factoryFunc func(ServiceResolver) (any, error)

type serviceRegistration struct {
	lifetime ServiceLifetime
	factory  factoryFunc
}

// RegisterType registers the type of the given T as the concrete type to satisfy the service type
// T when instances are resolved from a [ServiceProvider] built from the given [ServiceCollection].
// After the instance is resolved, every exported field will be initialized by the same
// [ServiceProvider]. Note that the given instance of T is not used directly even for types
// registered with Singleton lifetime.
func RegisterType[T any](services *ServiceCollection, lifetime ServiceLifetime, type_ T) error {
	if services == nil {
		return errors.New("cannot register types to a nil ServiceProvider")
	}

	implType := reflect.TypeOf(type_)

	if lifetime != Transient && implType.Kind() == reflect.Struct {
		return ErrNonTransientStruct
	}

	factory, err := getDefaultFactory(implType)
	if err != nil {
		return err
	}

	services.addRegistration(reflect.TypeFor[T](), serviceRegistration{
		lifetime: lifetime,
		factory:  factory,
	})

	return nil
}

func getDefaultFactory(type_ reflect.Type) (factoryFunc, error) {
	// How we initialize the impl depends on the kind.
	if type_.Kind() == reflect.Struct {
		return func(ServiceResolver) (any, error) {
			return reflect.Zero(type_).Interface(), nil
		}, nil
	}
	if type_.Kind() == reflect.Pointer && type_.Elem().Kind() == reflect.Struct {
		elemType := type_.Elem()
		return func(ServiceResolver) (any, error) {
			return reflect.New(elemType).Interface(), nil
		}, nil
	}
	panic("unimplemented")
}

func RegisterFunc[Service any, Impl any](
	services *ServiceCollection,
	lifetime ServiceLifetime,
	factory func(ServiceResolver) (Impl, error),
) error {
	if services == nil {
		return errors.New("cannot register types to a nil ServiceProvider")
	}

	serviceType := reflect.TypeFor[Service]()
	implType := reflect.TypeFor[Impl]()

	if !implType.AssignableTo(serviceType) {
		return ErrInvalidImplementation
	}

	if lifetime != Transient && implType.Kind() == reflect.Struct {
		return ErrNonTransientStruct
	}

	services.addRegistration(reflect.TypeFor[Service](), serviceRegistration{
		lifetime: lifetime,
		factory: func(resolver ServiceResolver) (any, error) {
			return factory(resolver)
		},
	})

	return nil
}
